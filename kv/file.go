package kv

import (
	"encoding/json"
	"os"
)

type FileStore struct {
	file  *os.File
	store map[string]string
}

func NewFileStore(p string) *FileStore {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		f, _ := os.Create(p)
		return &FileStore{
			file:  f,
			store: make(map[string]string),
		}
	}

	f, _ := os.Open(p)
	fileinfo, err := f.Stat()
	if err != nil {
		panic(err)
	}

	filesize := fileinfo.Size()
	buffer := make([]byte, filesize)

	_, err = f.Read(buffer)
	if err != nil {
		panic(err)
	}

	var store map[string]string
	err = json.Unmarshal(buffer, &store)
	if err != nil {
		panic(err)
	}
	return &FileStore{
		file:  f,
		store: store,
	}
}

func (fs *FileStore) Get(key string, value interface{}) error {
	v, ok := fs.store[key]
	if !ok {
		return ErrNotFound
	}
	return json.Unmarshal([]byte(v), value)
}

func (fs *FileStore) Set(key string, value interface{}) error {
	bt, err := json.Marshal(value)
	if err != nil {
		return err
	}
	fs.store[key] = string(bt)
	storeBt, err := json.Marshal(fs.store)
	if err != nil {
		return err
	}

	_, err = fs.file.Write(storeBt)
	return err
}
