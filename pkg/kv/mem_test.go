package kv_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zcong1993/rss-watcher/pkg/kv"
)

func TestMemStore_Get(t *testing.T) {
	ms := kv.NewMemStore()
	val, err := ms.Get(context.Background(), "test1")
	assert.Error(t, err)
	_ = ms.Set(context.Background(), "test", "string")
	val, err = ms.Get(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, "string", val)
}

func TestMemStore_Set(t *testing.T) {
	ms := kv.NewMemStore()
	err := ms.Set(context.Background(), "test", "string")
	assert.Nil(t, err)
	val, err := ms.Get(context.Background(), "test")
	assert.Nil(t, err)
	assert.Equal(t, "string", val)
}
