package main

import (
	"context"
	"io/ioutil"
	"path"

	"crawlerd/pkg/store"
)

type pluginsRepository struct {
}

func (p *pluginsRepository) GetByID(c context.Context, name string) (string, error) {
	folder := "./pkg/runner"

	if b, err := ioutil.ReadFile(path.Join(folder, name)); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

type fakeStorage struct {
	pluginsRepository *pluginsRepository
}

func newFakeStorage() store.Repository {
	return &fakeStorage{
		pluginsRepository: &pluginsRepository{},
	}
}

func (s *fakeStorage) RequestQueue() store.RequestQueue {
	return nil
}

func (s *fakeStorage) Linker() store.Linker {
	return nil
}

func (s *fakeStorage) URL() store.URL {
	return nil
}

func (s *fakeStorage) History() store.History {
	return nil
}

func (s *fakeStorage) Registry() store.Registry {
	return nil
}

func (s *fakeStorage) Job() store.Job {
	return nil
}

func (s *fakeStorage) Runner() store.Runner {
	return nil
}

func (s *fakeStorage) RunnerFunctions() store.RunnerFunctions {
	return s.pluginsRepository
}
