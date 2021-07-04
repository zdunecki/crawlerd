package etcdstorage

import (
	"bytes"
	"context"
	"encoding/json"
	fmt "fmt"

	"crawlerd/pkg/storage"
	"crawlerd/pkg/storage/objects"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const PrefixKeyCrawlURL = "crawl."

type registry struct {
	etcd *clientv3.Client
}

func NewRegistryRepository(etcd *clientv3.Client) storage.RegistryRepository {
	return &registry{
		etcd: etcd,
	}
}

func (r *registry) GetURLByID(ctx context.Context, id int) (*objects.CrawlURL, error) {
	resp, err := r.etcd.Get(ctx, r.crawlID(id))
	if err != nil {
		return nil, err
	}

	exists := resp.Kvs != nil && len(resp.Kvs) > 0

	if !exists {
		return nil, err
	}

	var crawlURL *objects.CrawlURL

	if err := json.NewDecoder(bytes.NewReader(resp.Kvs[0].Value)).Decode(&crawlURL); err != nil {
		return nil, err
	}

	return crawlURL, nil
}

func (r *registry) PutURL(ctx context.Context, url objects.CrawlURL) error {
	crawlUrlB, err := json.Marshal(url)
	if err != nil {
		return nil
	}

	if _, err := r.etcd.Put(ctx, r.crawlID(int(url.Id)), string(crawlUrlB)); err != nil {
		return err
	}

	return nil
}

func (r *registry) DeleteURL(ctx context.Context, url objects.CrawlURL) error {
	if _, err := r.etcd.Delete(ctx, r.crawlID(int(url.Id))); err != nil {
		return err
	}

	return nil
}

//TODO: scroll
func (r *registry) FindURLByWorkerID(ctx context.Context, id string) ([]objects.CrawlURL, error) {
	var result []objects.CrawlURL

	if resp, err := r.etcd.Get(ctx, PrefixKeyCrawlURL, clientv3.WithPrefix()); err != nil {
		return nil, err
	} else {
		exists := resp.Kvs != nil && len(resp.Kvs) > 0

		if !exists {
			return nil, err
		}

		for _, kv := range resp.Kvs {
			var crawlURL *objects.CrawlURL

			if err := json.NewDecoder(bytes.NewReader(kv.Value)).Decode(&crawlURL); err != nil {
				return nil, err
			}

			if crawlURL.WorkerID != id {
				continue
			}

			if _, err := r.etcd.Delete(ctx, string(kv.Key)); err != nil {
				return nil, err
			}

			result = append(result, *crawlURL)
		}
	}

	return result, nil
}

func (r *registry) DeleteURLByID(ctx context.Context, id int) error {
	if _, err := r.etcd.Delete(ctx, r.crawlID(id)); err != nil {
		return err
	}

	return nil
}

func (r *registry) crawlID(id int) string {
	return fmt.Sprintf("%s.%d", PrefixKeyCrawlURL, id)
}
