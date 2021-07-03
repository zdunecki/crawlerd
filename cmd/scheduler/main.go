package main

import (
	"flag"
	"fmt"
	"time"

	"crawlerd/pkg/scheduler"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		dbName    string
		mongo     bool
		mongoHost string
		etcd      bool
		etcdHost  string
		grpcAddr  string
	)

	flag.StringVar(&dbName, "db", "crawlerd", "database name")
	flag.BoolVar(&mongo, "mongo", true, "use mongodb as a database source")
	flag.StringVar(&mongoHost, "mongo-host", "", "mongo host")
	flag.BoolVar(&etcd, "etcd", true, "use etcd as a registry source")
	flag.StringVar(&etcdHost, "etcd-host", "", "etcd host")
	flag.StringVar(&grpcAddr, "grpc-addr", scheduler.DefaultSchedulerGRPCServerAddr, "grpc addr")
	flag.Parse()

	var opts []scheduler.Option

	if mongo {
		if mongoHost == "" {
			opts = append(opts, scheduler.WithMongoDBStorage(dbName))
		} else {
			opts = append(
				opts,
				scheduler.WithMongoDBStorage(dbName, options.Client().ApplyURI(
					fmt.Sprintf("mongodb://%s:27017", mongoHost),
				)),
			)
		}
	}

	if etcd {
		if etcdHost == "" {
			opts = append(opts, scheduler.WithWatcher())
		} else {
			opts = append(opts, scheduler.WithWatcher(
				scheduler.NewWatcherOption().WithETCD(clientv3.Config{
					Endpoints:   []string{etcdHost + ":2379"},
					DialTimeout: time.Second * 15,
				}),
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
