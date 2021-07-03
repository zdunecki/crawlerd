package scheduler

import "errors"

var (
	ErrNoWorkers   = errors.New("no workers")
	ErrNoWorkerGen = errors.New("no worker generator")
	ErrWorkerType  = errors.New("invalid worker type")

	ErrStorageIsRequired = errors.New("storage is required")
	ErrWatcherIsRequired = errors.New("watcher is required")
	ErrLeasingIsRequired = errors.New("leasing is required")
	ErrClusterIsRequired = errors.New("cluster is required")
)
