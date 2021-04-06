package worker

import (
	"context"
	"time"

	"crawlerd/pkg/storage/mgostorage"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Option func(*worker) error

	EtcdOption struct {
		config clientv3.Config
	}
)

var (
	DefaultMongoAddr  = "mongodb://localhost:27017"
	DefaultETCDConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 15,
	}
)

func WithSchedulerGRPCAddr(addr string) Option {
	return func(w *worker) error {
		w.schedulerAddr = addr
		return nil
	}
}

func NewETCDOption() *EtcdOption {
	return &EtcdOption{}
}

func (o *EtcdOption) ApplyConfig(cfg clientv3.Config) *EtcdOption {
	o.config = cfg

	return o
}

func WithETCDRegistry(opts ...*EtcdOption) Option {
	return func(w *worker) error {
		cfg := DefaultETCDConfig

		var optsCfg *clientv3.Config

		for _, o := range opts {
			optsCfg = &o.config
		}

		if optsCfg != nil {
			cfg = *optsCfg
		}

		etcd, err := clientv3.New(cfg)

		if err != nil {
			return err
		}

		w.registry = NewETCDRegistry(etcd)

		return nil
	}
}

func WithMongoDBStorage(dbName string, opts ...*options.ClientOptions) Option {
	return func(w *worker) error {
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
		w.storage = mgostorage.NewClient(db)

		return nil
	}
}
