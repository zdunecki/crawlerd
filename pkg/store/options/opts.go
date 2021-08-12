package options

import (
	"crawlerd/pkg/store"
	"crawlerd/pkg/store/cachestore"
	"crawlerd/pkg/store/etcdstore"
	"crawlerd/pkg/store/mgostore"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type repositoryOpt struct {
	requestQueue store.RequestQueueRepository
	url          store.URLRepository
	history      store.HistoryRepository
	registry     store.RegistryRepository
	job          store.JobRepository
}

func (s *repositoryOpt) Linker() store.LinkerRepository {
	panic("implement me")
}

func (s *repositoryOpt) RequestQueue() store.RequestQueueRepository {
	return s.requestQueue
}

func (s *repositoryOpt) URL() store.URLRepository {
	return s.url
}

func (s *repositoryOpt) History() store.HistoryRepository {
	return s.history
}

func (s *repositoryOpt) Registry() store.RegistryRepository {
	return s.registry
}

func (s *repositoryOpt) Job() store.JobRepository {
	return s.job
}

func withRequestQueue(r store.RequestQueueRepository) repositoryOptFn {
	return func(s *repositoryOpt) {
		s.requestQueue = r
	}
}

func withURL(r store.URLRepository) repositoryOptFn {
	return func(s *repositoryOpt) {
		s.url = r
	}
}

func withHistory(r store.HistoryRepository) repositoryOptFn {
	return func(s *repositoryOpt) {
		s.history = r
	}
}

func withRegistry(r store.RegistryRepository) repositoryOptFn {
	return func(s *repositoryOpt) {
		s.registry = r
	}
}

func withJob(j store.JobRepository) repositoryOptFn {
	return func(s *repositoryOpt) {
		s.job = j
	}
}

type repositoryOptFn func(*repositoryOpt)

type RepositoryOption struct {
	storage store.Storage
	options map[string]repositoryOptFn
	err     error
}

func (o *RepositoryOption) RequestQueue() *RepositoryOption {
	o.options["request_queue"] = withRequestQueue(o.storage.RequestQueue())
	return o
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

func (o *RepositoryOption) Job() *RepositoryOption {
	o.options["job"] = withJob(o.storage.Job())
	return o
}

type clientOpt struct {
}

func (o *clientOpt) WithMongoDB(urlDBName string, urlCfg *options.ClientOptions) *RepositoryOption {
	client, err := mgostore.NewClient(urlCfg)
	if err != nil {
		return &RepositoryOption{
			err: err,
		}
	}

	if urlDBName != "" {
		client.SetDatabaseName(urlDBName)
	}

	s := mgostore.NewStore(client.DB())

	return &RepositoryOption{
		storage: s,
		options: map[string]repositoryOptFn{
			"request_queue": nil,
			"url":           nil,
			"history":       nil,
			"registry":      nil,
			"job":           nil,
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

	etcdStorage := etcdstore.NewStorage(etcd, registryTTLBuffer)

	return &RepositoryOption{
		storage: etcdStorage,
		options: map[string]repositoryOptFn{
			"request_queue": nil,
			"url":           nil,
			"history":       nil,
			"registry":      nil,
			"job":           nil,
		},
	}
}

func (o *clientOpt) WithCache() *RepositoryOption {
	return &RepositoryOption{
		storage: cachestore.NewStorage(),
		options: map[string]repositoryOptFn{
			"request_queue": nil,
			"url":           nil,
			"history":       nil,
			"registry":      nil,
			"job":           nil,
		},
	}
}

func WithStorage(opts ...*RepositoryOption) (store.Storage, error) {
	options := map[string]repositoryOptFn{
		"request_queue": nil,
		"url":           nil,
		"history":       nil,
		"registry":      nil,
		"job":           nil,
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
		options["request_queue"],
		options["url"],
		options["history"],
		options["registry"],
		options["job"],
	)

	return s, nil
}

func Client() *clientOpt {
	return new(clientOpt)
}

func newStorage(opts ...repositoryOptFn) store.Storage {
	s := &repositoryOpt{}

	for _, o := range opts {
		o(s)
	}

	return s
}
