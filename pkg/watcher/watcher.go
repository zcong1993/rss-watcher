package watcher

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/zcong1993/rss-watcher/pkg/logger"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"github.com/zcong1993/notifiers/v2"
	"github.com/zcong1993/rss-watcher/pkg/rss"
	"github.com/zcong1993/rss-watcher/pkg/store"
)

type Watcher struct {
	store    store.Store
	notifier notifiers.Notifier
	url      string
	md5URL   string
	interval time.Duration
	stopCh   chan struct{}
	logger   *logger.Logger
}

func NewWatcher(logger *logger.Logger, url string, store store.Store, notifier notifiers.Notifier, interval time.Duration) *Watcher {
	w := &Watcher{
		store:    store,
		notifier: notifier,
		url:      url,
		md5URL:   fmt.Sprintf("%x", md5.Sum([]byte(url))),
		interval: interval,
		stopCh:   make(chan struct{}, 1),
		logger:   logger.WithField("source", url),
	}

	return w
}

func (w *Watcher) Run() {
	t := time.NewTicker(w.interval)
	defer t.Stop()

	for {
		ctx, cancel := context.WithCancel(context.Background())
		err := w.handle(ctx)

		if err != nil {
			w.logger.Errorf("rss watcher handle error: %s", err.Error())
		}

		select {
		case <-w.stopCh:
			cancel()
			w.logger.Info("watcher exit")

			return
		case <-t.C:
		}
	}
}

func (w *Watcher) Single(ctx context.Context) error {
	if err := w.handle(ctx); err != nil {
		w.logger.Errorf("rss watcher handle error: %s", err.Error())

		return err
	}

	w.notifier.Wait()

	return nil
}

func (w *Watcher) Close() {
	close(w.stopCh)
}

func (w *Watcher) handle(ctx context.Context) error {
	w.logger.Infof("run tick %s", time.Now().String())

	feedItems, err := rss.GetFeedItems(w.url)

	if err != nil {
		return errors.Wrap(err, "get feed")
	}

	lastItemID, err := w.store.Get(ctx, w.md5URL)

	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return errors.Wrap(err, "store get")
	}

	isNew := errors.Is(err, store.ErrNotFound)

	if len(feedItems) == 0 {
		return errors.New("empty feed items")
	}

	items := make([]*gofeed.Item, 0)

	if isNew {
		w.logger.Info("new feed source")

		items = append(items, feedItems[0])
	} else {
		for i, item := range feedItems {
			if rss.GetItemId(item) == lastItemID {
				break
			}
			if i > 4 {
				// max 5 items.
				break
			}
			items = append(items, item)
		}
	}

	if len(items) == 0 {
		w.logger.Info("no new feed")

		return nil
	}

	for _, item := range items {
		msg := feed2Message(item)

		w.logger.Infof("notifier %s notify msg", w.notifier.GetName())
		err := w.notifier.Notify(ctx, "", msg)

		if err != nil {
			w.logger.Errorf("notify %s error: %+v", w.notifier.GetName(), err)
		}
	}

	err = w.store.Set(ctx, w.md5URL, rss.GetItemId(items[0]))
	if err != nil {
		w.logger.Errorf("kv store add last item error: %+v", err)
	}

	return nil
}

func normalizeContent(content string, length int) string {
	if len(content) <= length {
		return content
	}

	return content[:length] + "..."
}

func feed2Message(item *gofeed.Item) notifiers.Message {
	content := fmt.Sprintf(`RSS-WATCHER
%s
%s
%s`, item.Title, normalizeContent(item.Description, 300), item.Link)

	return notifiers.MessageFromContent(content)
}
