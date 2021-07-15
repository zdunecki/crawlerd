package e2e

import (
	"context"
	"fmt"
	"time"

	"crawlerd/api/v1/client"
	"crawlerd/pkg/pubsub"
	"crawlerd/pkg/scheduler"
	storageopt "crawlerd/pkg/storage/options"
	"crawlerd/pkg/worker"
	"github.com/orlangure/gnomock"
	kafkapreset "github.com/orlangure/gnomock/preset/kafka"
	mongopreset "github.com/orlangure/gnomock/preset/mongo"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func testK8sWorker(mongoDBName, mongoURI, schedulerGRPCAddr, etcdAddr, kafkaBroker string, k8s kubernetes.Interface, k8sNamespace string) {
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
				}, 0).Registry(), // TODO: registryTTLBuffer > 0
		),
		worker.WithSchedulerGRPCAddr(schedulerGRPCAddr),
		worker.WithK8sCluster(k8s, k8sNamespace),
		worker.WithPubSub(kafka),
	)

	if err != nil {
		panic(err)
	}

	if err := work.Serve(context.Background()); err != nil {
		panic(err)
	}
}

func testK8sScheduler(grpcAddr, mongoDBName, mongoURI string, k8s kubernetes.Interface, k8sNamespace string) {
	schedule, err := scheduler.New(
		scheduler.WithWorkerClusterConfig(worker.InitConfig()),
		scheduler.WithMongoDBStorage(mongoDBName, options.Client().ApplyURI(mongoURI)),
		scheduler.WithWatcher(scheduler.NewWatcherOption().
			WithK8s(k8s, k8sNamespace),
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
func setupK8sClient(k8sNamespace string, k8sObjects ...runtime.Object) (*setup, error) {
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
		testK8sScheduler(schedulerGRPCAddr, dbName, mongoURI, k8s, k8sNamespace)
	}()

	go func() {
		testK8sWorker(dbName, mongoURI, schedulerGRPCAddr, etcdAddr, kafkaBroker, k8s, k8sNamespace)
	}()

	time.Sleep(time.Second * 2)

	go func() {
		testApi(apiHost, schedulerGRPCAddr, dbName, mongoURI)
	}()

	c, err := client.NewWithOpts(client.WithHTTP("http://localhost:6666"))

	return &setup{
		etcdContainer: etcdContainer,
		crawld:        c,
		done:          done,
	}, err
}
