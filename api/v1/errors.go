package v1

import "errors"

var (
	ErrNoStorage   = errors.New("no storage")
	ErrNoScheduler = errors.New("no scheduler")
)
