package cachestore

import (
	"crawlerd/pkg/runner"
	"crawlerd/pkg/store"
)

type cachestorage struct {
	registryRepo store.Registry
}

func (c *cachestorage) Runner() runner.Runner {
	panic("implement me")
}

func (c *cachestorage) RequestQueue() store.RequestQueue {
	panic("implement me")
}

func (c *cachestorage) Linker() store.Linker {
	panic("implement me")
}

func (c *cachestorage) Job() store.Job {
	panic("implement me")
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
