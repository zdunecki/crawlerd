package v1

import "errors"

var (
	ErrNoStorage   = errors.New("no store")
	ErrNoScheduler = errors.New("no scheduler")
)
