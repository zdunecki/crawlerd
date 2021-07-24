package main

import (
	"flag"
	"fmt"

	"crawlerd/api"
	"crawlerd/api/v1"
	"crawlerd/pkg/scheduler"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		dbName string

		mongo     bool
		mongoHost string
		mongoPort string

		grpc bool

		schedulerAddr string

		addr string
	)

	flag.StringVar(&dbName, "db", "crawlerd", "database name")
	flag.BoolVar(&mongo, "mongo", true, "use mongodb as a database source")
	flag.StringVar(&mongoHost, "mongo-host", "", "mongo host")
	flag.StringVar(&mongoPort, "mongo-port", "27017", "mongo port")
	flag.BoolVar(&grpc, "grpc", true, "use grpc as a scheduler serer")
	flag.StringVar(&schedulerAddr, "scheduler-addr", scheduler.DefaultSchedulerGRPCServerAddr, "scheduler grpc addr")
	flag.StringVar(&addr, "addr", ":8080", "http server addr")

	flag.Parse()

	var opts []v1.Option

	if mongo {
		if mongoHost == "" {
			opts = append(opts, v1.WithMongoDBStorage(dbName))
		} else {
			opts = append(opts, v1.WithMongoDBStorage(dbName, options.Client().ApplyURI(
				fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort),
			)))
		}
	}
	if grpc {
		opts = append(opts, v1.WithGRPCSchedulerServer(schedulerAddr))
	}

	apiV1, err := v1.New(opts...)
	if err != nil {
		panic(err)
	}

	if err := apiV1.Serve(addr, api.New(chi.NewRouter())); err != nil {
		panic(err)
	}
}
