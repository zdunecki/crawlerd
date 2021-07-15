package e2e

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"crawlerd/api"
	v1 "crawlerd/api/v1"
	"crawlerd/api/v1/client"
	"crawlerd/pkg/pubsub"
	"crawlerd/pkg/scheduler"
	storageopt "crawlerd/pkg/storage/options"
	"crawlerd/pkg/worker"
	"github.com/go-chi/chi/v5"
	"github.com/orlangure/gnomock"
	kafkapreset "github.com/orlangure/gnomock/preset/kafka"
	mongopreset "github.com/orlangure/gnomock/preset/mongo"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type setupOptions struct {
	withCacheRegistry bool
	registryTTLBuffer int64
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

func testWorker(mongoDBName, mongoURI, schedulerGRPCAddr, etcdAddr, kafkaBroker string, opts *setupOptions) {
	kafka, err := pubsub.NewKafka(kafkaBroker)
	if err != nil {
		panic(err)
	}

	work, err := worker.New(
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

	if err != nil {
		panic(err)
	}

	if err := work.Serve(context.Background()); err != nil {
		panic(err)
	}
}

func testScheduler(grpcAddr, mongoDBName, mongoURI, etcdAddr string) {
	schedule, err := scheduler.New(
		scheduler.WithMongoDBStorage(mongoDBName, options.Client().ApplyURI(mongoURI)),
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

func testApi(appAddr, schedulerGRPCAddr, mongoDBName, mongoURI string) {
	apiV1, err := v1.New(
		v1.WithMongoDBStorage(mongoDBName, options.Client().ApplyURI(mongoURI)),
		v1.WithGRPCSchedulerServer(schedulerGRPCAddr),
	)

	if err != nil {
		panic(err)
	}

	if err := apiV1.Serve(appAddr, api.New(chi.NewMux())); err != nil {
		panic(err)
	}
}

func randomPort() string {
	min := 9890
	max := 9900
	return fmt.Sprintf("%d", rand.Intn(max-min)+min)
}

type setup struct {
	etcdContainer *gnomock.Container
	crawld        v1.V1
	done          func()
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

	go func() {
		testApi(apiHost, schedulerGRPCAddr, dbName, mongoURI)
	}()

	go func() {
		testScheduler(schedulerGRPCAddr, dbName, mongoURI, etcdAddr)
	}()

	go func() {
		testWorker(dbName, mongoURI, schedulerGRPCAddr, etcdAddr, kafkaBroker, opts)
	}()

	time.Sleep(time.Second * 1)

	c, err := client.NewWithOpts(client.WithHTTP("http://localhost:6666"))

	return &setup{
		etcdContainer: etcdContainer,
		crawld:        c,
		done:          done,
	}, err
}
