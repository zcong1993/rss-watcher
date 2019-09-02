package watcher

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/zcong1993/notifiers/types"
	"github.com/zcong1993/rss-watcher/kv"
)

type RSSWatcher struct {
	source    string
	md5Source string
	store     kv.Store
	skip      int
	notifiers []types.Notifier
	interval  time.Duration
	closeCh   chan struct{}
	parser    *gofeed.Parser
}

func NewRSSWatcher(source string, interval time.Duration, store kv.Store, notifiers []types.Notifier, skip int) *RSSWatcher {
	return &RSSWatcher{
		source:    source,
		interval:  interval,
		store:     store,
		md5Source: fmt.Sprintf("%x", md5.Sum([]byte(source))),
		notifiers: notifiers,
		closeCh:   make(chan struct{}),
		parser:    gofeed.NewParser(),
		skip:      skip,
	}
}

func (rw *RSSWatcher) Single() error {
	err := rw.handle()
	if err != nil {
		fmt.Printf("rss watcher handle error: %+v\n", err)
		return err
	}
	return nil
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
	fmt.Printf("run tick %s\n", time.Now().String())
	feed, err := rw.parser.ParseURL(rw.source)
	if err != nil {
		return err
	}
	var items []*gofeed.Item
	var last gofeed.Item
	err = rw.store.Get(rw.md5Source, &last)
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		fmt.Printf("err %+v\n", err)
		return err
	}

	feedItems := feed.Items

	if rw.skip > 0 {
		if len(feedItems) < rw.skip {
			return errors.New("feed length less than skip")
		}
		feedItems = feed.Items[rw.skip:]
	}

	if err != nil {
		items = append(items, feedItems[0])
	} else {
		for i, item := range feedItems {
			if item.GUID == last.GUID {
				break
			}
			if i > 4 {
				// max 5 items
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
			msgCp := msg.Clone()
			fmt.Printf("notifier %s notify msg\n", notifier.GetName())
			err := notifier.Notify(msgCp)
			if err != nil {
				fmt.Printf("notify %s error: %+v\n", notifier.GetName(), err)
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
