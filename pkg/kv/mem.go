package kv

import "context"

type MemStore struct {
	store map[string]string
}

func NewMemStore() *MemStore {
	return &MemStore{store: make(map[string]string)}
}

func (ms *MemStore) Get(_ context.Context, key string) (string, error) {
	v, ok := ms.store[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (ms *MemStore) Set(_ context.Context, key string, value string) error {
	ms.store[key] = value
	return nil
}

func (ms *MemStore) Close() error {
	return nil
}

func (ms *MemStore) Name() string {
	return "mem"
}

var _ Store = (*MemStore)(nil)
