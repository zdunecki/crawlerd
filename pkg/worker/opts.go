package worker

import (
	"time"

	"crawlerd/pkg/pubsub"
	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/etcdstorage"
	"crawlerd/pkg/storage/mgostorage"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Option func(*worker) error

	EtcdOption struct {
		config clientv3.Config
	}
)

var (
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

func WithPubSub(pubsub pubsub.PubSub) Option {
	return func(w *worker) error {
		w.pubsub = pubsub
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

// TODO: k8s cluster
func WithETCDCluster(opts ...*EtcdOption) Option {
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

		w.cluster = NewETCDCluster(etcd)

		return nil
	}
}

// TODO:
//func WithK8sCluster() {
//
//}

// deprecated: dont use it
func WithMongoDBStorage(dbName string, opts ...*options.ClientOptions) Option {
	return func(w *worker) error {
		client, err := mgostorage.NewClient(opts...)
		if err != nil {
			return err
		}

		if dbName != "" {
			client.SetDatabaseName(dbName)
		}

		w.storage = mgostorage.NewStorage(client.DB())

		return nil
	}
}

type repositoryOption struct {
	storage storage.Storage
	options map[string]storage.Option
	err     error
}

func (o *repositoryOption) URL() *repositoryOption {
	o.options["url"] = storage.WithURL(o.storage.URL())
	return o
}
func (o *repositoryOption) History() *repositoryOption {
	o.options["history"] = storage.WithHistory(o.storage.History())
	return o
}
func (o *repositoryOption) Registry() *repositoryOption {
	o.options["registry"] = storage.WithRegistry(o.storage.Registry())
	return o
}

type StorageOption struct {
}

func (o *StorageOption) WithMongoDB(urlDBName string, urlCfg *options.ClientOptions) *repositoryOption {
	client, err := mgostorage.NewClient(urlCfg)
	if err != nil {
		return &repositoryOption{
			err: err,
		}
	}

	if urlDBName != "" {
		client.SetDatabaseName(urlDBName)
	}

	s := mgostorage.NewStorage(client.DB())

	return &repositoryOption{
		storage: s,
		options: map[string]storage.Option{
			"url":      nil,
			"history":  nil,
			"registry": nil,
		},
	}
}

func (o *StorageOption) WithETCD(registryCfg clientv3.Config) *repositoryOption {
	etcd, err := clientv3.New(registryCfg)
	if err != nil {
		return &repositoryOption{
			err: err,
		}
	}

	etcdStorage := etcdstorage.NewStorage(etcd)

	return &repositoryOption{
		storage: etcdStorage,
		options: map[string]storage.Option{
			"url":      nil,
			"history":  nil,
			"registry": nil,
		},
	}
}

func WithStorage(opts ...*repositoryOption) Option {
	return func(w *worker) error {
		options := map[string]storage.Option{
			"url":      nil,
			"history":  nil,
			"registry": nil,
		}

		for _, opt := range opts {
			if opt.err != nil {
				return opt.err
			}

			for key, o := range opt.options {
				if o != nil {
					options[key] = o
				}
			}
		}

		s := storage.NewStorage(
			options["url"],
			options["history"],
			options["registry"],
		)

		w.storage = s

		return nil
	}
}

func NewStorageOption() *StorageOption {
	return new(StorageOption)
}
