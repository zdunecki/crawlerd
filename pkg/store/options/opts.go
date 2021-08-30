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
	requestQueue    store.RequestQueue
	url             store.URL
	history         store.History
	registry        store.Registry
	job             store.Job
	linker          store.Linker
	runner          store.Runner
	runnerFunctions store.RunnerFunctions
}

type RepositoryOption struct {
	storage store.Repository

	err        error
	repository *repositoryOpt
	apply      func(o *RepositoryOption) // TODO: find better solution. apply should add changes from option to orignal WithStorage opts
}

type clientOpt struct {
}

func Client() *clientOpt {
	return new(clientOpt)
}

func WithStorage(opts ...*RepositoryOption) (store.Repository, error) {
	s := &repositoryOpt{}

	apply := func(o *RepositoryOption) {
		if o.repository.requestQueue != nil {
			s.requestQueue = o.repository.requestQueue
		}
		if o.repository.url != nil {
			s.url = o.repository.url
		}
		if o.repository.history != nil {
			s.history = o.repository.history
		}
		if o.repository.registry != nil {
			s.registry = o.repository.registry
		}
		if o.repository.job != nil {
			s.job = o.repository.job
		}
		if o.repository.linker != nil {
			s.linker = o.repository.linker
		}
		if o.repository.runner != nil {
			s.runner = o.repository.runner
		}
		if o.repository.runnerFunctions != nil {
			s.runnerFunctions = o.repository.runnerFunctions
		}
	}

	for _, o := range opts {
		if o.err != nil {
			return nil, o.err
		}
		apply(o)

		o.apply = apply
	}

	return s, nil
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
		repository: &repositoryOpt{
			requestQueue:    s.RequestQueue(),
			url:             s.URL(),
			history:         s.History(),
			registry:        s.Registry(),
			job:             s.Job(),
			linker:          s.Linker(),
			runner:          s.Runner(),
			runnerFunctions: s.RunnerFunctions(),
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

	s := etcdstore.NewStorage(etcd, registryTTLBuffer)

	return &RepositoryOption{
		storage: s,
		repository: &repositoryOpt{
			requestQueue:    s.RequestQueue(),
			url:             s.URL(),
			history:         s.History(),
			registry:        s.Registry(),
			job:             s.Job(),
			linker:          s.Linker(),
			runner:          s.Runner(),
			runnerFunctions: s.RunnerFunctions(),
		},
	}
}

// TODO: implementation
func (o *clientOpt) WithCache() *RepositoryOption {
	return &RepositoryOption{
		storage: cachestore.NewStorage(),
	}
}

func (s *repositoryOpt) RunnerFunctions() store.RunnerFunctions {
	return s.runnerFunctions
}

func (s *repositoryOpt) Runner() store.Runner {
	return s.runner
}

func (s *repositoryOpt) Linker() store.Linker {
	return s.linker
}

func (s *repositoryOpt) RequestQueue() store.RequestQueue {
	return s.requestQueue
}

func (s *repositoryOpt) URL() store.URL {
	return s.url
}

func (s *repositoryOpt) History() store.History {
	return s.history
}

func (s *repositoryOpt) Registry() store.Registry {
	return s.registry
}

func (s *repositoryOpt) Job() store.Job {
	return s.job
}

func (o *RepositoryOption) RequestQueue() *RepositoryOption {
	o.repository.requestQueue = o.storage.RequestQueue()
	return o
}

func (o *RepositoryOption) URL() *RepositoryOption {
	o.repository.url = o.storage.URL()
	return o
}

func (o *RepositoryOption) History() *RepositoryOption {
	o.repository.history = o.storage.History()
	return o
}

func (o *RepositoryOption) Registry() *RepositoryOption {
	o.repository.registry = o.storage.Registry()
	return o
}

func (o *RepositoryOption) Job() *RepositoryOption {
	o.repository.job = o.storage.Job()
	return o
}

func (o *RepositoryOption) Runner() *RepositoryOption {
	o.repository.runner = o.storage.Runner()
	return o
}

func (o *RepositoryOption) RunnerFunctions() *RepositoryOption {
	o.repository.runnerFunctions = o.storage.RunnerFunctions()
	return o
}

func (o *RepositoryOption) CustomRunnerFunctions(rf store.RunnerFunctions) *RepositoryOption {
	o.repository.runnerFunctions = rf
	return o
}

func (o *RepositoryOption) Apply() {
	if o.apply == nil {
		panic("apply is not defined")
		return
	}

	o.apply(o)
}
