package scheduler

import (
	"context"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/roundrobin"
	"crawlerd/pkg/worker"
	"github.com/cenkalti/backoff/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type leasing struct {
	workerCluster worker.Cluster
	srv           Server
}

type Leasing interface {
	Lease() error

	newBackoff(fn func() error) error
}

func NewLeasing(workerCluster worker.Cluster, srv Server) Leasing {
	return &leasing{workerCluster: workerCluster, srv: srv}
}

func (l *leasing) Lease() error {
	return l.newBackoff(func() error {
		workers, err := l.workerCluster.GetAll(context.Background())
		if err != nil {
			return err
		}

		if workers == nil || len(workers) == 0 {
			l.srv.PutWorkerGen(nil)
			return ErrNoWorkers
		}

		workerClients := make([]interface{}, len(workers))

		for i, w := range workers {
			var grpcconn *grpc.ClientConn

			err := l.newBackoff(func() error {
				ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
				//TODO: insecure

				conn, err := grpc.DialContext(ctx, w.Addr, grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					return err
				}
				grpcconn = conn
				return nil
			})

			if err != nil {
				workerID := w.ID
				log.Warn("delete worker: ", workerID)
				if err := l.workerCluster.DeleteByID(context.Background(), workerID); err != nil {
					return err
				}

				return err
			}

			workerClients[i] = crawlerdpb.NewWorkerClient(grpcconn)
		}

		robinGen := roundrobin.New(workerClients)

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
