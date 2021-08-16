package cachestore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"crawlerd/internal/cache"
	"crawlerd/pkg/meta/v1"
	"crawlerd/pkg/store"
	"github.com/allegro/bigcache/v3"
)

const minuteTTL = 60
const ttlBuffer = 1 * minuteTTL
const KeyCrawlURL = "crawl"
const PrefixKeyCrawlURL = KeyCrawlURL + "."

type registry struct {
	cache cache.Cache
}

//TODO: finish, cache implementation from argument
func NewRegistryRepository() store.Registry {
	return &registry{
		cache: cache.NewCache(bigcache.DefaultConfig(0)),
	}
}

func (r *registry) GetURLByID(ctx context.Context, id int) (*v1.CrawlURL, error) {
	resp, err := r.cache.Get(r.crawlID(id))
	if err != nil {
		return nil, err
	}

	var crawlURL *v1.CrawlURL

	if err := json.NewDecoder(bytes.NewReader(resp)).Decode(&crawlURL); err != nil {
		return nil, err
	}

	return crawlURL, nil
}

func (r *registry) PutURL(ctx context.Context, url v1.CrawlURL) error {
	crawlUrlB, err := json.Marshal(url)
	if err != nil {
		return nil
	}

	// TODO: does in-memory registry need ttl? worker does crawl until process is live
	//ttl := time.Second * time.Duration(url.Interval+ttlBuffer)

	if err := r.cache.Set(r.crawlID(int(url.Id)), crawlUrlB); err != nil {
		return err
	}

	return nil
}

func (r *registry) DeleteURL(ctx context.Context, url v1.CrawlURL) error {
	if err := r.cache.Del(r.crawlID(int(url.Id))); err != nil {
		return err
	}

	return nil
}

func (r *registry) DeleteURLByID(ctx context.Context, id int) error {
	if err := r.cache.Del(r.crawlID(id)); err != nil {
		return err
	}
	return nil
}

func (r *registry) crawlID(id int) string {
	return fmt.Sprintf("%s%d", PrefixKeyCrawlURL, id)
}
