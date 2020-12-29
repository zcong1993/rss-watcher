package config

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"

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
	KvStore             string             `json:"kv_store" validate:"required,oneof=mem file fauna"`
	FileStoreConfigPath string             `json:"file_store_config_path" validate:"omitempty"`
	RedisUri            string             `json:"redis_uri" validate:"omitempty"`
	WatcherConfigs      []WatcherConfig    `json:"watcher_configs" validate:"gt=0,dive"`
	Single              bool               `json:"single" validate:"omitempty"`
	CloudConfigConfig   *CloudConfigConfig `json:"cloud_config_config" validate:"omitempty,dive"`
	FaunaConfig         *FaunaConfig       `json:"fauna_config" validate:"omitempty,dive"`
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

type FaunaConfig struct {
	Secret     string `json:"secret" validate:"required"`
	Collection string `json:"collection" validate:"required"`
	IndexName  string `json:"index_name" validate:"required"`
}

func LoadConfigFromFile(f string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, errors.Wrap(err, "read config file")
	}
	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config file")
	}
	err = validateConfig(&config)
	return &config, err
}

func validateConfig(c *Config) error {
	err := validator.New().Struct(c)
	if err != nil {
		return errors.Wrap(err, "validate config")
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
				return errors.New("ding config is required")
			}
		case "mail":
			if c.MailConfig == nil {
				return errors.New("mail config is required")
			}
		case "tg":
			if c.TelegramConfig == nil {
				return errors.New("telegram config is required")
			}
		}
	}
	switch c.KvStore {
	case "file":
		if c.FileStoreConfigPath == "" {
			return errors.New("file_store_config_path is required when kv_store is file")
		}
	case "cloud-config":
		if c.CloudConfigConfig == nil {
			return errors.New("cloud_config_config is required when kv_store is cloud-config")
		}
	case "fauna":
		if c.FaunaConfig == nil {
			return errors.New("fauna_config is required when kv_store is cloud-config")
		}
	}
	return nil
}
