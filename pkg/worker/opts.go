package worker

import (
	"bytes"
	"net/http"
	"time"

	"crawlerd/pkg/pubsub"
	"crawlerd/pkg/store/options"
	"github.com/andybalholm/brotli"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/client-go/kubernetes"
)

type (
	Option     func(*worker) error
	Compressor func([]byte) ([]byte, error)
)

var (
	DefaultETCDConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Second * 15,
	}
)

func WithSchedulerGRPCAddr(addr string) Option {
	return func(w *worker) error {
		w.schedulerAddr = addr
		return nil
	}
}

func WithPubSub(pubsub pubsub.PubSub) Option {
	return func(w *worker) error {
		w.pubsub = pubsub
		return nil
	}
}

func WithCompression(compressor Compressor) Option {
	return func(w *worker) error {
		w.compressor = compressor
		return nil
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(w *worker) error {
		w.httpClient = client
		return nil
	}
}

func WithBrotliCompression() Option { // TODO:  cbrotli vs brotli
	return func(w *worker) error {
		return WithCompression(func(d []byte) ([]byte, error) {
			buff := new(bytes.Buffer)
			compressed := brotli.NewWriter(buff)

			if _, err := compressed.Write(d); err != nil {
				return nil, err
			}

			if err := compressed.Close(); err != nil {
				return nil, err
			}

			return buff.Bytes(), nil
		})(w)
	}
}

func WithETCDCluster(optCfg ...clientv3.Config) Option {
	return func(w *worker) error {
		cfg := DefaultETCDConfig

		for _, c := range optCfg {
			cfg = c
		}

		etcd, err := clientv3.New(cfg)

		if err != nil {
			return err
		}

		w.cluster = NewETCDCluster(etcd)

		return nil
	}
}

func WithK8sCluster(client kubernetes.Interface, namespace string, workerSelector string) Option {
	return func(w *worker) error {
		w.cluster = NewK8sCluster(client, namespace, workerSelector, w.config)

		return nil
	}
}

func WithStorage(opts ...*options.RepositoryOption) Option {
	return func(w *worker) error {
		s, err := options.WithStorage(opts...)

		if err != nil {
			return err
		}

		w.storage = s

		return nil
	}
}
