package scheduler

import (
	"context"

	"crawlerd/crawlerdpb"
	scheduler "crawlerd/pkg/core/scheduler"
)

type Server interface {
	crawlerdpb.SchedulerServer
	PutWorkerGen(WorkerGen)

	setLasing(Leasing)
	getWorker() (crawlerdpb.WorkerClient, error)
}

type server struct {
	workerGen WorkerGen
	leasing   Leasing
}

func NewServer() Server {
	return &server{}
}

func (s *server) AddURL(ctx context.Context, url *crawlerdpb.RequestURL) (*crawlerdpb.ResponseURL, error) {
	if url.Lease {
		if err := s.leasing.Lease(); err != nil {
			return nil, err
		}
	}

	workerClient, err := s.getWorker()
	if err != nil {
		return nil, err
	}

	resp, err := workerClient.AddURL(ctx, url)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// TODO: in-memory registry may not find worker because workerGen is roundrobin and state of worker's is internal
func (s *server) UpdateURL(ctx context.Context, url *crawlerdpb.RequestURL) (*crawlerdpb.ResponseURL, error) {
	if url.Lease {
		if err := s.leasing.Lease(); err != nil {
			return nil, err
		}
	}

	workerClient, err := s.getWorker()
	if err != nil {
		return nil, err
	}

	resp, err := workerClient.UpdateURL(ctx, url)
	if err != nil {
		return nil, err
	}

	return resp, err
}

// TODO: in-memory registry may not find worker because workerGen is roundrobin and state of worker's is internal
func (s *server) DeleteURL(ctx context.Context, url *crawlerdpb.RequestDeleteURL) (*crawlerdpb.ResponseURL, error) {
	workerClient, err := s.getWorker()
	if err != nil {
		return nil, err
	}

	resp, err := workerClient.DeleteURL(ctx, url)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (s *server) PutWorkerGen(scheduler WorkerGen) {
	s.workerGen = scheduler
}

func (s *server) setLasing(l Leasing) {
	s.leasing = l
}

func (s *server) getWorker() (crawlerdpb.WorkerClient, error) {
	if s.workerGen == nil {
		return nil, scheduler.ErrNoWorkerGen
	}

	workerClient, ok := s.workerGen.Next().(crawlerdpb.WorkerClient)
	if !ok {
		return nil, scheduler.ErrWorkerType
	}

	return workerClient, nil
}
