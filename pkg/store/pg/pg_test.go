package pg_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zcong1993/rss-watcher/pkg/store/pg"
)

func TestPg(t *testing.T) {
	p := pg.NewPg()
	err := p.Init(&pg.Config{
		DbURL: os.Getenv("PG"),
		Table: "rss_kv",
	})
	assert.Nil(t, err)

	err = p.Set(context.Background(), "test", "haha2")
	assert.Nil(t, err)
}
