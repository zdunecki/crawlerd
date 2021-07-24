package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"crawlerd/api/v1"
	"crawlerd/pkg/storage/etcdstorage"
	"crawlerd/pkg/worker"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestCrawlOneURL(t *testing.T) {
	os.Setenv("DEBUG", "1")
	os.Setenv("WORKER_HOST", "localhost")

	setup, err := setupClient(&setupOptions{})
	if err != nil {
		t.Error(err)
		return
	}
	defer setup.done()
	crawldURL := setup.crawld.URL()

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

	workerKv, err := etcd.Get(context.Background(), worker.KeyWorker, clientv3.WithPrefix())
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

func TestCrawlOneURLWithMultipleWorkers(t *testing.T) {
	os.Setenv("DEBUG", "1")
	os.Setenv("WORKER_HOST", "localhost")

	additionalWorkerCount := 1
	setup, err := setupClient(&setupOptions{
		additionalWorkerCount: additionalWorkerCount,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer setup.done()
	crawldURL := setup.crawld.URL()

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

	workerKv, err := etcd.Get(context.Background(), worker.KeyWorker, clientv3.WithPrefix())
	if err != nil {
		t.Error(err)
		return
	}

	expectWorkers := 1 + additionalWorkerCount
	if len(workerKv.Kvs) != expectWorkers {
		t.Error("invalid expected running crawl workers length")
		return
	}

	runningCrawlUrlsKv, err := etcd.Get(context.Background(), "crawl", clientv3.WithPrefix())
	if err != nil {
		t.Error(err)
		return
	}

	expectRunningCrawls := 1
	if len(runningCrawlUrlsKv.Kvs) != expectRunningCrawls {
		t.Error("invalid expected running crawl urls length")
		return
	}
}

func TestCrawlWithGracefullShutDown(t *testing.T) {
	os.Setenv("DEBUG", "1")
	os.Setenv("WORKER_HOST", "localhost")

	setup, err := setupClient(&setupOptions{})
	if err != nil {
		t.Error(err)
		return
	}
	defer setup.done()

	etcdCfg := clientv3.Config{
		Endpoints:   []string{setup.etcdContainer.DefaultAddress()}, // get host from container
		DialTimeout: time.Second * 15,
	}

	etcd, err := clientv3.New(etcdCfg)
	if err != nil {
		t.Error(err)
		return
	}

	// worker should be registered before cancel
	{
		workerKv, err := etcd.Get(context.Background(), worker.KeyWorker, clientv3.WithPrefix())
		if err != nil {
			t.Error(err)
			return
		}

		if len(workerKv.Kvs) != 1 {
			t.Error("invalid expected running crawl workers length")
			return
		}
	}

	crawldURL := setup.crawld.URL()

	// crawl url successfully
	{
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
	}

	// crawl url should exists in registry after api request
	{
		workerKv, err := etcd.Get(context.Background(), etcdstorage.KeyCrawlURL, clientv3.WithPrefix())
		if err != nil {
			t.Error(err)
			return
		}

		if len(workerKv.Kvs) == 0 {
			t.Error("invalid expected running crawl workers length")
			return
		}
	}

	workere2e := setup.worker_e2e[0]
	// cancel worker
	workere2e.ctxCancel()

	// wait for graceful end
wait:
	for {
		select {
		case <-workere2e.done():
			break wait
		}
	}

	// worker should be unregistered
	{
		workerKv, err := etcd.Get(context.Background(), worker.KeyWorker, clientv3.WithPrefix())
		if err != nil {
			t.Error(err)
			return
		}

		if len(workerKv.Kvs) != 0 {
			t.Error("invalid expected running crawl workers length")
			return
		}
	}

	// crawl url should be empty because worker shutdown
	{
		workerKv, err := etcd.Get(context.Background(), etcdstorage.KeyCrawlURL, clientv3.WithPrefix())
		if err != nil {
			t.Error(err)
			return
		}

		if len(workerKv.Kvs) != 0 {
			t.Error("invalid expected running crawl workers length")
			return
		}
	}
}

func TestCrawlWithGracefullShutDownWithMultipleWorkers(t *testing.T) {
	os.Setenv("DEBUG", "1")
	os.Setenv("WORKER_HOST", "localhost")

	additionalWorkerCount := 1
	setup, err := setupClient(&setupOptions{
		additionalWorkerCount: additionalWorkerCount,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer setup.done()

	etcdCfg := clientv3.Config{
		Endpoints:   []string{setup.etcdContainer.DefaultAddress()}, // get host from container
		DialTimeout: time.Second * 15,
	}

	etcd, err := clientv3.New(etcdCfg)
	if err != nil {
		t.Error(err)
		return
	}

	// worker should be registered before cancel
	{
		workerKv, err := etcd.Get(context.Background(), worker.KeyWorker, clientv3.WithPrefix())
		if err != nil {
			t.Error(err)
			return
		}

		expectWorkersCounts := 1 + additionalWorkerCount
		if len(workerKv.Kvs) != expectWorkersCounts {
			t.Error("invalid expected running crawl workers length")
			return
		}
	}

	crawldURL := setup.crawld.URL()

	// crawl url successfully
	{
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
	}

	// crawl url should exists in registry after api request
	{
		workerKv, err := etcd.Get(context.Background(), etcdstorage.KeyCrawlURL, clientv3.WithPrefix())
		if err != nil {
			t.Error(err)
			return
		}

		if len(workerKv.Kvs) != 1 {
			t.Error("invalid expected running crawl workers length")
			return
		}
	}

	workere2e := setup.worker_e2e[0]
	// cancel worker
	workere2e.ctxCancel()

	// wait for graceful end
wait:
	for {
		select {
		case <-workere2e.done():
			break wait
		}
	}

	// worker should be unregistered
	{
		workerKv, err := etcd.Get(context.Background(), worker.KeyWorker, clientv3.WithPrefix())
		if err != nil {
			t.Error(err)
			return
		}

		expectRunningWorkers := additionalWorkerCount // how many workers after one cancelled

		if len(workerKv.Kvs) != expectRunningWorkers {
			t.Error("invalid expected running crawl workers length")
			return
		}
	}

	// crawl url should be still one because of Controller.ReAttachResources
	{
		workerKv, err := etcd.Get(context.Background(), etcdstorage.KeyCrawlURL, clientv3.WithPrefix())
		if err != nil {
			t.Error(err)
			return
		}

		if len(workerKv.Kvs) != 1 {
			t.Error("invalid expected running crawl workers length")
			return
		}
	}
}

func TestCrawlWithUnexpectedShutdown(t *testing.T) {

}
