package scheduler

import (
	"context"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/roundrobin"
	"github.com/cenkalti/backoff/v3"
	log "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type leasing struct {
	etcd *clientv3.Client
	srv  Server
}

type Leasing interface {
	Lease() error

	newBackoff(fn func() error) error
}

func NewETCDLeasing(etcd *clientv3.Client, srv Server) Leasing {
	return &leasing{etcd: etcd, srv: srv}
}

func (l *leasing) Lease() error {
	return l.newBackoff(func() error {
		resp, err := l.etcd.Get(context.Background(), "worker.", clientv3.WithPrefix())
		if err != nil {
			return err
		}

		if resp.Kvs == nil || len(resp.Kvs) == 0 {
			l.srv.PutWorkerGen(nil)
			return ErrNoWorkers
		}

		workers := make([]interface{}, len(resp.Kvs))
		for i, kv := range resp.Kvs {
			var grpcconn *grpc.ClientConn

			err := l.newBackoff(func() error {
				ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
				//TODO: insecure
				conn, err := grpc.DialContext(ctx, string(kv.Value), grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					return err
				}
				grpcconn = conn
				return nil
			})

			if err != nil {
				workerKey := string(kv.Key)
				log.Warn("delete worker: ", workerKey)
				if _, err := l.etcd.Delete(context.Background(), workerKey); err != nil {
					return err
				}

				return err
			}

			workers[i] = crawlerdpb.NewWorkerClient(grpcconn)
		}

		robinGen := roundrobin.New(workers)

		l.srv.PutWorkerGen(robinGen)

		return nil
	})
}

func (l *leasing) newBackoff(fn func() error) error {
	maxWait := time.Second * 15

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = maxWait

	return backoff.Retry(fn, bo)
}
