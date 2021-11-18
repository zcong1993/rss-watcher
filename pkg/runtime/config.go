package runtime

import (
	"encoding/json"
	"time"

	"github.com/zcong1993/rss-watcher/pkg/store/pg"

	"github.com/zcong1993/rss-watcher/pkg/notify"

	"github.com/pkg/errors"
	"github.com/zcong1993/rss-watcher/pkg/store/fauna"
	"github.com/zcong1993/rss-watcher/pkg/store/file"
)

type duration struct {
	time.Duration
}

func (d duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *duration) UnmarshalJSON(b []byte) error {
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

		return err
	default:
		return errors.New("invalid duration")
	}
}

type Config struct {
	DingConfig          *notify.DingConfig     `json:"ding_config" validate:"omitempty"`
	MailConfig          *notify.MailConfig     `json:"mail_config" validate:"omitempty"`
	TelegramConfig      *notify.TelegramConfig `json:"telegram_config" validate:"omitempty"`
	KvStore             string                 `json:"kv_store" validate:"required,oneof=mem file fauna pg"`
	FileStoreConfigPath string                 `json:"file_store_config_path" validate:"omitempty"`
	WatcherConfigs      []WatcherConfig        `json:"watcher_configs" validate:"gt=0,dive"`
	FaunaConfig         *fauna.Config          `json:"fauna_config"`
	FileConfig          *file.Config           `json:"file_config"`
	PgConfig            *pg.Config             `json:"pg_config"`
}

type WatcherConfig struct {
	Source    string    `json:"source" validate:"required"`
	Interval  *duration `json:"interval" validate:"required"`
	Notifiers []string  `json:"notifiers" validate:"gt=0,dive,oneof=ding mail tg printer"`
	Skip      int       `json:"skip" validate:"omitempty,gte=0"`
}
