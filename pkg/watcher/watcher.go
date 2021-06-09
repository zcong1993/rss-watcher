package watcher

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"

	"github.com/mmcdole/gofeed"
	"github.com/zcong1993/notifiers/v2"
	"github.com/zcong1993/rss-watcher/pkg/kv"
)

// sort items by date desc
type Items []*gofeed.Item

func (p Items) Len() int { return len(p) }
func (p Items) Less(i, j int) bool {
	iDate := p[i].PublishedParsed
	if iDate == nil {
		iDate = p[i].UpdatedParsed
	}

	jDate := p[j].PublishedParsed
	if jDate == nil {
		jDate = p[j].UpdatedParsed
	}

	return iDate.After(*jDate)
}
func (p Items) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func getItemId(item *gofeed.Item) string {
	if len(item.GUID) > 0 {
		return item.GUID
	}

	return item.Link
}

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
	notifier  notifiers.Notifier
	interval  time.Duration
	closeCh   chan struct{}
	parser    *gofeed.Parser
}

func NewRSSWatcher(logger log.Logger, source string, interval time.Duration, store kv.Store, notifier notifiers.Notifier, skip int) *RSSWatcher {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{Timeout: time.Second * 10}
	return &RSSWatcher{
		logger:    log.WithPrefix(logger, "component", "watcher", "source", source),
		source:    source,
		interval:  interval,
		store:     store,
		md5Source: fmt.Sprintf("%x", md5.Sum([]byte(source))),
		notifier:  notifier,
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

	return rw.notifier.Close()
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
	isNew := false
	lastItemId, err := rw.store.Get(ctx, rw.md5Source)
	if err != nil && !errors.Is(err, kv.ErrNotFound) {
		return errors.Wrap(err, "store get")
	}

	if errors.Is(err, kv.ErrNotFound) {
		isNew = true
	}

	feedItems := feed.Items
	sort.Sort(Items(feedItems))

	//js, _ := json.Marshal(feedItems)
	//fmt.Println(string(js))

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
			if getItemId(item) == lastItemId {
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
		level.Info(rw.logger).Log("msg", fmt.Sprintf("notifier %s notify msg", rw.notifier.GetName()))
		err := rw.notifier.Notify(ctx, "", msg)
		if err != nil {
			level.Error(rw.logger).Log("msg", fmt.Sprintf("notify %s error: %+v", rw.notifier.GetName(), err))
		}
	}

	err = rw.store.Set(ctx, rw.md5Source, getItemId(items[0]))
	if err != nil {
		level.Error(rw.logger).Log("msg", fmt.Sprintf("kv store add last item error: %+v", err))
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

	return notifiers.MessageFromContent(content)
}
