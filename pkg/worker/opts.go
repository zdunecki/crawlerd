package worker

import (
	"time"

	"crawlerd/pkg/pubsub"
	"crawlerd/pkg/storage/options"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"
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
func WithETCDCluster(optCfg ...clientv3.Config) Option {
	return func(w *worker) error {
		cfg := DefaultETCDConfig

		for _, c := range optCfg {
			cfg = c
		}

		etcd, err := clientv3.New(cfg)

		if err != nil {
			return err
		}

		w.cluster = NewETCDCluster(etcd)

		return nil
	}
}

func WithK8sCluster(client kubernetes.Interface, namespace string) Option {
	return func(w *worker) error {
		w.cluster = NewK8sCluster(client, namespace, w.config)

		return nil
	}
}

func WithStorage(opts ...*options.RepositoryOption) Option {
	return func(w *worker) error {
		s, err := options.WithStorage(opts...)

		if err != nil {
			return err
		}

		w.storage = s

		return nil
	}
}
