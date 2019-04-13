package kv

import (
	"github.com/jinzhu/copier"
)

type MemStore struct {
	store map[string]interface{}
}

func NewMemStore() *MemStore {
	return &MemStore{store: make(map[string]interface{})}
}

func (ms *MemStore) Get(key string, value interface{}) error {
	v, ok := ms.store[key]
	if !ok {
		return ErrNotFound
	}
	return copier.Copy(value, v)
}

func (ms *MemStore) Set(key string, value interface{}) error {
	ms.store[key] = value
	return nil
}
