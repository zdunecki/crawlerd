package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"crawlerd/pkg/pubsub"
	"crawlerd/pkg/worker"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		dbName      string
		mongo       bool
		mongoHost   string
		etcd        bool
		etcdHost    string
		kafkaBroker string
	)

	flag.StringVar(&dbName, "db", "crawlerd", "database name")
	flag.BoolVar(&mongo, "mongo", true, "use mongodb as a database source")
	flag.StringVar(&mongoHost, "mongo-host", "", "mongo host")
	flag.BoolVar(&etcd, "etcd", true, "use etcd as a registry source")
	flag.StringVar(&etcdHost, "etcd-host", "", "etcd host")
	flag.StringVar(&kafkaBroker, "kafka-broker", "localhost:9093", "kafka broker")

	flag.Parse()

	var opts []worker.Option

	if etcd {
		if etcdHost == "" {
			opts = append(opts, worker.WithETCDCluster())
		} else {
			opts = append(opts, worker.WithETCDCluster(
				worker.NewETCDOption().ApplyConfig(clientv3.Config{
					Endpoints:   []string{etcdHost + ":2379"},
					DialTimeout: time.Second * 15,
				}),
			))
		}
	}

	if mongo {
		if mongoHost == "" {
			opts = append(opts, worker.WithMongoDBStorage(dbName))
		} else {
			opts = append(
				opts,
				worker.WithMongoDBStorage(dbName, options.Client().ApplyURI(
					fmt.Sprintf("mongodb://%s:27017", mongoHost),
				)),
			)
		}
	}

	pubsub, err := pubsub.NewKafka(kafkaBroker)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	opts = append(opts, worker.WithPubSub(pubsub))

	work, err := worker.New(opts...)

	if err != nil {
		log.Error(err)
		panic(err)
	}

	if err := work.Serve(context.Background()); err != nil {
		log.Error(err)
		panic(err)
	}
}
