package kv

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/jinzhu/copier"
)

type FileStore struct {
	p     string
	store map[string]interface{}
}

func NewFileStore(p string) *FileStore {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		f, _ := os.Create(p)
		_, _ = f.Write([]byte("{}"))
		_ = f.Close()
	}

	return &FileStore{
		p:     p,
		store: make(map[string]interface{}),
	}
}

func (fs *FileStore) Get(key string, value interface{}) error {
	v, ok := fs.store[key]
	if !ok {
		return ErrNotFound
	}
	return copier.Copy(value, v)
}

func (fs *FileStore) Set(key string, value interface{}) error {
	fs.store[key] = value
	storeBt, err := json.Marshal(fs.store)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fs.p, storeBt, 0644)
}
