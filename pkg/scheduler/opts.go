package scheduler

import (
	"time"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/mgostorage"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Option func(*scheduler) error

	WatcherOption struct {
		config       *clientv3.Config
		timerTimeout *time.Duration
		storage      storage.Client
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

func (o *WatcherOption) ApplyConfig(cfg clientv3.Config) *WatcherOption {
	o.config = &cfg

	return o
}

func (o *WatcherOption) ApplyTimerTimeout(t time.Duration) *WatcherOption {
	o.timerTimeout = &t

	return o
}

func (o *WatcherOption) ApplyStorage(s storage.Client) *WatcherOption {
	o.storage = s

	return o
}

func WithETCDWatcher(opts ...*WatcherOption) Option {
	return func(s *scheduler) error {
		cfg := DefaultETCDConfig
		timerTimeOut := DefaultTimerTimeout
		storage := s.storage

		for _, o := range opts {
			if o.config != nil {
				cfg = *o.config
			}

			if o.timerTimeout != nil {
				timerTimeOut = *o.timerTimeout
			}

			if o.storage != nil {
				storage = o.storage
			}
		}

		etcd, err := clientv3.New(cfg)
		if err != nil {
			return err
		}

		s.leasing = NewETCDLeasing(etcd, s.server)

		if storage == nil {
			return ErrStorageIsRequired
		}

		s.watcher = NewETCDWatcher(etcd, storage, timerTimeOut)

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
