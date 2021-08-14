package testkit

import (
	"context"

	"crawlerd/pkg/runner"
	"crawlerd/pkg/store"
)

type runnerStore struct {
	runner    runner.Runner
	functions runner.Functions
}

func newRunnerStore(s store.Repository) runner.Store {
	return &runnerStore{
		runner:    s.Runner(),
		functions: s.Job().Functions(),
	}
}

func (r runnerStore) Runner() runner.Runner {
	return r.runner
}

func (r runnerStore) Functions() runner.Functions {
	return r.functions
}

type GetFn func(context.Context, string) (string, error)

type testFunctions struct {
	getFn GetFn
}

func (f *testFunctions) GetByID(c context.Context, name string) (string, error) {
	return f.getFn(c, name)
}

type testStore struct {
	functions *testFunctions
	runner    runner.Store
}

func NewTestStore(getFn GetFn, store store.Repository) runner.Store {
	return &testStore{
		functions: &testFunctions{
			getFn: getFn,
		},
		runner: newRunnerStore(store),
	}
}

func (s *testStore) Functions() runner.Functions {
	return s.functions
}

func (s *testStore) Runner() runner.Runner {
	return s.runner.Runner()
}
