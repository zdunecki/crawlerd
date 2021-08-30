package testkit

import (
	"context"

	"crawlerd/pkg/store"
)

type GetFn func(context.Context, string) (string, error)

type testRunnerFunctions struct {
	getFn GetFn
}

func NewTestRunnerFunctions(getFn GetFn) store.RunnerFunctions {
	return &testRunnerFunctions{
		getFn: getFn,
	}
}

func (t testRunnerFunctions) GetByID(c context.Context, name string) (string, error) {
	return t.getFn(c, name)
}
