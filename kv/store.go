package kv

import (
	"context"

	"cloud.google.com/go/firestore"
)

type Store interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
}

type FireStore struct {
	ProjectId     string
	Collection    string
	CollectionRef *firestore.CollectionRef
}

func NewFireStore(projectId, collection string) *FireStore {
	client, err := firestore.NewClient(context.Background(), projectId)
	if err != nil {
		panic(err)
	}
	return &FireStore{
		ProjectId:     projectId,
		Collection:    collection,
		CollectionRef: client.Collection(collection),
	}
}

func (fs *FireStore) Get(key string, value interface{}) error {
	res, err := fs.CollectionRef.Doc(key).Get(context.Background())
	if err != nil {
		return err
	}
	err = res.DataTo(&value)
	return err
}

func (fs *FireStore) Set(key string, value interface{}) error {
	_, err := fs.CollectionRef.Doc(key).Set(context.Background(), value)
	return err
}
