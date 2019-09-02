package kv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const endpoint = "https://pfqxxox5kg.execute-api.ap-southeast-1.amazonaws.com/production/api"

type DynamoKvClient struct {
	token     string
	client    *http.Client
	Namespace string
}

type Item struct {
	Namespace string    `json:"namespace"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

func NewDynamoKvClient(namespace, token string) *DynamoKvClient {
	client := &http.Client{Timeout: time.Second * 10}
	return &DynamoKvClient{
		token:     token,
		client:    client,
		Namespace: namespace,
	}
}

func (dk *DynamoKvClient) auth(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("token %s", dk.token))
}

func (dk *DynamoKvClient) Get(key string, value interface{}) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/%s", endpoint, dk.Namespace, key), nil)
	if err != nil {
		return err
	}

	dk.auth(req)

	res, err := dk.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	var item Item

	err = json.NewDecoder(res.Body).Decode(&item)

	if err != nil {
		return err
	}

	// not found
	if item.Key != key || item.Namespace != dk.Namespace {
		return ErrNotFound
	}

	return json.Unmarshal([]byte(item.Value), value)
}

func (dk *DynamoKvClient) Set(key string, value interface{}) error {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(value)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s/%s", endpoint, dk.Namespace, key), &body)
	if err != nil {
		return err
	}

	dk.auth(req)

	_, err = dk.client.Do(req)
	return err
}
