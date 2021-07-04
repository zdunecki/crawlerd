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
	GetURLByID(context.Context, int) (*objects.CrawlURL, error)
	PutURL(context.Context, objects.CrawlURL) error
	DeleteURL(context.Context, objects.CrawlURL) error
	FindURLByWorkerID(context.Context, string) ([]objects.CrawlURL, error)
	DeleteURLByID(context.Context, int) error
}

type Storage interface {
	URL() URLRepository
	History() HistoryRepository
	Registry() RegistryRepository
}
