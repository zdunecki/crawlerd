package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/orlangure/gnomock"
	kafkapreset "github.com/orlangure/gnomock/preset/kafka"
	mongopreset "github.com/orlangure/gnomock/preset/mongo"
	"github.com/zdunecki/crawlerd/api"
	v1 "github.com/zdunecki/crawlerd/api/v1"
	"github.com/zdunecki/crawlerd/api/v1/sdk"
	"github.com/zdunecki/crawlerd/pkg/pubsub"
	"github.com/zdunecki/crawlerd/pkg/scheduler"
	"github.com/zdunecki/crawlerd/pkg/store"
	storageopt "github.com/zdunecki/crawlerd/pkg/store/options"
	"github.com/zdunecki/crawlerd/pkg/worker"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type setupOptions struct {
	withCacheRegistry     bool
	registryTTLBuffer     int64
	additionalWorkerCount int
}

type ETCDPreset struct {
	Version string `json:"version"`
}

func (p *ETCDPreset) Image() string {
	return fmt.Sprintf("docker.io/bitnami/etcd:%s", p.Version)
}

func (p *ETCDPreset) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(2379)
}

func (p *ETCDPreset) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithEnv("ALLOW_NONE_AUTHENTICATION=yes"),
	}

	return opts
}

func (p *ETCDPreset) setDefaults() {
	if p.Version == "" {
		p.Version = "3"
	}
}

func NewETCDPreset() gnomock.Preset {
	return &ETCDPreset{}
}

func testWorker(e2e *worker_e2e, mongoDBName, mongoURI, schedulerGRPCAddr, etcdAddr, kafkaBroker string, opts *setupOptions) worker.Worker {
	kafka, err := pubsub.NewKafka(kafkaBroker)
	if err != nil {
		panic(err)
	}

	cfg := worker.InitConfig()

	work, err := worker.New(
		cfg,
		worker.WithStorage(
			storageopt.Client().
				WithMongoDB(mongoDBName, options.Client().ApplyURI(mongoURI)).URL().History(),
			storageopt.Client().
				WithETCD(clientv3.Config{
					Endpoints:   []string{etcdAddr},
					DialTimeout: time.Second * 15,
				}, opts.registryTTLBuffer).Registry(),
		),
		worker.WithSchedulerGRPCAddr(schedulerGRPCAddr),
		worker.WithETCDCluster(
			clientv3.Config{
				Endpoints:   []string{etcdAddr},
				DialTimeout: time.Second * 15,
			},
		),
		worker.WithPubSub(kafka),
		worker.WithBrotliCompression(),
	)

	work.ID()
	if err != nil {
		panic(err)
	}

	go func() {
		if err := work.Serve(e2e.ctx); err != nil {
			panic(err)
		}
		e2e.doneC <- true
	}()

	return work
}

func testScheduler(grpcAddr, mongoDBName, mongoURI, etcdAddr string) {
	schedule, err := scheduler.New(
		scheduler.WithStorage(
			storageopt.Client().
				WithMongoDB(mongoDBName, options.Client().ApplyURI(mongoURI)).URL().History(),
			storageopt.Client().
				WithETCD(clientv3.Config{
					Endpoints:   []string{etcdAddr},
					DialTimeout: time.Second * 15,
				}, 0).Registry(), // TODO: registry ttl buffer
		),
		scheduler.WithWatcher(scheduler.NewWatcherOption().WithETCD(
			clientv3.Config{
				Endpoints:   []string{etcdAddr}, // get host from container
				DialTimeout: time.Second * 15,
			},
		)),
	)
	if err != nil {
		panic(err)
	}

	if err := schedule.Serve(grpcAddr); err != nil {
		panic(err)
	}
}

func testApi(appAddr, schedulerGRPCAddr, mongoDBName, mongoURI string) store.Repository {
	apiV1, err := v1.New(
		v1.WithMongoDBStorage(mongoDBName, options.Client().ApplyURI(mongoURI)),
		v1.WithGRPCSchedulerServer(schedulerGRPCAddr),
	)

	if err != nil {
		panic(err)
	}

	go func() {
		if err := apiV1.Serve(appAddr, api.New(chi.NewMux())); err != nil {
			panic(err)
		}
	}()

	return apiV1.Store()
}

func randomPort() string {
	min := 9890
	max := 9900
	return fmt.Sprintf("%d", rand.Intn(max-min)+min)
}

type worker_e2e struct {
	ctx       context.Context
	ctxCancel func()

	doneC chan bool
}

func (w *worker_e2e) done() <-chan bool {
	return w.doneC
}

type setup struct {
	crawld        v1.V1
	cluster       worker.Cluster
	store         store.Repository
	etcdContainer *gnomock.Container
	done          func()
	worker_e2e    []*worker_e2e
}

func setupClient(opts *setupOptions) (*setup, error) {
	containers := make([]*gnomock.Container, 0)

	done := func() {
		gnomock.Stop(containers...)
	}

	schedulerGRPCAddr := ":" + randomPort()
	apiHost := ":6666"
	dbName := "test"

	mgoPreset := mongopreset.Preset()
	mongoContainer, err := gnomock.Start(mgoPreset)
	if err != nil {
		done()
		return nil, err
	}
	kafkaPreset := kafkapreset.Preset()
	kafkaContainer, err := gnomock.Start(kafkaPreset)
	if err != nil {
		done()
		return nil, err
	}

	etcdPreset := NewETCDPreset()
	etcdContainer, err := gnomock.Start(etcdPreset)
	if err != nil {
		done()
		return nil, err
	}

	containers = append(containers, mongoContainer, kafkaContainer, etcdContainer)

	kafkaBroker := kafkaContainer.Address("broker")
	mongoURI := fmt.Sprintf("mongodb://%s", mongoContainer.DefaultAddress())
	etcdAddr := etcdContainer.DefaultAddress()

	testAPIStore := testApi(apiHost, schedulerGRPCAddr, dbName, mongoURI)

	go func() {
		testScheduler(schedulerGRPCAddr, dbName, mongoURI, etcdAddr)
	}()

	workere2e := make([]*worker_e2e, 0)

	var cluster worker.Cluster

	go func() {
		workerCtx, workerCtxCancel := context.WithCancel(context.Background())
		e2e := &worker_e2e{
			ctx:       workerCtx,
			ctxCancel: workerCtxCancel,
			doneC:     make(chan bool),
		}
		workere2e = append(workere2e, e2e)

		w := testWorker(e2e, dbName, mongoURI, schedulerGRPCAddr, etcdAddr, kafkaBroker, opts)
		cluster = w.Cluster()
	}()

	if opts.additionalWorkerCount > 0 {
		for i := 0; i < opts.additionalWorkerCount; i++ {
			time.Sleep(time.Second * 1)

			go func() {
				workerCtx, workerCtxCancel := context.WithCancel(context.Background())
				e2e := &worker_e2e{
					ctx:       workerCtx,
					ctxCancel: workerCtxCancel,
					doneC:     make(chan bool),
				}
				workere2e = append(workere2e, e2e)

				testWorker(e2e, dbName, mongoURI, schedulerGRPCAddr, etcdAddr, kafkaBroker, opts)
			}()
		}
	}

	time.Sleep(time.Second * 1)

	c, err := sdk.NewWithOpts(sdk.WithHTTPAddr("http://localhost:6666"))

	return &setup{
		crawld:        c,
		cluster:       cluster,
		store:         testAPIStore,
		etcdContainer: etcdContainer,
		done:          done,
		worker_e2e:    workere2e,
	}, err
}
