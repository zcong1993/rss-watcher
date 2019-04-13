package kv

type Store interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
}
