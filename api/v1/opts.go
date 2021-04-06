package v1

import (
	"context"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/scheduler"
	"crawlerd/pkg/storage/mgostorage"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

type Option func(*v1) error

var (
	DefaultMongoAddr  = "mongodb://localhost:27017"
	DefaultETCDConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 15,
	}
)

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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var passOpts []*options.ClientOptions

		passOpts = append(passOpts, options.Client().ApplyURI(DefaultMongoAddr))
		for _, o := range opts {
			passOpts = append(passOpts, o)
		}

		client, err := mongo.Connect(ctx, passOpts...)
		if err != nil {
			return err
		}

		db := client.Database(dbName)
		a.storage = mgostorage.NewClient(db)

		return nil
	}
}
