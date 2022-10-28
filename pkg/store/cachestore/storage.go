package cachestore

import (
	"crawlerd/pkg/store"
)

type cachestorage struct {
	registryRepo store.Registry
}

type Storage interface {
	store.Repository
}

func NewStorage() Storage {
	return &cachestorage{
		registryRepo: NewRegistryRepository(),
	}
}

// available via mgo storage
func (c *cachestorage) URL() store.URL {
	return nil
}

// available via mgo storage
func (c *cachestorage) History() store.History {
	return nil
}

func (c *cachestorage) Registry() store.Registry {
	return c.registryRepo
}

func (c *cachestorage) RunnerFunctions() store.RunnerFunctions {
	return nil
}

func (c *cachestorage) Runner() store.Runner {
	return nil

}

func (c *cachestorage) RequestQueue() store.RequestQueue {
	return nil
}

func (c *cachestorage) Linker() store.Linker {
	return nil
}

func (c *cachestorage) Job() store.Job {
	return nil
}

func (c *cachestorage) Seed() store.Seed {
	return nil
}
