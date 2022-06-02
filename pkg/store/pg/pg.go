package pg

import (
	"context"
	"encoding/json"
	"github.com/zcong1993/rss-watcher/pkg/store"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const Name = "pg"

type Config struct {
	DbURL string `json:"db_url" validate:"required"`
	Table string `json:"table" validate:"required"`
}

type pg struct {
	client *gorm.DB
	table  string
}

type kv struct {
	Key       string `gorm:"primarykey"`
	Content   string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewPg() *pg {
	return &pg{}
}

func (p *pg) Init(cfg interface{}) error {
	config, ok := cfg.(*Config)
	if !ok {
		return errors.New("invalid pg config")
	}

	err := validator.New().Struct(config)
	if err != nil {
		return errors.Wrap(err, "valid config error")
	}

	conn, err := gorm.Open(postgres.Open(config.DbURL))

	if err != nil {
		return errors.Wrapf(err, "create pg connect error")
	}

	p.client = conn
	p.table = config.Table

	return nil
}

func (p *pg) Get(ctx context.Context, key string) (string, error) {
	var res kv
	err := p.client.WithContext(ctx).Table(p.table).First(&res, "key = ?", key).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", store.ErrNotFound
		}
		return "", err
	}

	return res.Content, nil
}

func (p *pg) Set(ctx context.Context, key string, value string) error {
	res := p.client.WithContext(ctx).Table(p.table).Where("key = ?", key).Update("content", value)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return p.client.WithContext(ctx).Table(p.table).Create(&kv{
			Key:     key,
			Content: value,
		}).Error
	}

	return nil
}

func (p *pg) Close() error {
	db, err := p.client.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (p *pg) Import(str string) error {
	var datas []kv
	err := json.Unmarshal([]byte(str), &datas)
	if err != nil {
		return err
	}
	return p.client.Table(p.table).CreateInBatches(datas, 100).Error
}
