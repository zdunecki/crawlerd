package cachestorage

import "crawlerd/pkg/storage"

type cachestorage struct {
	registryRepo storage.RegistryRepository
}

func (c *cachestorage) Job() storage.JobRepository {
	panic("implement me")
}

type Storage interface {
	storage.Storage
}

func NewStorage() Storage {
	return &cachestorage{
		registryRepo: NewRegistryRepository(),
	}
}

// available via mgo storage
func (c *cachestorage) URL() storage.URLRepository {
	return nil
}

// available via mgo storage
func (c *cachestorage) History() storage.HistoryRepository {
	return nil
}

func (c *cachestorage) Registry() storage.RegistryRepository {
	return c.registryRepo
}
