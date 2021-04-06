package worker

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/scheduler"
	"crawlerd/pkg/storage"
	"crawlerd/pkg/util"

	"github.com/cenkalti/backoff/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Worker interface {
	ID() string
	Addr() string
	Serve(ctx context.Context) error

	gracefulShutdown()
	addrGen() (workerID, workerAddr string, err error)
	newGRPC() (*grpc.Server, crawlerdpb.SchedulerClient, net.Listener, error)
}

type worker struct {
	id            string
	addr          string
	schedulerAddr string

	storage  storage.Client
	registry Registry
	crawler  Crawler
	ctrl     Controller

	grpcserver *grpc.Server
	listener   net.Listener
}

func New(opts ...Option) (Worker, error) {
	worker := &worker{
		schedulerAddr: scheduler.DefaultSchedulerGRPCServerAddr,
	}

	for _, o := range opts {
		if err := o(worker); err != nil {
			return nil, err
		}
	}

	if worker.storage == nil {
		return nil, ErrStorageIsRequired
	}

	if worker.registry == nil {
		return nil, ErrRegistryIsRequired
	}

	{
		id, addr, err := worker.addrGen()
		if err != nil {
			return nil, err
		}

		worker.id = id
		worker.addr = addr

		worker.crawler = NewCrawler(worker.storage, worker.registry, worker)

		grpcsrv, schedulercli, lis, err := worker.newGRPC()
		if err != nil {
			return nil, err
		}

		worker.ctrl = NewController(schedulercli, worker.registry)

		worker.grpcserver = grpcsrv
		worker.listener = lis

		worker.registry.WithWorker(worker)
	}

	return worker, nil
}

func (c *worker) ID() string {
	return c.id
}

func (c *worker) Addr() string {
	return c.addr
}

func (c *worker) Serve(ctx context.Context) error {
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
				c.listener.Close()
			}
		}
	}()

	return backoff.Retry(func() error {
		ctx := ctx
		if err := c.registry.RegisterWorker(); err != nil {
			return err
		}

		once.Do(func() {
			go c.gracefulShutdown() // TODO: return channel
		})

		log.Info("listening on: ", c.listener.Addr())
		err := c.grpcserver.Serve(c.listener)

		if err != nil && ctx.Err() != context.Canceled {
			return err
		}

		return nil
	}, bo)
}

func (c *worker) gracefulShutdown() {
	sigint := make(chan os.Signal, 1)

	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-sigint

	{
		if err := c.registry.UnregisterWorker(); err != nil {
			log.Error(err)
			return
		}

		c.crawler.Stop(c.ctrl.ReAttachResources)
	}

	os.Exit(0)
}

func (c *worker) addrGen() (workerID, workerAddr string, err error) {
	workerID = util.RandomString(10)
	// TODO: attach another port if already exists
	workerPort := strconv.Itoa(util.Between(9111, 9555))
	workerHost, err := os.Hostname()
	hostEnv := os.Getenv("HOST")
	if hostEnv != "" {
		workerHost = hostEnv
	} else {
		if err != nil {
			return "", "", err
		}
	}

	workerAddr = net.JoinHostPort(workerHost, workerPort)

	return workerID, workerAddr, nil
}

func (c *worker) newGRPC() (*grpc.Server, crawlerdpb.SchedulerClient, net.Listener, error) {
	if c.addr == "" {
		return nil, nil, nil, ErrEmptySchedulerGRPCSrvAddr
	}
	lis, err := net.Listen("tcp", c.addr)
	grpcsrv := grpc.NewServer()

	crawlerdpb.RegisterWorkerServer(grpcsrv, NewServer(c.crawler))

	schedulerconn, err := grpc.Dial(c.schedulerAddr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, err
	}

	return grpcsrv, crawlerdpb.NewSchedulerClient(schedulerconn), lis, nil
}
