package storage

import (
	"context"
	"time"

	"crawlerd/pkg/storage/objects"
)

type RepositoryURL interface {
	Scroll(context.Context, func([]objects.URL)) error
	FindAll(context.Context) ([]objects.URL, error)
	InsertOne(ctx context.Context, url string, interval int) (bool, int, error)
	UpdateOneByID(ctx context.Context, id int, update interface{}) (bool, error)
	DeleteOneByID(ctx context.Context, id int) (bool, error)
}

type RepositoryHistory interface {
	FindByID(ctx context.Context, id int) ([]objects.History, error)
	InsertOne(ctx context.Context, id int, response []byte, duration time.Duration, createdAt time.Time) (bool, int, error)
}

type Client interface {
	URL() RepositoryURL
	History() RepositoryHistory
}
