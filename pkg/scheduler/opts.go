package scheduler

import (
	"time"

	"github.com/zdunecki/crawlerd/pkg/core/scheduler"
	"github.com/zdunecki/crawlerd/pkg/store"
	"github.com/zdunecki/crawlerd/pkg/store/options"
	"github.com/zdunecki/crawlerd/pkg/worker"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"
)

type k8sConfig struct {
	client         kubernetes.Interface
	namespace      string
	workerSelector string
}

type (
	Option func(*schedulerT) error

	WatcherOption struct {
		etcdConfig   *clientv3.Config
		k8sConfig    *k8sConfig
		timerTimeout *time.Duration
		storage      store.Repository
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

func (o *WatcherOption) WithK8s(client kubernetes.Interface, namespace, workerSelector string) *WatcherOption {
	o.k8sConfig = &k8sConfig{
		client:         client,
		namespace:      namespace,
		workerSelector: workerSelector,
	}

	return o
}

func (o *WatcherOption) WithTimerTimeout(t time.Duration) *WatcherOption {
	o.timerTimeout = &t

	return o
}

func (o *WatcherOption) WithStorage(s store.Repository) *WatcherOption {
	o.storage = s

	return o
}

func WithWatcher(opts ...*WatcherOption) Option {
	return func(s *schedulerT) error {
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
				workerCluster = worker.NewK8sCluster(o.k8sConfig.client, o.k8sConfig.namespace, o.k8sConfig.workerSelector, s.clusterConfig)
				s.leasing = NewLeasing(workerCluster, s.server)
			} else if o.etcdConfig != nil {
				s.log.Debug("use etcd leasing")
				if err := etcdLeasing(*o.etcdConfig); err != nil {
					return err
				}
			}
		}

		if storage == nil {
			return scheduler.ErrStorageIsRequired
		}

		if storage.URL() == nil {
			return scheduler.ErrURLRepositoryIsRequired
		}

		if storage.Registry() == nil {
			return scheduler.ErrRegistryRepositoryIsRequired
		}

		if workerCluster == nil {
			s.log.Debug("use default etcd leasing")
			if err := etcdLeasing(DefaultETCDConfig); err != nil {
				return err
			}
		}

		if workerCluster == nil {
			return scheduler.ErrClusterIsRequired
		}

		s.watcher = NewWatcher(workerCluster, storage.Linker(), timerTimeOut)

		return nil
	}
}

func WithStorage(opts ...*options.RepositoryOption) Option {
	return func(s *schedulerT) error {
		storage, err := options.WithStorage(opts...)

		if err != nil {
			return err
		}

		s.storage = storage

		return nil
	}
}

func WithWorkerClusterConfig(cfg *worker.Config) Option {
	return func(s *schedulerT) error {
		s.clusterConfig = cfg

		return nil
	}
}
