package scheduler

import (
	"context"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v3"
	log "github.com/sirupsen/logrus"
	"github.com/zdunecki/crawlerd/crawlerdpb"
	"github.com/zdunecki/crawlerd/pkg/core/scheduler"
	"github.com/zdunecki/crawlerd/pkg/store"
	"github.com/zdunecki/crawlerd/pkg/worker"
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

type schedulerT struct {
	storage       store.Repository
	watcher       Watcher
	leasing       Leasing
	server        Server
	clusterConfig *worker.Config

	log *log.Entry
}

func New(opts ...Option) (Scheduler, error) {
	if os.Getenv("DEBUG") == "1" { // TODO: find better place but init is not the best because it runs before tests and we can't set DEBUG=1 programmatically during tests
		log.SetLevel(log.DebugLevel)
	}

	srv := NewServer()

	s := &schedulerT{
		server: srv,
		log: log.WithFields(map[string]interface{}{
			"service": "scheduler",
		}),
	}

	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}

	if s.storage == nil {
		return nil, scheduler.ErrStorageIsRequired
	}

	if s.watcher == nil {
		return nil, scheduler.ErrWatcherIsRequired
	}

	if s.leasing == nil {
		return nil, scheduler.ErrLeasingIsRequired
	}

	srv.setLasing(s.leasing)

	return s, nil
}

func (s *schedulerT) Serve(addr string) error {
	maxWait := time.Minute

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 5
	bo.MaxElapsedTime = maxWait

	once := &sync.Once{}

	return backoff.Retry(func() error {
		s.log.Debug("lease")
		if err := s.leasing.Lease(); err != nil && err != scheduler.ErrNoWorkers {
			s.log.Debug("lease err: " + err.Error())
			return err
		}

		lis, err := net.Listen("tcp", addr)
		if err != nil {
			s.log.Debug("listen err: " + err.Error())
			return err
		}
		grpcsrv := grpc.NewServer()
		s.log.Info("register grpc server")
		crawlerdpb.RegisterSchedulerServer(grpcsrv, s.server)

		once.Do(func() {
			go s.watchWorkers()
			go s.watchNewURLs()
		})

		s.log.Info("listening on: ", lis.Addr())
		if err := grpcsrv.Serve(lis); err != nil {
			return err
		}

		return nil
	}, bo)
}

func (s *schedulerT) watchWorkers() {
	s.watcher.WatchWorkers(func(ev WorkerWatcherEvent) {
		switch ev {
		case WorkerWatcherEventDelete, WorkerWatcherEventPut, WorkerWatcherEventTicker:
			if err := s.leasing.Lease(); err != nil && err != scheduler.ErrNoWorkers {
				s.log.Error(err)
				return
			}

			if ev == WorkerWatcherEventPut {
				s.watcher.ResetTimer()
			}
		}
	})
}

func (s *schedulerT) watchNewURLs() {
	s.watcher.WatchNewURLs(func(url *crawlerdpb.RequestURL) {
		if _, err := s.server.AddURL(context.Background(), url); err != nil {
			s.log.Error(err)
		}
	})
}
