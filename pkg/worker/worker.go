package worker

import (
	"context"
	"net"
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
	Serve(ctx context.Context) error

	gracefulShutdown()
	newGRPC() (*grpc.Server, crawlerdpb.SchedulerClient, net.Listener, error)
}

type worker struct {
	id            string
	addr          string
	schedulerAddr string

	storage storage.Storage
	crawler Crawler
	ctrl    Controller
	pubsub  pubsub.PubSub

	grpcserver *grpc.Server
	listener   net.Listener

	cluster Cluster

	log *log.Entry

	config *Config
}

//func init() {
//	if os.Getenv("DEBUG") == "1" { /
//		log.SetLevel(log.DebugLevel)
//	}
//}

func New(opts ...Option) (Worker, error) {
	if os.Getenv("DEBUG") == "1" { // TODO: find better place but init is not the best because it runs before tests and we can't set DEBUG=1 programmatically during tests
		log.SetLevel(log.DebugLevel)
	}

	worker := &worker{
		schedulerAddr: ":9888",
		log: log.WithFields(map[string]interface{}{
			"service": "worker",
		}),
		config: InitConfig(),
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

	if worker.pubsub == nil {
		return nil, ErrPubSubIsRequired
	}

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

		worker.crawler = NewCrawler(worker.storage, worker, worker.pubsub)

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

	ctx.Done()
	maxWait := time.Minute

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 5
	bo.MaxElapsedTime = maxWait

	once := &sync.Once{}

	go func() {
		for {
			select {
			case <-ctx.Done():
				w.listener.Close()
			}
		}
	}()

	return backoff.Retry(func() error {
		ctx := ctx

		if err := w.cluster.Register(context.Background(), w); err != nil {
			return err
		}

		once.Do(func() {
			go w.gracefulShutdown() // TODO: return channel
		})

		w.log.Info("listening on: ", w.listener.Addr())
		err := w.grpcserver.Serve(w.listener)

		if err != nil && ctx.Err() != context.Canceled {
			return err
		}

		return nil
	}, bo)
}

func (w *worker) gracefulShutdown() {
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-sigint

	{

		if err := w.cluster.Unregister(context.Background(), w); err != nil {
			w.log.Error(err)
			return
		}

		w.crawler.Stop(w.ctrl.ReAttachResources)
	}

	os.Exit(0)
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
