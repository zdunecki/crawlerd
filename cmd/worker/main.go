package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"crawlerd/pkg/pubsub"
	storageopt "crawlerd/pkg/storage/options"
	"crawlerd/pkg/worker"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var (
		port string

		dbName string

		mongo     bool
		mongoHost string
		mongoPort string

		etcd     bool
		etcdHost string // TODO: deprecated
		etcdAddr string

		schedulerAddr string

		kafkaBroker string

		k8sNamespace      string
		k8sWorkerSelector string
	)

	flag.StringVar(&port, "port", "", "grpc port listening on")

	flag.StringVar(&dbName, "db", "crawlerd", "database name")

	flag.BoolVar(&mongo, "mongo", true, "use mongodb as a database source")
	flag.StringVar(&mongoHost, "mongo-host", "", "mongo host")
	flag.StringVar(&mongoPort, "mongo-port", "27017", "mongo port")

	flag.BoolVar(&etcd, "etcd", true, "use etcd as a registry source")
	flag.StringVar(&etcdHost, "etcd-host", "", "etcd host")
	flag.StringVar(&etcdAddr, "etcd-addr", "", "etcd address")

	flag.StringVar(&schedulerAddr, "scheduler-addr", "", "scheduler address")

	flag.StringVar(&kafkaBroker, "kafka-broker", "", "kafka broker")

	flag.StringVar(&k8sNamespace, "k8s-namespace", "", "k8s namespace")
	flag.StringVar(&k8sWorkerSelector, "k8s-worker-selector", "", "selector match for worker")

	flag.Parse()

	var opts []worker.Option

	var etcdEndpoints []string

	if etcdAddr != "" {
		etcdEndpoints = append(etcdEndpoints, etcdAddr)
	} else {
		etcdEndpoints = append(etcdEndpoints, etcdHost+":2379")
	}

	cfg := worker.InitConfig()

	etcdConfig := clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: time.Second * 15,
	}

	if port != "" {
		cfg.WorkerGRPCAddr = port
	}

	if schedulerAddr != "" {
		cfg.SchedulerGRPCAddr = schedulerAddr
	}

	if k8sNamespace != "" && k8sWorkerSelector != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		opts = append(opts, worker.WithK8sCluster(clientset, k8sNamespace, k8sWorkerSelector))
	} else if etcd {
		if etcdHost == "" {
			opts = append(opts, worker.WithETCDCluster())
		} else {
			opts = append(opts, worker.WithETCDCluster(
				etcdConfig,
			))
		}
	}

	var registryTTLBuffer int64 = 60

	if mongo {
		if mongoHost == "" {
			opts = append(opts, worker.WithStorage(
				storageopt.Client().
					WithMongoDB(dbName, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:27017", "localhost"))).URL().History(),
				storageopt.Client().
					WithETCD(etcdConfig, registryTTLBuffer).Registry(),
			))
		} else {
			opts = append(opts, worker.WithStorage(
				storageopt.Client().
					WithMongoDB(dbName, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort))).URL().History(),
				storageopt.Client().
					WithETCD(etcdConfig, registryTTLBuffer).Registry(),
			))
		}
	}

	if kafkaBroker != "" {
		pubsub, err := pubsub.NewKafka(kafkaBroker)
		if err != nil {
			log.Error(err)
			panic(err)
		}

		opts = append(opts, worker.WithPubSub(pubsub))
	}

	work, err := worker.New(cfg, opts...)

	if err != nil {
		log.Error(err)
		panic(err)
	}

	if err := work.Serve(context.Background()); err != nil {
		log.Error(err)
		panic(err)
	}
}
