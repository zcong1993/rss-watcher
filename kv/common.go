package kv

import "errors"

type Store interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
}

// nolint
var ErrNotFound = errors.New("NotFound")
