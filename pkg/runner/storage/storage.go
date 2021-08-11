package storage

import "context"

type Functions interface {
	Get(context.Context, string) (string, error)
}

type Storage interface {
	Functions() Functions
}
