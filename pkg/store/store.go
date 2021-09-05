package store

import (
	"context"
	"time"

	metav1 "crawlerd/pkg/meta/v1"
)

// TODO: aliases - needed for multi-tenant in same collection

// TODO: build cache layer for fast retrieving
// TODO: don't mix request queue for seed and search engine
type RequestQueue interface {
	List(ctx context.Context, filters *metav1.RequestQueueListFilter) ([]*metav1.RequestQueue, error)

	InsertMany(context.Context, []*metav1.RequestQueueCreate) ([]string, error)

	UpdateByID(context.Context, string, *metav1.RequestQueuePatch) error
}

type Linker interface {
	InsertManyIfNotExists(context.Context, []*metav1.LinkNodeCreate) ([]string, error)

	FindAll(context.Context) ([]*metav1.LinkNode, error)
}

// TODO: better name
// Deprecated: URL is now Linker
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

// Deprecated: Registry is now RequestQueue
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

	Functions() Functions
}

type Runner interface {
	List(context.Context) ([]*metav1.Runner, error)

	GetByID(context.Context, string) (*metav1.Runner, error)

	Create(context.Context, *metav1.RunnerCreate) (string, error)

	UpdateByID(context.Context, string, *metav1.RunnerPatch) error
}

type Functions interface {
	// GetByID getting function by id and return their content
	GetByID(context.Context, string) (string, error)
}

// RunnerFunctions is store repository for persistence operations on JavaScript functions
type RunnerFunctions interface {
	Functions
}

// TODO: different name than Repository?
type Repository interface {
	RequestQueue() RequestQueue
	Linker() Linker
	URL() URL
	History() History
	Registry() Registry
	Job() Job
	Runner() Runner
	RunnerFunctions() RunnerFunctions
}
