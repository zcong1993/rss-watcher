package fauna

import (
	"context"

	"github.com/fauna/faunadb-go/v3/faunadb"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zcong1993/rss-watcher/pkg/store"
)

const Name = "fauna"

type Store struct {
	client     *faunadb.FaunaClient
	collection string
	indexName  string
}

type Config struct {
	Secret     string `json:"secret" validate:"required"`
	Collection string `json:"collection" validate:"required"`
	IndexName  string `json:"index_name" validate:"required"`
}

type Record struct {
	Key   string `fauna:"key"`
	Value string `fauna:"value"`
}

func NewFanuaStore() store.Store {
	return &Store{}
}

func (f *Store) Init(cfg interface{}) error {
	config, ok := cfg.(*Config)
	if !ok {
		return errors.New("invalid fanua config")
	}

	err := validator.New().Struct(config)
	if err != nil {
		return errors.Wrap(err, "valid config error")
	}

	f.client = faunadb.NewFaunaClient(config.Secret)
	f.collection = config.Collection
	f.indexName = config.IndexName

	return nil
}

func (f *Store) Get(_ context.Context, key string) (string, error) {
	res, err := f.get(key)
	if err != nil {
		return "", err
	}

	var val Record
	if err := res.At(faunadb.ObjKey("data")).Get(&val); err != nil {
		return "", errors.Wrap(err, "decode data")
	}

	return val.Value, nil
}

func (f *Store) Set(_ context.Context, key string, value string) error {
	res, err := f.get(key)
	if errors.Is(err, store.ErrNotFound) {
		return f.create(key, value)
	}

	return f.update(key, value, res)
}

func (f *Store) create(key string, value string) error {
	record := Record{
		Key:   key,
		Value: value,
	}

	_, err := f.client.Query(faunadb.Create(faunadb.Collection(f.collection), faunadb.Obj{
		"data": record,
	}))

	return errors.Wrap(err, "create")
}

func (f *Store) update(key string, value string, base faunadb.Value) error {
	record := Record{
		Key:   key,
		Value: value,
	}

	var refv faunadb.RefV

	err := base.At(faunadb.ObjKey("ref")).Get(&refv)
	if err != nil {
		return errors.Wrap(err, "get ref")
	}

	_, err = f.client.Query(faunadb.Update(refv, faunadb.Obj{
		"data": record,
	}))

	return errors.Wrap(err, "update")
}

func (f *Store) get(key string) (faunadb.Value, error) {
	res, err := f.client.Query(faunadb.Get(faunadb.MatchTerm(faunadb.Index(f.indexName), key)))
	if err != nil {
		if errors.As(err, &faunadb.NotFound{}) {
			return nil, store.ErrNotFound
		}

		return nil, errors.Wrap(err, "query data")
	}

	return res, nil
}

func (f *Store) Close() error {
	return nil
}
