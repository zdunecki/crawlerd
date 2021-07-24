package worker

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/pubsub"
	"crawlerd/pkg/storage"
	"github.com/cenkalti/backoff/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Worker interface {
	ID() string
	Addr() string
	Serve(context.Context) error

	gracefulShutdown(context.Context)
	newGRPC() (*grpc.Server, crawlerdpb.SchedulerClient, net.Listener, error)
}

type worker struct {
	id            string
	addr          string
	schedulerAddr string

	httpClient *http.Client
	storage    storage.Storage
	crawler    Crawler
	ctrl       Controller
	pubsub     pubsub.PubSub
	compressor Compressor

	grpcserver *grpc.Server
	listener   net.Listener

	cluster Cluster

	log *log.Entry

	config *Config

	finishedC chan bool
}

//func init() {
//	if os.Getenv("DEBUG") == "1" { /
//		log.SetLevel(log.DebugLevel)
//	}
//}

func New(cfg *Config, opts ...Option) (Worker, error) {
	if os.Getenv("DEBUG") == "1" { // TODO: find better place but init is not the best because it runs before tests and we can't set DEBUG=1 programmatically during tests
		log.SetLevel(log.DebugLevel)
	}

	worker := &worker{
		schedulerAddr: cfg.SchedulerGRPCAddr,
		log: log.WithFields(map[string]interface{}{
			"service": "worker",
		}),
		config:    cfg,
		finishedC: make(chan bool),
	}

	if worker.schedulerAddr == "" {
		// TODO: import cycle not allowed
		// scheduler.DefaultSchedulerGRPCServerAddr
		worker.schedulerAddr = ":9888"
	}

	for _, o := range opts {
		if err := o(worker); err != nil {
			return nil, err
		}
	}

	if worker.storage == nil {
		return nil, ErrStorageIsRequired
	}

	if worker.storage.Registry() == nil {
		return nil, ErrRegistryIsRequired
	}

	// TODO: for k8s testing it's discarded
	//if worker.pubsub == nil {
	//	return nil, ErrPubSubIsRequired
	//}

	if worker.cluster == nil {
		return nil, ErrWorkerIsRequired
	}

	{
		id, addr, err := worker.cluster.WorkerAddr()
		if err != nil {
			return nil, err
		}

		worker.id = id
		worker.addr = addr

		worker.crawler = NewCrawler(worker.storage, worker, worker.pubsub, worker.compressor, worker.httpClient)

		grpcsrv, schedulercli, lis, err := worker.newGRPC()
		if err != nil {
			return nil, err
		}

		worker.ctrl = NewController(schedulercli, worker.storage.Registry())

		worker.grpcserver = grpcsrv
		worker.listener = lis
	}

	return worker, nil
}

func (w *worker) ID() string {
	return w.id
}

func (w *worker) Addr() string {
	return w.addr
}

func (w *worker) Serve(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	//ctx.Done()
	maxWait := time.Minute

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 5
	bo.MaxElapsedTime = maxWait

	once := &sync.Once{}

	once.Do(func() {
		go w.gracefulShutdown(ctx)
	})

	err := backoff.Retry(func() error {
		if err := w.cluster.Register(context.Background(), w); err != nil {
			return err
		}

		w.log.Info("listening on: ", w.listener.Addr())
		err := w.grpcserver.Serve(w.listener)

		if err != nil && ctx.Err() != context.Canceled {
			return err
		}

		return nil
	}, bo)

	<-w.finishedC

	w.log.Debug("successfully gracefully shut down")

	return err
}

func (w *worker) gracefulShutdown(ctx context.Context) {
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	unregister := func() {
		w.log.Debug("gracefully shutting down...")

		if err := w.cluster.DeleteByID(context.Background(), w.ID()); err != nil {
			w.log.Error(err)
			return
		}

		w.crawler.Stop(w.ctrl.ReAttachResources)
	}

	// TODO: need to support this case - grpc server failed but context and sigint didn't call
	for {
		select {
		case <-sigint:
			w.grpcserver.GracefulStop()
			unregister()
			w.finishedC <- true
			return
		case <-ctx.Done():
			w.grpcserver.GracefulStop()
			unregister()
			w.finishedC <- true
			return
		}
	}

	//os.Exit(0)
}

func (w *worker) newGRPC() (*grpc.Server, crawlerdpb.SchedulerClient, net.Listener, error) {
	if w.addr == "" {
		return nil, nil, nil, ErrEmptySchedulerGRPCSrvAddr
	}
	lis, err := net.Listen("tcp", w.addr)
	if err != nil {
		return nil, nil, nil, err
	}
	grpcsrv := grpc.NewServer()

	crawlerdpb.RegisterWorkerServer(grpcsrv, NewServer(w.crawler))

	schedulerconn, err := grpc.Dial(w.schedulerAddr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, err
	}

	return grpcsrv, crawlerdpb.NewSchedulerClient(schedulerconn), lis, nil
}
