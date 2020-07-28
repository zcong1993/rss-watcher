package kv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	group = "rss"
	meta  = `{ "headers": {"Content-Type": "application/json"} }`
)

type CloudConfig struct {
	endpoint string
	token    string
	client   *http.Client
}

func NewCloudConfig(endpoint string, token string) *CloudConfig {
	client := &http.Client{Timeout: time.Second * 10}
	return &CloudConfig{
		client:   client,
		endpoint: endpoint,
		token:    token,
	}
}

func (cc *CloudConfig) Get(key string, value interface{}) error {
	u := fmt.Sprintf("%s/api/private?token=%s&key=%s&group=%s", cc.endpoint, cc.token, key, group)
	resp, err := cc.client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusBadRequest {
		return ErrNotFound
	}
	return json.NewDecoder(resp.Body).Decode(&value)
}

func (cc *CloudConfig) Set(key string, value interface{}) error {
	u := fmt.Sprintf("%s/api/put?token=%s", cc.endpoint, cc.token)
	valueStr, err := json.Marshal(value)
	if err != nil {
		return err
	}
	body := map[string]string{
		"key":     key,
		"val":     string(valueStr),
		"group":   group,
		"visible": "PRIVATE",
		"meta":    meta,
	}
	var bf bytes.Buffer
	err = json.NewEncoder(&bf).Encode(body)
	if err != nil {
		return err
	}
	resp, err := http.Post(u, "application/json", &bf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			fmt.Printf("error: %s\n", body)
		}
		return errors.New("set error")
	}
	return nil
}
