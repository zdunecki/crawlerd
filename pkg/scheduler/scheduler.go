package scheduler

import (
	"context"
	"net"
	"sync"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/storage"
	"github.com/cenkalti/backoff/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	DefaultSchedulerGRPCServerAddr = ":9888"
)

type Scheduler interface {
	Serve(string) error

	watchWorkers()
	watchNewURLs()
}

type scheduler struct {
	storage storage.Client
	watcher Watcher
	leasing Leasing
	server  Server
}

func New(opts ...Option) (Scheduler, error) {
	srv := NewServer()

	s := &scheduler{
		server: srv,
	}

	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}

	if s.storage == nil {
		return nil, ErrStorageIsRequired
	}

	if s.watcher == nil {
		return nil, ErrWatcherIsRequired
	}

	if s.leasing == nil {
		return nil, ErrLeasingIsRequired
	}

	return s, nil
}

func (s *scheduler) Serve(addr string) error {
	maxWait := time.Minute

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 5
	bo.MaxElapsedTime = maxWait

	once := &sync.Once{}

	return backoff.Retry(func() error {
		if err := s.leasing.Lease(); err != nil && err != ErrNoWorkers {
			return err
		}

		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		grpcsrv := grpc.NewServer()
		crawlerdpb.RegisterSchedulerServer(grpcsrv, s.server)

		once.Do(func() {
			go s.watchWorkers()
			go s.watchNewURLs()
		})

		log.Info("listening on: ", lis.Addr())
		if err := grpcsrv.Serve(lis); err != nil {
			return err
		}

		return nil
	}, bo)
}

func (s *scheduler) watchWorkers() {
	s.watcher.WatchWorkers(func(ev WorkerWatcherEvent) {
		switch ev {
		case WorkerWatcherEventDelete, WorkerWatcherEventPut, WorkerWatcherEventTicker:
			if err := s.leasing.Lease(); err != nil && err != ErrNoWorkers {
				log.Error(err)
				return
			}

			if ev == WorkerWatcherEventPut {
				s.watcher.ResetTimer()
			}
		}
	})
}

func (s *scheduler) watchNewURLs() {
	s.watcher.WatchNewURLs(func(url *crawlerdpb.RequestURL) {
		if _, err := s.server.AddURL(context.Background(), url); err != nil {
			log.Error(err)
		}
	})
}
