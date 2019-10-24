package kv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const endpoint = "https://0os7mxakwk.execute-api.ap-southeast-1.amazonaws.com/production/api"

type DynamoKvClient struct {
	token  string
	client *http.Client
}

func NewDynamoKvClient(token string) *DynamoKvClient {
	client := &http.Client{Timeout: time.Second * 10}
	return &DynamoKvClient{
		token:  token,
		client: client,
	}
}

func (dk *DynamoKvClient) Get(key string, value interface{}) error {
	res, err := dk.client.Get(fmt.Sprintf("%s/%s?token=%s", endpoint, key, dk.token))
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNoContent {
		return ErrNotFound
	}

	return json.NewDecoder(res.Body).Decode(&value)
}

func (dk *DynamoKvClient) Set(key string, value interface{}) error {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(value)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s?token=%s", endpoint, key, dk.token), &body)
	if err != nil {
		return err
	}
	_, err = dk.client.Do(req)
	return err
}
