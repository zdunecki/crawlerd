package scheduler

import (
	"context"

	"crawlerd/crawlerdpb"
)

type Server interface {
	crawlerdpb.SchedulerServer
	PutWorkerGen(WorkerGen)

	getWorker() (crawlerdpb.WorkerClient, error)
}

type server struct {
	workerGen WorkerGen
}

func NewServer() Server {
	return &server{}
}

func (s *server) AddURL(ctx context.Context, url *crawlerdpb.RequestURL) (*crawlerdpb.ResponseURL, error) {
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

func (s *server) UpdateURL(ctx context.Context, url *crawlerdpb.RequestURL) (*crawlerdpb.ResponseURL, error) {
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

func (s *server) getWorker() (crawlerdpb.WorkerClient, error) {
	if s.workerGen == nil {
		return nil, ErrNoWorkerGen
	}

	workerClient, ok := s.workerGen.Next().(crawlerdpb.WorkerClient)
	if !ok {
		return nil, ErrWorkerType
	}

	return workerClient, nil
}
