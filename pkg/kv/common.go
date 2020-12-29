package kv

import "github.com/pkg/errors"

type Store interface {
	Get(key string) (string, error)
	Set(key string, value string) error
	Close() error
}

// nolint
var ErrNotFound = errors.New("NotFound")
