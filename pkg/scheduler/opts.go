package scheduler

import (
	"time"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/mgostorage"
	"crawlerd/pkg/worker"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/client-go/kubernetes"
)

type k8sConfig struct {
	client    kubernetes.Interface
	namespace string
}

type (
	Option func(*scheduler) error

	WatcherOption struct {
		etcdConfig   *clientv3.Config
		k8sConfig    *k8sConfig
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

func (o *WatcherOption) WithK8s(client kubernetes.Interface, namespace string) *WatcherOption {
	o.k8sConfig = &k8sConfig{
		client:    client,
		namespace: namespace,
	}

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

		etcdLeasing := func(c clientv3.Config) error {
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
				s.log.Debug("use k8s leasing")
				workerCluster = worker.NewK8sCluster(o.k8sConfig.client, o.k8sConfig.namespace, s.clusterConfig)
				s.leasing = NewLeasing(workerCluster, s.server)
			} else if o.etcdConfig != nil {
				s.log.Debug("use etcd leasing")
				if err := etcdLeasing(*o.etcdConfig); err != nil {
					return err
				}
			}
		}

		if storage == nil {
			return ErrStorageIsRequired
		}

		if workerCluster == nil {
			s.log.Debug("use default etcd leasing")
			if err := etcdLeasing(DefaultETCDConfig); err != nil {
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

func WithWorkerClusterConfig(cfg *worker.Config) Option {
	return func(s *scheduler) error {
		s.clusterConfig = cfg

		return nil
	}
}
