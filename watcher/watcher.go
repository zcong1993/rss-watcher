package watcher

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/zcong1993/notifiers/types"
	"github.com/zcong1993/rss-watcher/kv"
)

type IWatcher interface {
}

type Config struct {
}

type RSSWatcher struct {
	config    Config
	source    string
	md5Source string
	store     kv.Store
	notifiers []types.Notifier
	interval  time.Duration
	closeCh   chan struct{}
	parser    *gofeed.Parser
}

func NewRSSWatcher(source string, interval time.Duration, store kv.Store, notifiers []types.Notifier) *RSSWatcher {
	return &RSSWatcher{
		source:    source,
		interval:  interval,
		store:     store,
		md5Source: fmt.Sprintf("%x", md5.Sum([]byte(source))),
		notifiers: notifiers,
		closeCh:   make(chan struct{}),
		parser:    gofeed.NewParser(),
	}
}

func (rw *RSSWatcher) Run() {
	t := time.NewTicker(rw.interval)
	defer t.Stop()
	err := rw.handle()
	if err != nil {
		fmt.Printf("rss watcher handle error: %+v\n", err)
	}
	for {
		select {
		case <-t.C:
			err := rw.handle()
			if err != nil {
				fmt.Printf("rss watcher handle error: %+v\n", err)
			}
		case <-rw.closeCh:
			fmt.Printf("source for %s watcher exist.\n", rw.source)
			return
		}
	}
}

func (rw *RSSWatcher) Close() {
	rw.closeCh <- struct{}{}
}

func (rw *RSSWatcher) handle() error {
	feed, err := rw.parser.ParseURL(rw.source)
	if err != nil {
		return err
	}
	var items []*gofeed.Item
	var last gofeed.Item
	err = rw.store.Get(rw.md5Source, &last)
	if err != nil && !strings.Contains(err.Error(), "code = NotFound") {
		return err
	}
	if err != nil {
		items = append(items, feed.Items[0])
	} else {
		for _, item := range feed.Items {
			if item.GUID == last.GUID {
				break
			}
			items = append(items, item)
		}
	}

	for _, item := range items {
		msg := types.WrapMsg(&types.Message{
			Title:   item.Title,
			Content: item.Description,
			Tags:    []string{"rss-watcher", rw.source},
			URL:     item.Link,
		})

		for _, notifier := range rw.notifiers {
			err := notifier.Notify(msg)
			if err != nil {
				fmt.Printf("notify error: %+v\n", err)
			}
		}
	}

	if len(items) > 0 {
		err = rw.store.Set(rw.md5Source, items[0])
		if err != nil {
			fmt.Printf("kv store add last item error: %+v\n", err)
		}
	}

	return nil
}
