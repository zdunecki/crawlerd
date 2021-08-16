package etcdstore

import (
	"bytes"
	"context"
	"encoding/json"
	fmt "fmt"

	"crawlerd/pkg/meta/v1"
	"crawlerd/pkg/store"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const minuteTTL = 60
const defaultTTLBuffer = 1 * minuteTTL
const KeyCrawlURL = "crawl"
const PrefixKeyCrawlURL = KeyCrawlURL + "."

// TODO: logger
type registry struct {
	etcd      *clientv3.Client
	ttlBuffer int64
}

func NewRegistryRepository(etcd *clientv3.Client, ttlBuffer int64) store.Registry {
	return &registry{
		etcd:      etcd,
		ttlBuffer: ttlBuffer,
	}
}

func (r *registry) GetURLByID(ctx context.Context, id int) (*v1.CrawlURL, error) {
	resp, err := r.etcd.Get(ctx, r.crawlID(id))
	if err != nil {
		return nil, err
	}

	exists := resp.Kvs != nil && len(resp.Kvs) > 0

	if !exists {
		return nil, err
	}

	var crawlURL *v1.CrawlURL

	kv := resp.Kvs[0]
	if err := json.NewDecoder(bytes.NewReader(kv.Value)).Decode(&crawlURL); err != nil {
		return nil, err
	}

	// bump crawl url ttl before expiration
	if _, err := r.etcd.KeepAliveOnce(context.Background(), clientv3.LeaseID(kv.Lease)); err != nil {
		return nil, err
	}

	return crawlURL, nil
}

// TODO: ttl should be from paramter not directly in storage
// TODO: lease ttl bump
func (r *registry) PutURL(ctx context.Context, url v1.CrawlURL) error {
	crawlUrlB, err := json.Marshal(url)
	if err != nil {
		return nil
	}

	buffer := r.ttlBuffer

	if buffer == 0 {
		buffer = defaultTTLBuffer
	}

	// lease is helpful for ensuring that if worker crash/restart key will expire soon and another process (Watcher.WatchNewURLs) take urls again to pool
	// TODO: it's not a ideal solution - find better
	lease, err := r.etcd.Grant(context.TODO(), url.Interval+buffer)
	if err != nil {
		return err
	}

	if _, err := r.etcd.Put(ctx, r.crawlID(int(url.Id)), string(crawlUrlB), clientv3.WithLease(lease.ID)); err != nil {
		return err
	}

	return nil
}

func (r *registry) DeleteURL(ctx context.Context, url v1.CrawlURL) error {
	if _, err := r.etcd.Delete(ctx, r.crawlID(int(url.Id))); err != nil {
		return err
	}

	return nil
}

func (r *registry) DeleteURLByID(ctx context.Context, id int) error {
	if _, err := r.etcd.Delete(ctx, r.crawlID(id)); err != nil {
		return err
	}

	return nil
}

func (r *registry) crawlID(id int) string {
	return fmt.Sprintf("%s%d", PrefixKeyCrawlURL, id)
}
