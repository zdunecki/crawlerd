package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/orlangure/gnomock"
	kafkapreset "github.com/orlangure/gnomock/preset/kafka"
	mongopreset "github.com/orlangure/gnomock/preset/mongo"
	"github.com/zdunecki/crawlerd/api/v1/sdk"
	"github.com/zdunecki/crawlerd/pkg/pubsub"
	"github.com/zdunecki/crawlerd/pkg/scheduler"
	storageopt "github.com/zdunecki/crawlerd/pkg/store/options"
	"github.com/zdunecki/crawlerd/pkg/worker"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func testK8sWorker(mongoDBName, mongoURI, schedulerGRPCAddr, etcdAddr, kafkaBroker string, k8s kubernetes.Interface, k8sNamespace, k8sWorkerSelector string) {
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
				}, 0).Registry(), // TODO: registryTTLBuffer > 0
		),
		worker.WithSchedulerGRPCAddr(schedulerGRPCAddr),
		worker.WithK8sCluster(k8s, k8sNamespace, k8sWorkerSelector),
		worker.WithPubSub(kafka),
	)

	if err != nil {
		panic(err)
	}

	if err := work.Serve(context.Background()); err != nil {
		panic(err)
	}
}

func testK8sScheduler(grpcAddr, mongoDBName, mongoURI, etcdAddr string, k8s kubernetes.Interface, k8sNamespace, workerSelector string) {
	schedule, err := scheduler.New(
		scheduler.WithWorkerClusterConfig(worker.InitConfig()),
		scheduler.WithStorage(
			storageopt.Client().
				WithMongoDB(mongoDBName, options.Client().ApplyURI(mongoURI)).URL().History(),
			storageopt.Client().
				WithETCD(clientv3.Config{
					Endpoints:   []string{etcdAddr},
					DialTimeout: time.Second * 15,
				}, 0).Registry(), // TODO: registry ttl buffer
		),
		scheduler.WithWatcher(scheduler.NewWatcherOption().
			WithK8s(k8s, k8sNamespace, workerSelector),
		),
	)
	if err != nil {
		panic(err)
	}

	if err := schedule.Serve(grpcAddr); err != nil {
		panic(err)
	}
}

// TODO: find better solution than sleep for waiting on services healtcheck
func setupK8sClient(k8sNamespace, k8sWorkerSelector string, k8sObjects ...runtime.Object) (*setup, error) {
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
	k8s := fake.NewSimpleClientset(k8sObjects...)

	go func() {
		testK8sScheduler(schedulerGRPCAddr, dbName, mongoURI, etcdAddr, k8s, k8sNamespace, k8sWorkerSelector)
	}()

	go func() {
		testK8sWorker(dbName, mongoURI, schedulerGRPCAddr, etcdAddr, kafkaBroker, k8s, k8sNamespace, k8sWorkerSelector)
	}()

	time.Sleep(time.Second * 2)

	go func() {
		testApi(apiHost, schedulerGRPCAddr, dbName, mongoURI)
	}()

	c, err := sdk.NewWithOpts(sdk.WithHTTPAddr("http://localhost:6666"))

	return &setup{
		etcdContainer: etcdContainer,
		crawld:        c,
		done:          done,
	}, err
}
