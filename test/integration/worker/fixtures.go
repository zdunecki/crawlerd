package worker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/mgostorage"
	"crawlerd/pkg/worker"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newFixtureStorage() (storage.Client, error) {
	var (
		dbName    = "crawlerd-test"
		mongoHost = os.Getenv("MONGO_HOST")
	)

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(
		fmt.Sprintf("mongodb://%s:27017", mongoHost),
	))
	if err != nil {
		return nil, err
	}

	storage := mgostorage.NewStorage(client.Database(dbName))

	return storage, nil
}
func newFixtures() (worker.Worker, error) {
	var (
		dbName        = "crawlerd-test"
		schedulerAddr = os.Getenv("SCHEDULER_ADDR")
		mongoHost     = os.Getenv("MONGO_HOST")
		etcdHost      = os.Getenv("ETCD_HOST")
	)

	if schedulerAddr == "" {
		return nil, errors.New("scheduler addr is required")
	}

	if mongoHost == "" {
		return nil, errors.New("mongo host is required")
	}

	if etcdHost == "" {
		return nil, errors.New("etcd host is required")
	}

	work, err := worker.New(
		worker.WithSchedulerGRPCAddr(schedulerAddr),
		worker.WithETCDRegistry(
			worker.NewETCDOption().ApplyConfig(clientv3.Config{
				Endpoints:   []string{etcdHost + ":2379"},
				DialTimeout: time.Second * 15,
			}),
		),
		worker.WithMongoDBStorage(dbName, options.Client().ApplyURI(
			fmt.Sprintf("mongodb://%s:27017", mongoHost),
		)),
	)

	return work, err
}
