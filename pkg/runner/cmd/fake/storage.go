package main

import (
	"context"
	"io/ioutil"
	"path"

	"crawlerd/pkg/runner"
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

func newFakeStorage() runner.Store {
	return &fakeStorage{
		pluginsRepository: &pluginsRepository{},
	}
}

func (s *fakeStorage) Functions() runner.Functions {
	return s.pluginsRepository
}

func (s *fakeStorage) Runner() runner.Runner {
	panic("implement me")
}
