package main

import (
	"context"
	"io/ioutil"
	"path"

	"crawlerd/pkg/runner/storage"
)

type pluginsRepository struct {
}

func (p *pluginsRepository) Get(c context.Context, name string) (string, error) {
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

func newFakeStorage() storage.Storage {
	return &fakeStorage{
		pluginsRepository: &pluginsRepository{},
	}
}

func (s *fakeStorage) Functions() storage.Functions {
	return s.pluginsRepository
}
