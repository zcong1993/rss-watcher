package kv

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/mmcdole/gofeed"

	"github.com/jinzhu/copier"
)

type FileStore struct {
	p     string
	store map[string]gofeed.Item
}

func NewFileStore(p string) *FileStore {
	store := make(map[string]gofeed.Item)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		f, _ := os.Create(p)
		_, _ = f.Write([]byte("{}"))
		_ = f.Close()
	} else {
		data, err := ioutil.ReadFile(p)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(data, &store)
		if err != nil {
			panic(err)
		}
	}

	return &FileStore{
		p:     p,
		store: store,
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
	// value is *gofeed.Item type
	fs.store[key] = *(value.(*gofeed.Item))
	storeBt, err := json.Marshal(fs.store)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fs.p, storeBt, 0644)
}
