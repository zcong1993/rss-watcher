package kv_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zcong1993/rss-watcher/pkg/kv"
)

func TestMemStore_Get(t *testing.T) {
	ms := kv.NewMemStore()
	err := ms.Get("test1", nil)
	assert.Error(t, err)
	_ = ms.Set("test", "string")
	var val string
	err = ms.Get("test", &val)
	assert.Nil(t, err)
	assert.Equal(t, "string", val)
}

func TestMemStore_Set(t *testing.T) {
	ms := kv.NewMemStore()
	err := ms.Set("test", "string")
	assert.Nil(t, err)
	var val string
	err = ms.Get("test", &val)
	assert.Nil(t, err)
	assert.Equal(t, "string", val)
}

func TestMemStore_Complex(t *testing.T) {
	ms := kv.NewMemStore()
	type person struct {
		Name string
		Age  int
	}
	p := person{
		Name: "test",
		Age:  18,
	}
	err := ms.Set("p", p)
	assert.Nil(t, err)
	var pp person
	err = ms.Get("p", &pp)
	assert.Nil(t, err)
	assert.Equal(t, p, pp)
}
