package store

import (
	"context"
	"time"

	metav1 "crawlerd/pkg/meta/v1"
	runnerstorage "crawlerd/pkg/runner/storage"
)

// TODO: aliases - needed for multi-tenant in same collection

type RequestQueueRepository interface {
	InsertMany(context.Context, []*metav1.RequestQueueCreate) ([]string, error)
}

type LinkerRepository interface {
	InsertManyIfNotExists(context.Context, []*metav1.RequestQueueCreate) ([]string, error)
}

// TODO: better name
// Deprecated: URLRepository is now RequestQueueRepository
type URLRepository interface {
	Scroll(context.Context, func([]metav1.URL)) error

	FindOne(context.Context) (metav1.URL, error)
	FindAll(context.Context) ([]metav1.URL, error)

	InsertOne(ctx context.Context, url string, interval int) (bool, int, error)

	// TODO: replace update interface{} with some metav1 struct
	UpdateOneByID(ctx context.Context, id int, update interface{}) (bool, error)

	DeleteOneByID(ctx context.Context, id int) (bool, error)
}

type HistoryRepository interface {
	FindByID(ctx context.Context, id int) ([]metav1.History, error)

	InsertOne(ctx context.Context, id int, response []byte, duration time.Duration, createdAt time.Time) (bool, int, error)
}

type RegistryRepository interface {
	GetURLByID(context.Context, int) (*metav1.CrawlURL, error)

	PutURL(context.Context, metav1.CrawlURL) error

	DeleteURL(context.Context, metav1.CrawlURL) error
	DeleteURLByID(context.Context, int) error
}

type JobRepository interface {
	FindOneByID(context.Context, string) (*metav1.Job, error)
	FindAll(context.Context) ([]metav1.Job, error)

	InsertOne(context.Context, *metav1.JobCreate) (string, error)

	PatchOneByID(ctx context.Context, id string, job *metav1.JobPatch) error

	Functions() runnerstorage.Functions
}

type Storage interface {
	RequestQueue() RequestQueueRepository
	Linker() LinkerRepository
	URL() URLRepository
	History() HistoryRepository
	Registry() RegistryRepository
	Job() JobRepository
}
