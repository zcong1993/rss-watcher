package file

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zcong1993/rss-watcher/pkg/store"
)

const Name = "file"

type Store struct {
	p     string
	store map[string]string
}

type Config struct {
	Path string `json:"path" validate:"required"`
}

func NewFileStore() *Store {
	return &Store{}
}

func (fs *Store) Init(cfg interface{}) error {
	config, ok := cfg.(*Config)
	if !ok {
		return errors.New("invalid fs config")
	}

	err := validator.New().Struct(config)
	if err != nil {
		return errors.Wrap(err, "valid config error")
	}

	p := config.Path
	fs.p = p
	fs.store = make(map[string]string)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		f, _ := os.Create(p)
		_, _ = f.Write([]byte("{}"))
		_ = f.Close()
	} else {
		data, err := ioutil.ReadFile(p)
		if err != nil {
			return errors.Wrap(err, "read file")
		}
		err = json.Unmarshal(data, &fs.store)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *Store) Get(_ context.Context, key string) (string, error) {
	v, ok := fs.store[key]
	if !ok {
		return "", store.ErrNotFound
	}

	return v, nil
}

func (fs *Store) Set(_ context.Context, key string, value string) error {
	fs.store[key] = value

	return fs.save()
}

func (fs *Store) save() error {
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

func (fs *Store) Close() error {
	return nil
}

var _ store.Store = (*Store)(nil)
