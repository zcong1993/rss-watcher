package kv

import (
	"context"

	"github.com/pkg/errors"
)

type Store interface {
	Name() string
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	Close() error
}

// nolint
var ErrNotFound = errors.New("NotFound")
