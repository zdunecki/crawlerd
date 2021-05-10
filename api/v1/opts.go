package v1

import (
	"crawlerd/crawlerdpb"
	"crawlerd/pkg/scheduler"
	"crawlerd/pkg/storage/mgostorage"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

type Option func(*v1) error

func WithGRPCSchedulerServer(addr string) Option {
	return func(a *v1) error {
		if addr == "" {
			addr = scheduler.DefaultSchedulerGRPCServerAddr
		}

		//TODO:
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		//defer conn.Close() TODO:

		a.scheduler = crawlerdpb.NewSchedulerClient(conn)
		return nil
	}
}

func WithMongoDBStorage(dbName string, opts ...*options.ClientOptions) Option {
	return func(a *v1) error {
		client, err := mgostorage.NewClient(opts...)
		if err != nil {
			return err
		}

		if dbName != "" {
			client.SetDatabaseName(dbName)
		}

		a.storage = mgostorage.NewStorage(client.DB())

		return nil
	}
}
