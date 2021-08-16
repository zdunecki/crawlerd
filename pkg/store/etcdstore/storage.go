package etcdstore

import (
	"crawlerd/pkg/runner"
	"crawlerd/pkg/store"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcd struct {
	registryRepo store.Registry
}

func (e etcd) Runner() runner.Runner {
	panic("implement me")
}

func (e etcd) RequestQueue() store.RequestQueue {
	panic("implement me")
}

func (e etcd) Linker() store.Linker {
	panic("implement me")
}

func (e etcd) Job() store.Job {
	panic("implement me")
}

type Storage interface {
	store.Repository
}

func NewStorage(c *clientv3.Client, registryTTLBuffer int64) Storage {
	return &etcd{
		registryRepo: NewRegistryRepository(c, registryTTLBuffer),
	}
}

// available via mgo storage
func (e etcd) URL() store.URL {
	return nil
}

// available via mgo storage
func (e etcd) History() store.History {
	return nil
}

func (e etcd) Registry() store.Registry {
	return e.registryRepo
}
