package kv_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zcong1993/rss-watcher/pkg/kv"
)

func TestMemStore_Get(t *testing.T) {
	ms := kv.NewMemStore()
	val, err := ms.Get("test1")
	assert.Error(t, err)
	_ = ms.Set("test", "string")
	val, err = ms.Get("test")
	assert.Nil(t, err)
	assert.Equal(t, "string", val)
}

func TestMemStore_Set(t *testing.T) {
	ms := kv.NewMemStore()
	err := ms.Set("test", "string")
	assert.Nil(t, err)
	val, err := ms.Get("test")
	assert.Nil(t, err)
	assert.Equal(t, "string", val)
}
