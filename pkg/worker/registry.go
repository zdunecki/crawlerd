package worker

import (
	"bytes"
	"context"
	"encoding/json"
	fmt "fmt"
	"strconv"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Registry interface {
	WorkerID() string

	RegisterWorker() error
	UnregisterWorker() error

	GetURLByID(int) (*CrawlURL, error)
	PutURL(CrawlURL) error
	DeleteURL(CrawlURL) error
	FindURLByWorkerID(string) ([]CrawlURL, error)
	DeleteURLByID(int) error

	WithWorker(Worker)

	crawlID(int) string
}

type etcdregistry struct {
	etcd *clientv3.Client

	worker Worker
}

func NewETCDRegistry(etcd *clientv3.Client) Registry {
	return &etcdregistry{
		etcd: etcd,
	}
}

func (r *etcdregistry) WorkerID() string {
	return r.worker.ID()
}

func (r *etcdregistry) GetURLByID(id int) (*CrawlURL, error) {
	resp, err := r.etcd.Get(context.Background(), r.crawlID(id))
	if err != nil {
		return nil, err
	}

	exists := resp.Kvs != nil && len(resp.Kvs) > 0

	if !exists {
		return nil, err
	}

	var crawlURL *CrawlURL

	if err := json.NewDecoder(bytes.NewReader(resp.Kvs[0].Value)).Decode(&crawlURL); err != nil {
		return nil, err
	}

	return crawlURL, nil
}

func (r *etcdregistry) PutURL(url CrawlURL) error {
	crawlUrlB, err := json.Marshal(url)
	if err != nil {
		return nil
	}

	if _, err := r.etcd.Put(context.Background(), r.crawlID(int(url.Id)), string(crawlUrlB)); err != nil {
		return err
	}

	return nil
}

func (r *etcdregistry) DeleteURL(url CrawlURL) error {
	if _, err := r.etcd.Delete(context.Background(), r.crawlID(int(url.Id))); err != nil {
		return err
	}

	return nil
}

//TODO: scroll
func (r *etcdregistry) FindURLByWorkerID(id string) ([]CrawlURL, error) {
	var result []CrawlURL

	if resp, err := r.etcd.Get(context.Background(), "crawl.", clientv3.WithPrefix()); err != nil {
		return nil, err
	} else {
		exists := resp.Kvs != nil && len(resp.Kvs) > 0

		if !exists {
			return nil, err
		}

		for _, kv := range resp.Kvs {
			var crawlURL *CrawlURL

			if err := json.NewDecoder(bytes.NewReader(kv.Value)).Decode(&crawlURL); err != nil {
				return nil, err
			}

			if crawlURL.WorkerID != id {
				continue
			}

			if _, err := r.etcd.Delete(context.Background(), string(kv.Key)); err != nil {
				return nil, err
			}

			result = append(result, *crawlURL)
		}
	}

	return result, nil
}

func (r *etcdregistry) DeleteURLByID(id int) error {
	if _, err := r.etcd.Delete(context.Background(), r.crawlID(id)); err != nil {
		return err
	}

	return nil
}

func (r *etcdregistry) UnregisterWorker() error {
	if _, err := r.etcd.Delete(context.Background(), "worker."+r.worker.ID()); err != nil {
		return err
	}

	return nil
}

func (r *etcdregistry) RegisterWorker() error {
	if _, err := r.etcd.Put(context.Background(), "worker."+r.worker.ID(), r.worker.Addr()); err != nil {
		return err
	}

	return nil
}

func (r *etcdregistry) WithWorker(w Worker) {
	r.worker = w
}

func (r *etcdregistry) crawlID(id int) string {
	return fmt.Sprintf("crawl.%s", strconv.Itoa(id))
}
