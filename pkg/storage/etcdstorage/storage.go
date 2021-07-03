package etcdstorage

import (
	"crawlerd/pkg/storage"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcd struct {
	registryRepo storage.RegistryRepository
}

type Storage interface {
	storage.Storage
}

func NewStorage(c *clientv3.Client) Storage {
	return &etcd{
		registryRepo: NewRegistryRepository(c),
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
