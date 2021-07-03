package e2e

import (
	"context"
	"testing"
	"time"

	"crawlerd/api/v1"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestCrawlOneURL(t *testing.T) {
	setup, done, err := setupClient()
	if err != nil {
		t.Error(err)
		return
	}
	defer done()
	crawldURL := setup.crawld.URL()

	setup.etcdContainer.DefaultAddress()
	createResp, err := crawldURL.Create(&v1.RequestPostURL{
		URL:      "https://httpbin.org/range/1",
		Interval: 15,
	})
	if err != nil {
		t.Error(err)
		return
	}

	if createResp.ID != 0 {
		t.Error("invalid expected created id")
		return
	}

	urls, err := crawldURL.All()
	if err != nil {
		t.Error(err)
		return
	}

	if len(urls) != 1 {
		t.Error("invalid expected urls length")
		return
	}

	etcdCfg := clientv3.Config{
		Endpoints:   []string{setup.etcdContainer.DefaultAddress()}, // get host from container
		DialTimeout: time.Second * 15,
	}

	etcd, err := clientv3.New(etcdCfg)
	if err != nil {
		t.Error(err)
		return
	}

	workerKv, err := etcd.Get(context.Background(), "worker", clientv3.WithPrefix())
	if err != nil {
		t.Error(err)
		return
	}

	if len(workerKv.Kvs) != 1 {
		t.Error("invalid expected running crawl workers length")
		return
	}

	runningCrawlUrlsKv, err := etcd.Get(context.Background(), "crawl", clientv3.WithPrefix())
	if err != nil {
		t.Error(err)
		return
	}

	if len(runningCrawlUrlsKv.Kvs) != 1 {
		t.Error("invalid expected running crawl urls length")
		return
	}
}
