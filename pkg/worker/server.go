package worker

import (
	"context"

	"crawlerd/crawlerdpb"
	"crawlerd/pkg/storage/objects"
)

type server struct {
	crawler Crawler
}

func NewServer(crawler Crawler) crawlerdpb.WorkerServer {
	return &server{crawler}
}

func (s *server) AddURL(ctx context.Context, req *crawlerdpb.RequestURL) (*crawlerdpb.ResponseURL, error) {
	s.crawler.Enqueue(objects.CrawlURL{
		Id:       req.Id,
		Url:      req.Url,
		Interval: req.Interval,
	})

	return &crawlerdpb.ResponseURL{}, nil
}

func (s *server) UpdateURL(ctx context.Context, req *crawlerdpb.RequestURL) (*crawlerdpb.ResponseURL, error) {
	s.crawler.Update(objects.CrawlURL{
		Id:       req.Id,
		Url:      req.Url,
		Interval: req.Interval,
	})

	return &crawlerdpb.ResponseURL{}, nil
}

func (s *server) DeleteURL(ctx context.Context, deleteURL *crawlerdpb.RequestDeleteURL) (*crawlerdpb.ResponseURL, error) {
	s.crawler.Dequeue(deleteURL.Id)

	return &crawlerdpb.ResponseURL{}, nil
}
