package worker

import (
	"context"

	"crawlerd/crawlerdpb"
)

type server struct {
	crawler Crawler
}

func NewServer(crawler Crawler) crawlerdpb.WorkerServer {
	return &server{crawler}
}

func (s *server) AddURL(ctx context.Context, req *crawlerdpb.RequestURL) (*crawlerdpb.ResponseURL, error) {
	s.crawler.Enqueue(CrawlURL{
		Id:       req.Id,
		Url:      req.Url,
		Interval: req.Interval,
	})

	return &crawlerdpb.ResponseURL{}, nil
}

func (s *server) UpdateURL(ctx context.Context, req *crawlerdpb.RequestURL) (*crawlerdpb.ResponseURL, error) {
	s.crawler.Update(CrawlURL{
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
