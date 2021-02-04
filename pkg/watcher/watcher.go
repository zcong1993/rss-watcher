package watcher

import (
	"context"
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
	"github.com/zcong1993/notifiers/v2"
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
	notifiers []notifiers.Notifier
	interval  time.Duration
	closeCh   chan struct{}
	parser    *gofeed.Parser
}

func NewRSSWatcher(logger log.Logger, source string, interval time.Duration, store kv.Store, notifiers []notifiers.Notifier, skip int) *RSSWatcher {
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

func (rw *RSSWatcher) Single(ctx context.Context) error {
	err := rw.handle(ctx)
	if err != nil {
		level.Error(rw.logger).Log("msg", fmt.Sprintf("rss watcher handle error: %s", err))
		return err
	}
	return nil
}

func (rw *RSSWatcher) Run(ctx context.Context) {
	t := time.NewTicker(rw.interval)
	defer t.Stop()
	err := rw.handle(ctx)
	if err != nil {
		level.Error(rw.logger).Log("msg", fmt.Sprintf("rss watcher handle error: %s", err))
	}
	for {
		select {
		case <-t.C:
			err := rw.handle(ctx)
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

func (rw *RSSWatcher) handle(ctx context.Context) error {
	level.Info(rw.logger).Log("msg", fmt.Sprintf("run tick %s", time.Now().String()))
	feed, err := rw.parser.ParseURL(rw.source)
	if err != nil {
		return errors.Wrapf(err, "get feed error, url: %s", rw.source)
	}
	var items []*gofeed.Item
	var last gofeed.Item
	isNew := false
	val, err := rw.store.Get(ctx, rw.md5Source)
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
		msg := feed2Message(item)

		for _, notifier := range rw.notifiers {
			level.Info(rw.logger).Log("msg", fmt.Sprintf("notifier %s notify msg", notifier.GetName()))
			err := notifier.Notify(ctx, "", msg)
			if err != nil {
				level.Error(rw.logger).Log("msg", fmt.Sprintf("notify %s error: %+v", notifier.GetName(), err))
			}
		}
	}

	bt, err := json.Marshal(items[0])
	if err != nil {
		level.Error(rw.logger).Log("msg", fmt.Sprintf("kv store save json marshal error: %+v", err))
	} else {
		err = rw.store.Set(ctx, rw.md5Source, string(bt))
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

func feed2Message(item *gofeed.Item) notifiers.Message {
	content := fmt.Sprintf(`RSS-WATCHER

%s

%s

%s`, item.Title, normalizeContent(item.Description, 300), item.Link)

	return notifiers.Message{Content: content}
}
