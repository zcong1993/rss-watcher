package watcher

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"

	"github.com/mmcdole/gofeed"
	"github.com/zcong1993/notifiers/types"
	"github.com/zcong1993/rss-watcher/pkg/kv"
)

func normalizeContent(content string, length int) string {
	sl := len(content)
	if sl <= length {
		return content
	}
	return content[:length] + "..."
}

type RSSWatcher struct {
	logger    log.Logger
	source    string
	md5Source string
	store     kv.Store
	skip      int
	notifiers []types.Notifier
	interval  time.Duration
	closeCh   chan struct{}
	parser    *gofeed.Parser
}

func NewRSSWatcher(logger log.Logger, source string, interval time.Duration, store kv.Store, notifiers []types.Notifier, skip int) *RSSWatcher {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{Timeout: time.Second * 10}
	return &RSSWatcher{
		logger:    log.WithPrefix(logger, "component", "watcher", "source", source),
		source:    source,
		interval:  interval,
		store:     store,
		md5Source: fmt.Sprintf("%x", md5.Sum([]byte(source))),
		notifiers: notifiers,
		closeCh:   make(chan struct{}),
		parser:    parser,
		skip:      skip,
	}
}

func (rw *RSSWatcher) Single() error {
	err := rw.handle()
	if err != nil {
		level.Error(rw.logger).Log("msg", fmt.Sprintf("rss watcher handle error: %s", err))
		return err
	}
	return nil
}

func (rw *RSSWatcher) Run() {
	t := time.NewTicker(rw.interval)
	defer t.Stop()
	err := rw.handle()
	if err != nil {
		level.Error(rw.logger).Log("msg", fmt.Sprintf("rss watcher handle error: %s", err))
	}
	for {
		select {
		case <-t.C:
			err := rw.handle()
			if err != nil {
				level.Error(rw.logger).Log("msg", fmt.Sprintf("rss watcher handle error: %s", err))
			}
		case <-rw.closeCh:
			level.Info(rw.logger).Log("msg", fmt.Sprintf("source for %s watcher exit", rw.source))
			return
		}
	}
}

func (rw *RSSWatcher) Close() {
	rw.closeCloser(rw.store, "store")
	close(rw.closeCh)
}

func (rw *RSSWatcher) handle() error {
	level.Info(rw.logger).Log("msg", fmt.Sprintf("run tick %s", time.Now().String()))
	feed, err := rw.parser.ParseURL(rw.source)
	if err != nil {
		return errors.Wrapf(err, "get feed error, url: %s", rw.source)
	}
	var items []*gofeed.Item
	var last gofeed.Item
	isNew := false
	val, err := rw.store.Get(rw.md5Source)
	if err != nil && !errors.Is(err, kv.ErrNotFound) {
		return errors.Wrap(err, "store get")
	}

	if errors.Is(err, kv.ErrNotFound) {
		isNew = true
	} else {
		err = json.Unmarshal([]byte(val), &last)
		if err != nil && !errors.Is(err, kv.ErrNotFound) {
			return errors.Wrap(err, "store get unmarshal")
		}
	}

	feedItems := feed.Items

	if rw.skip > 0 {
		if len(feedItems) < rw.skip {
			return errors.New("feed length less than skip")
		}
		feedItems = feed.Items[rw.skip:]
	}

	if len(feedItems) == 0 {
		return errors.New("feed length = 0")
	}

	if isNew {
		level.Info(rw.logger).Log("msg", "new feed")
		items = append(items, feedItems[0])
	} else {
		for i, item := range feedItems {
			if item.Link == last.Link {
				break
			}
			if i > 4 {
				// max 5 items
				break
			}
			items = append(items, item)
		}
	}

	if len(items) == 0 {
		level.Info(rw.logger).Log("msg", "no new feed")
		return nil
	}

	for _, item := range items {
		msg := types.WrapMsg(&types.Message{
			Title:   item.Title,
			Content: normalizeContent(item.Description, 300),
			Tags:    []string{"rss-watcher", rw.source},
			URL:     item.Link,
		})

		for _, notifier := range rw.notifiers {
			msgCp := msg.Clone()
			level.Info(rw.logger).Log("msg", fmt.Sprintf("notifier %s notify msg", notifier.GetName()))
			err := notifier.Notify(msgCp)
			if err != nil {
				level.Error(rw.logger).Log("msg", fmt.Sprintf("notify %s error: %+v", notifier.GetName(), err))
			}
		}
	}

	bt, err := json.Marshal(items[0])
	if err != nil {
		level.Error(rw.logger).Log("msg", fmt.Sprintf("kv store save json marshal error: %+v", err))
	} else {
		err = rw.store.Set(rw.md5Source, string(bt))
		if err != nil {
			level.Error(rw.logger).Log("msg", fmt.Sprintf("kv store add last item error: %+v", err))
		}
	}

	return nil
}

func (rw *RSSWatcher) closeCloser(closer io.Closer, name string) {
	if err := closer.Close(); err != nil {
		level.Error(rw.logger).Log("msg", fmt.Sprintf("%s close error", name), "error", err.Error())
	}
}
