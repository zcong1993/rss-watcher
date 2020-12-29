package kv

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type FileStore struct {
	p     string
	store map[string]string
}

func NewFileStore(p string) *FileStore {
	store := make(map[string]string)
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

func (fs *FileStore) Get(key string) (string, error) {
	v, ok := fs.store[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (fs *FileStore) Set(key string, value string) error {
	fs.store[key] = value
	return fs.save()
}

func (fs *FileStore) save() error {
	tmpF := fs.p + ".tmp"
	storeBt, err := json.Marshal(fs.store)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(tmpF, storeBt, 0644)
	if err != nil {
		return err
	}
	return os.Rename(tmpF, fs.p)
}

func (fs *FileStore) Close() error {
	return nil
}
