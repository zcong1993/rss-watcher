package kv

import (
	"context"

	"github.com/fauna/faunadb-go/v3/faunadb"
	"github.com/pkg/errors"
	"github.com/zcong1993/rss-watcher/pkg/config"
)

type Fanua struct {
	client     *faunadb.FaunaClient
	collection string
	indexName  string
}

type Record struct {
	Key   string `fauna:"key"`
	Value string `fauna:"value"`
}

func NewFanua(fanuaConfig *config.FaunaConfig) (*Fanua, error) {
	client := faunadb.NewFaunaClient(fanuaConfig.Secret)
	return &Fanua{
		client:     client,
		collection: fanuaConfig.Collection,
		indexName:  fanuaConfig.IndexName,
	}, nil
}

func (f *Fanua) Name() string {
	return "fanua"
}

func (f *Fanua) Get(_ context.Context, key string) (string, error) {
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

func (f *Fanua) Set(_ context.Context, key string, value string) error {
	res, err := f.get(key)
	if errors.Is(err, ErrNotFound) {
		return f.create(key, value)
	}
	return f.update(key, value, res)
}

func (f *Fanua) create(key string, value string) error {
	record := Record{
		Key:   key,
		Value: value,
	}

	_, err := f.client.Query(faunadb.Create(faunadb.Collection(f.collection), faunadb.Obj{
		"data": record,
	}))
	return errors.Wrap(err, "create")
}

func (f *Fanua) update(key string, value string, base faunadb.Value) error {
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

func (f *Fanua) get(key string) (faunadb.Value, error) {
	res, err := f.client.Query(faunadb.Get(faunadb.MatchTerm(faunadb.Index(f.indexName), key)))
	if err != nil {
		if errors.As(err, &faunadb.NotFound{}) {
			return nil, ErrNotFound
		}
		return nil, errors.Wrap(err, "query data")
	}
	return res, err
}

func (f *Fanua) Close() error {
	return nil
}

var _ Store = (*Fanua)(nil)
