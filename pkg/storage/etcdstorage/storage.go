package etcdstorage

import (
	"crawlerd/pkg/storage"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcd struct {
	registryRepo storage.RegistryRepository
}

func (e etcd) Job() storage.JobRepository {
	panic("implement me")
}

type Storage interface {
	storage.Storage
}

func NewStorage(c *clientv3.Client, registryTTLBuffer int64) Storage {
	return &etcd{
		registryRepo: NewRegistryRepository(c, registryTTLBuffer),
	}
}

// available via mgo storage
func (e etcd) URL() storage.URLRepository {
	return nil
}

// available via mgo storage
func (e etcd) History() storage.HistoryRepository {
	return nil
}

func (e etcd) Registry() storage.RegistryRepository {
	return e.registryRepo
}
