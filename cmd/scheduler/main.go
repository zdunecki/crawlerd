package main

import (
	"flag"
	"fmt"
	"time"

	"crawlerd/pkg/scheduler"
	storageopt "crawlerd/pkg/store/options"
	"crawlerd/pkg/worker"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var (
		workerPort string

		dbName string

		mongo     bool
		mongoHost string
		mongoPort string

		etcdRegistry bool
		etcdHost     string // TODO: deprecated
		etcdAddr     string

		registryTTL int64

		grpcAddr string

		k8sNamespace      string
		k8sWorkerSelector string
	)

	flag.StringVar(&workerPort, "worker-port", "", "grpc worker port is listening on")

	flag.StringVar(&dbName, "db", "crawlerd", "database name")

	flag.BoolVar(&mongo, "mongo", true, "use mongodb as a database source")
	flag.StringVar(&mongoHost, "mongo-host", "", "mongo host")
	flag.StringVar(&mongoPort, "mongo-port", "27017", "mongo port")

	flag.BoolVar(&etcdRegistry, "etcd-registry", false, "use etcd as a registry source")
	flag.StringVar(&etcdHost, "etcd-host", "", "etcd host")
	flag.StringVar(&etcdAddr, "etcd-addr", "", "etcd address")

	flag.Int64Var(&registryTTL, "registry-ttl", 0, "registry ttl")

	flag.StringVar(&grpcAddr, "grpc-addr", scheduler.DefaultSchedulerGRPCServerAddr, "grpc addr")

	flag.StringVar(&k8sNamespace, "k8s-namespace", "", "k8s namespace")
	flag.StringVar(&k8sWorkerSelector, "k8s-worker-selector", "", "selector match for worker")

	flag.Parse()

	var opts []scheduler.Option

	var etcdEndpoints []string

	fmt.Println("update 3")

	if etcdAddr != "" {
		etcdEndpoints = append(etcdEndpoints, etcdAddr)
	} else {
		etcdEndpoints = append(etcdEndpoints, etcdHost+":2379")
	}

	// TODO: redis, cache registry

	etcdConfig := clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: time.Second * 15,
	}

	if mongo {
		if mongoHost == "" {
			storageOpts := []*storageopt.RepositoryOption{
				storageopt.Client().
					WithMongoDB(dbName, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:27017", "localhost"))).URL().History().Registry().RequestQueue(),
			}

			if etcdRegistry {
				storageOpts = append(storageOpts, storageopt.Client().WithETCD(etcdConfig, registryTTL).Registry())
			}

			opts = append(opts, scheduler.WithStorage(storageOpts...))
		} else {
			storageOpts := []*storageopt.RepositoryOption{
				storageopt.Client().
					WithMongoDB(dbName, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort))).URL().History().Registry().RequestQueue(),
			}

			if etcdRegistry {
				storageOpts = append(storageOpts, storageopt.Client().WithETCD(etcdConfig, registryTTL).Registry())
			}

			opts = append(opts, scheduler.WithStorage(storageOpts...))
		}
	}

	if k8sNamespace != "" && k8sWorkerSelector != "" {
		if workerPort == "" {
			panic("worker porter is required")
		}

		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		opts = append(opts,
			scheduler.WithWorkerClusterConfig(&worker.Config{WorkerGRPCAddr: workerPort}),
			scheduler.WithWatcher(
				scheduler.NewWatcherOption().WithK8s(clientset, k8sNamespace, k8sWorkerSelector),
			),
		)
	} else {
		if etcdHost == "" {
			opts = append(opts, scheduler.WithWatcher())
		} else {
			opts = append(opts, scheduler.WithWatcher(
				scheduler.NewWatcherOption().WithETCD(etcdConfig),
			))
		}
	}

	schedule, err := scheduler.New(opts...)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	if err := schedule.Serve(grpcAddr); err != nil {
		log.Error(err)
		panic(err)
	}
}
