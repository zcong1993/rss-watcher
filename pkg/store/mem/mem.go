package mem

import (
	"context"

	"github.com/zcong1993/rss-watcher/pkg/store"
)

const Name = "mem"

type Store struct {
	store map[string]string
}

func NewMemStore() store.Store {
	return &Store{store: make(map[string]string)}
}

func (ms *Store) Init(cfg interface{}) error {
	return nil
}

func (ms *Store) Get(_ context.Context, key string) (string, error) {
	v, ok := ms.store[key]
	if !ok {
		return "", store.ErrNotFound
	}

	return v, nil
}

func (ms *Store) Set(_ context.Context, key string, value string) error {
	ms.store[key] = value

	return nil
}

func (ms *Store) Close() error {
	return nil
}
