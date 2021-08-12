package etcdstore

import (
	"crawlerd/pkg/store"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcd struct {
	registryRepo store.RegistryRepository
}

func (e etcd) RequestQueue() store.RequestQueueRepository {
	panic("implement me")
}

func (e etcd) Linker() store.LinkerRepository {
	panic("implement me")
}

func (e etcd) Job() store.JobRepository {
	panic("implement me")
}

type Storage interface {
	store.Storage
}

func NewStorage(c *clientv3.Client, registryTTLBuffer int64) Storage {
	return &etcd{
		registryRepo: NewRegistryRepository(c, registryTTLBuffer),
	}
}

// available via mgo storage
func (e etcd) URL() store.URLRepository {
	return nil
}

// available via mgo storage
func (e etcd) History() store.HistoryRepository {
	return nil
}

func (e etcd) Registry() store.RegistryRepository {
	return e.registryRepo
}
