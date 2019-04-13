package kv

import "errors"

type Store interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
}

var ErrNotFound = errors.New("NOT_FOUND")
