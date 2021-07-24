package options

import (
	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/cachestorage"
	"crawlerd/pkg/storage/etcdstorage"
	"crawlerd/pkg/storage/mgostorage"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type repositoryOpt struct {
	url      storage.URLRepository
	history  storage.HistoryRepository
	registry storage.RegistryRepository
}

func (s *repositoryOpt) URL() storage.URLRepository {
	return s.url
}

func (s *repositoryOpt) History() storage.HistoryRepository {
	return s.history
}

func (s *repositoryOpt) Registry() storage.RegistryRepository {
	return s.registry
}

func withURL(r storage.URLRepository) repositoryOptFn {
	return func(s *repositoryOpt) {
		s.url = r
	}
}

func withHistory(r storage.HistoryRepository) repositoryOptFn {
	return func(s *repositoryOpt) {
		s.history = r
	}
}

func withRegistry(r storage.RegistryRepository) repositoryOptFn {
	return func(s *repositoryOpt) {
		s.registry = r
	}
}

type repositoryOptFn func(*repositoryOpt)

type RepositoryOption struct {
	storage storage.Storage
	options map[string]repositoryOptFn
	err     error
}

func (o *RepositoryOption) URL() *RepositoryOption {
	o.options["url"] = withURL(o.storage.URL())
	return o
}
func (o *RepositoryOption) History() *RepositoryOption {
	o.options["history"] = withHistory(o.storage.History())
	return o
}
func (o *RepositoryOption) Registry() *RepositoryOption {
	o.options["registry"] = withRegistry(o.storage.Registry())
	return o
}

type clientOpt struct {
}

func (o *clientOpt) WithMongoDB(urlDBName string, urlCfg *options.ClientOptions) *RepositoryOption {
	client, err := mgostorage.NewClient(urlCfg)
	if err != nil {
		return &RepositoryOption{
			err: err,
		}
	}

	if urlDBName != "" {
		client.SetDatabaseName(urlDBName)
	}

	s := mgostorage.NewStorage(client.DB())

	return &RepositoryOption{
		storage: s,
		options: map[string]repositoryOptFn{
			"url":      nil,
			"history":  nil,
			"registry": nil,
		},
	}
}

func (o *clientOpt) WithETCD(registryCfg clientv3.Config, registryTTLBuffer int64) *RepositoryOption {
	etcd, err := clientv3.New(registryCfg)
	if err != nil {
		return &RepositoryOption{
			err: err,
		}
	}

	etcdStorage := etcdstorage.NewStorage(etcd, registryTTLBuffer)

	return &RepositoryOption{
		storage: etcdStorage,
		options: map[string]repositoryOptFn{
			"url":      nil,
			"history":  nil,
			"registry": nil,
		},
	}
}

func (o *clientOpt) WithCache() *RepositoryOption {
	return &RepositoryOption{
		storage: cachestorage.NewStorage(),
		options: map[string]repositoryOptFn{
			"url":      nil,
			"history":  nil,
			"registry": nil,
		},
	}
}

func WithStorage(opts ...*RepositoryOption) (storage.Storage, error) {
	options := map[string]repositoryOptFn{
		"url":      nil,
		"history":  nil,
		"registry": nil,
	}

	for _, opt := range opts {
		if opt.err != nil {
			return nil, opt.err
		}

		for key, o := range opt.options {
			if o != nil {
				options[key] = o
			}
		}
	}

	s := newStorage(
		options["url"],
		options["history"],
		options["registry"],
	)

	return s, nil
}

func Client() *clientOpt {
	return new(clientOpt)
}

func newStorage(opts ...repositoryOptFn) storage.Storage {
	s := &repositoryOpt{}

	for _, o := range opts {
		o(s)
	}

	return s
}
