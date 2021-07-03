package storage

import (
	"context"
	"time"

	"crawlerd/pkg/storage/objects"
)

// TODO: better name
type URLRepository interface {
	Scroll(context.Context, func([]objects.URL)) error

	FindOne(context.Context) (objects.URL, error)
	FindAll(context.Context) ([]objects.URL, error)
	FindAllByWorkerID(context.Context) ([]objects.URL, error)

	InsertOne(ctx context.Context, url string, interval int) (bool, int, error)

	UpdateOneByID(ctx context.Context, id int, update interface{}) (bool, error)

	DeleteOneByID(ctx context.Context, id int) (bool, error)
}

type HistoryRepository interface {
	FindByID(ctx context.Context, id int) ([]objects.History, error)
	InsertOne(ctx context.Context, id int, response []byte, duration time.Duration, createdAt time.Time) (bool, int, error)
}

type RegistryRepository interface {
	GetURLByID(int) (*objects.CrawlURL, error)
	PutURL(objects.CrawlURL) error
	DeleteURL(objects.CrawlURL) error
	FindURLByWorkerID(string) ([]objects.CrawlURL, error)
	DeleteURLByID(int) error

	//crawlID(int) string
}

type Storage interface {
	URL() URLRepository
	History() HistoryRepository
	Registry() RegistryRepository
}

type storage struct {
	url      URLRepository
	history  HistoryRepository
	registry RegistryRepository
}

func (s *storage) URL() URLRepository {
	return s.url
}

func (s *storage) History() HistoryRepository {
	return s.history
}

func (s *storage) Registry() RegistryRepository {
	return s.registry
}

type Option func(*storage)

func WithURL(r URLRepository) Option {
	return func(s *storage) {
		s.url = r
	}
}

func WithHistory(r HistoryRepository) Option {
	return func(s *storage) {
		s.history = r
	}
}

func WithRegistry(r RegistryRepository) Option {
	return func(s *storage) {
		s.registry = r
	}
}

func NewStorage(opts ...Option) Storage {
	s := &storage{}

	for _, o := range opts {
		o(s)
	}

	return s
}
