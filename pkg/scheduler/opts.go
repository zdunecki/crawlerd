package scheduler

import (
	"time"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/mgostorage"
	"crawlerd/pkg/worker"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Option func(*scheduler) error

	WatcherOption struct {
		etcdConfig   *clientv3.Config
		k8sConfig    error // TODO:
		timerTimeout *time.Duration
		storage      storage.Storage
		clusterType  worker.ClusterType
	}
)

var (
	DefaultETCDConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 15,
	}
	DefaultTimerTimeout = time.Minute * 1
)

func NewWatcherOption() *WatcherOption {
	return &WatcherOption{}
}

func (o *WatcherOption) WithETCD(cfg clientv3.Config) *WatcherOption {
	o.etcdConfig = &cfg

	return o
}

func (o *WatcherOption) WithTimerTimeout(t time.Duration) *WatcherOption {
	o.timerTimeout = &t

	return o
}

func (o *WatcherOption) WithStorage(s storage.Storage) *WatcherOption {
	o.storage = s

	return o
}

func WithWatcher(opts ...*WatcherOption) Option {
	return func(s *scheduler) error {
		timerTimeOut := DefaultTimerTimeout
		storage := s.storage
		var workerCluster worker.Cluster

		setupETCD := func(c clientv3.Config) error {
			etcd, err := clientv3.New(c)
			if err != nil {
				return err
			}

			workerCluster = worker.NewETCDCluster(etcd)
			s.leasing = NewLeasing(workerCluster, s.server)

			return nil
		}

		for _, o := range opts {
			if o.timerTimeout != nil {
				timerTimeOut = *o.timerTimeout
			}

			if o.storage != nil {
				storage = o.storage
			}

			if o.k8sConfig != nil {
				// TODO: k8s
			} else if o.etcdConfig != nil {
				if err := setupETCD(*o.etcdConfig); err != nil {
					return err
				}
			} else {
				if err := setupETCD(DefaultETCDConfig); err != nil {
					return err
				}
			}
		}

		if storage == nil {
			return ErrStorageIsRequired
		}

		if workerCluster == nil {
			if err := setupETCD(DefaultETCDConfig); err != nil {
				return err
			}
		}

		if workerCluster == nil {
			return ErrClusterIsRequired
		}

		s.watcher = NewWatcher(workerCluster, storage.URL(), timerTimeOut)

		return nil
	}
}

func WithMongoDBStorage(dbName string, opts ...*options.ClientOptions) Option {
	return func(s *scheduler) error {
		client, err := mgostorage.NewClient(opts...)
		if err != nil {
			return err
		}

		if dbName != "" {
			client.SetDatabaseName(dbName)
		}

		s.storage = mgostorage.NewStorage(client.DB())

		return nil
	}
}
