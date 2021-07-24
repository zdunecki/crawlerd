package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	crawldv1 "crawlerd/api/v1"
	clientv3 "go.etcd.io/etcd/client/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: is it integration test?

func TestK8s(t *testing.T) {
	namespace := "test"
	k8sWorkerSelector := "app=test-worker"
	k8sHost := "test-k8s-host"
	k8sPodIP := "127.0.0.1" // TODO: tests different ips - needs fake network

	os.Setenv("DEBUG", "1")
	os.Setenv("WORKER_HOST", k8sHost)
	os.Setenv("WORKER_GRPC_ADDR", "9115")

	setup, err := setupK8sClient(
		namespace,
		k8sWorkerSelector,
		&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      k8sHost,
				Labels: map[string]string{
					"app": k8sWorkerSelector,
				},
			},
			Spec: v1.PodSpec{
				Hostname: k8sHost,
			},
			Status: v1.PodStatus{
				PodIP: k8sPodIP,
			},
		},
	)

	if err != nil {
		t.Error(err)
		return
	}

	defer setup.done()
	crawldURL := setup.crawld.URL()

	setup.etcdContainer.DefaultAddress()
	createResp, err := crawldURL.Create(&crawldv1.RequestPostURL{
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

	if len(workerKv.Kvs) != 0 {
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
