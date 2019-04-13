package kv

import (
	"errors"

	"github.com/jinzhu/copier"
)

var ErrNotFound = errors.New("NOT_FOUND")

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
