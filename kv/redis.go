package kv

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/go-redis/redis"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(uri string) *RedisStore {
	opts, err := redis.ParseURL(uri)
	if err != nil {
		log.Fatalf("redis init error, %+v\n", err)
	}
	client := redis.NewClient(opts)

	return &RedisStore{client: client}
}

func (rs *RedisStore) Get(key string, value interface{}) error {
	bt, err := rs.client.Get(key).Bytes()
	if err != nil {
		if err.Error() == "redis: nil" {
			return ErrNotFound
		}
		return err
	}
	return json.NewDecoder(bytes.NewBuffer(bt)).Decode(&value)
}

func (rs *RedisStore) Set(key string, value interface{}) error {
	bt, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rs.client.Set(key, bt, 0).Err()
}
