package v1

import (
	"crawlerd/crawlerdpb"
	"crawlerd/pkg/scheduler"
	"crawlerd/pkg/store/mgostore"
	storeOptions "crawlerd/pkg/store/options"
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
		client, err := mgostore.NewClient(opts...)
		if err != nil {
			return err
		}

		if dbName != "" {
			client.SetDatabaseName(dbName)
		}

		a.store = mgostore.NewStore(client.DB())

		return nil
	}
}

func WithStore(opts ...*storeOptions.RepositoryOption) Option {
	return func(a *v1) error {
		store, err := storeOptions.WithStorage(opts...)

		if err != nil {
			return err
		}

		a.store = store

		return nil
	}
}
