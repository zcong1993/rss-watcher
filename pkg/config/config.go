package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	"gopkg.in/go-playground/validator.v9"
)

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type Config struct {
	DingConfig          *DingConfig        `json:"ding_config" validate:"omitempty,dive"`
	MailConfig          *MailConfig        `json:"mail_config" validate:"omitempty,dive"`
	TelegramConfig      *TelegramConfig    `json:"telegram_config" validate:"omitempty,dive"`
	KvStore             string             `json:"kv_store" validate:"required,oneof=mem file"`
	FileStoreConfigPath string             `json:"file_store_config_path" validate:"omitempty"`
	RedisUri            string             `json:"redis_uri" validate:"omitempty"`
	WatcherConfigs      []WatcherConfig    `json:"watcher_configs" validate:"gt=0,dive"`
	Single              bool               `json:"single" validate:"omitempty"`
	CloudConfigConfig   *CloudConfigConfig `json:"cloud_config_config" validate:"omitempty,dive"`
}

type DingConfig struct {
	Token string `json:"token" validate:"required"`
}

type MailConfig struct {
	Domain     string `json:"domain" validate:"required"`
	PrivateKey string `json:"private_key" validate:"required"`
	From       string `json:"from" validate:"required"`
	To         string `json:"to" validate:"required"`
}

type WatcherConfig struct {
	Source    string    `json:"source" validate:"required"`
	Interval  *Duration `json:"interval" validate:"required"`
	Notifiers []string  `json:"notifiers" validate:"gt=0,dive,oneof=ding mail tg printer"`
	Skip      int       `json:"skip" validate:"omitempty,gte=0"`
}

type TelegramConfig struct {
	Token string `json:"token" validate:"required"`
	ToID  int64  `json:"to_id" validate:"required"`
}

type CloudConfigConfig struct {
	Token    string `json:"token" validate:"required"`
	Endpoint string `json:"endpoint" validate:"required"`
}

func LoadConfigFromFile(f string) *Config {
	configBytes, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		panic(err)
	}
	validateConfig(&config)
	return &config
}

func validateConfig(c *Config) {
	err := validator.New().Struct(c)
	if err != nil {
		panic(err)
	}
	notifiers := make(map[string]struct{})
	for _, rw := range c.WatcherConfigs {
		for _, ntf := range rw.Notifiers {
			switch ntf {
			case "ding":
				notifiers["ding"] = struct{}{}
			case "mail":
				notifiers["mail"] = struct{}{}
			case "tg":
				notifiers["tg"] = struct{}{}
			case "printer":
				notifiers["printer"] = struct{}{}
			}
		}
	}
	for k := range notifiers {
		switch k {
		case "ding":
			if c.DingConfig == nil {
				panic("ding config is required")
			}
		case "mail":
			if c.MailConfig == nil {
				panic("mail config is required")
			}
		case "tg":
			if c.TelegramConfig == nil {
				panic("telegram config is required")
			}
		}
	}
	switch c.KvStore {
	case "file":
		if c.FileStoreConfigPath == "" {
			panic("file_store_config_path is required when kv_store is file")
		}
	case "cloud-config":
		if c.CloudConfigConfig == nil {
			panic("cloud_config_config is required when kv_store is cloud-config")
		}
	}
}
