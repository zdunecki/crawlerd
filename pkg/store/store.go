package store

import (
	"context"
	"time"

	metav1 "crawlerd/pkg/meta/v1"
	"crawlerd/pkg/runner"
)

// TODO: aliases - needed for multi-tenant in same collection

type RequestQueue interface {
	List(ctx context.Context, options *metav1.RequestQueueListQuery)

	InsertMany(context.Context, []*metav1.RequestQueueCreate) ([]string, error)
}

type Linker interface {
	InsertManyIfNotExists(context.Context, []*metav1.LinkNodeCreate) ([]string, error)

	FindAll(context.Context) ([]*metav1.LinkNode, error)
}

// TODO: better name
// Deprecated: URL is now RequestQueue
type URL interface {
	Scroll(context.Context, func([]metav1.URL)) error

	FindOne(context.Context) (metav1.URL, error)
	FindAll(context.Context) ([]metav1.URL, error)

	InsertOne(ctx context.Context, url string, interval int) (bool, int, error)

	// TODO: replace update interface{} with some metav1 struct
	UpdateOneByID(ctx context.Context, id int, update interface{}) (bool, error)

	DeleteOneByID(ctx context.Context, id int) (bool, error)
}

type History interface {
	FindByID(ctx context.Context, id int) ([]metav1.History, error)

	InsertOne(ctx context.Context, id int, response []byte, duration time.Duration, createdAt time.Time) (bool, int, error)
}

type Registry interface {
	GetURLByID(context.Context, int) (*metav1.CrawlURL, error)

	PutURL(context.Context, metav1.CrawlURL) error

	DeleteURL(context.Context, metav1.CrawlURL) error
	DeleteURLByID(context.Context, int) error
}

type Job interface {
	FindOneByID(context.Context, string) (*metav1.Job, error)
	FindAll(context.Context) ([]metav1.Job, error)

	InsertOne(context.Context, *metav1.JobCreate) (string, error)

	PatchOneByID(ctx context.Context, id string, job *metav1.JobPatch) error

	Functions() runner.Functions
}

type Repository interface {
	RequestQueue() RequestQueue
	Linker() Linker
	URL() URL
	History() History
	Registry() Registry
	Job() Job
	Runner() runner.Runner
}
