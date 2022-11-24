package etcdstore

import (
	"github.com/zdunecki/crawlerd/pkg/store"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcd struct {
	registryRepo store.Registry
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

func (e etcd) RunnerFunctions() store.RunnerFunctions {
	return nil
}

func (e etcd) Runner() store.Runner {
	return nil
}

func (e etcd) RequestQueue() store.RequestQueue {
	return nil
}

func (e etcd) Linker() store.Linker {
	return nil
}

func (e etcd) Job() store.Job {
	return nil
}

func (e etcd) Seed() store.Seed {
	return nil
}
