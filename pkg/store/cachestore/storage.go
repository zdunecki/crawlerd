package cachestore

import "crawlerd/pkg/store"

type cachestorage struct {
	registryRepo store.RegistryRepository
}

func (c *cachestorage) RequestQueue() store.RequestQueueRepository {
	panic("implement me")
}

func (c *cachestorage) Linker() store.LinkerRepository {
	panic("implement me")
}

func (c *cachestorage) Job() store.JobRepository {
	panic("implement me")
}

type Storage interface {
	store.Storage
}

func NewStorage() Storage {
	return &cachestorage{
		registryRepo: NewRegistryRepository(),
	}
}

// available via mgo storage
func (c *cachestorage) URL() store.URLRepository {
	return nil
}

// available via mgo storage
func (c *cachestorage) History() store.HistoryRepository {
	return nil
}

func (c *cachestorage) Registry() store.RegistryRepository {
	return c.registryRepo
}
