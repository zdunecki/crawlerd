package runner

import (
	"context"

	metav1 "crawlerd/pkg/meta/v1"
)

type Runner interface {
	List(context.Context) ([]*metav1.Runner, error)

	GetByID(context.Context, string) (*metav1.Runner, error)

	Create(context.Context, *metav1.RunnerCreate) (string, error)

	UpdateByID(context.Context, string, *metav1.RunnerPatch) error
}

// Functions is store repository for persistence operations on JavaScript functions
type Functions interface {
	// GetByID getting function by id and return their content
	GetByID(context.Context, string) (string, error)
}

type Store interface {
	Runner() Runner
	Functions() Functions
}
