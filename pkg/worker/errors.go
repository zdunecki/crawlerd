package worker

import "errors"

var (
	ErrHTTPNotOK            = errors.New("invalid status is not ok")
	ErrHTTPMaxContentLength = errors.New("max content-length exceeded")

	ErrStorageIsRequired  = errors.New("storage is required")
	ErrRegistryIsRequired = errors.New("registry is required")
	ErrPubSubIsRequired   = errors.New("pubsub is required")

	ErrEmptySchedulerGRPCSrvAddr = errors.New("empty scheduler grpc server address")

	ErrWorkerIsRequired = errors.New("worker is required")
)
