package rss

import (
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
)

var parser = gofeed.NewParser()

// sort items by date desc.
type Items []*gofeed.Item

func (p Items) Len() int { return len(p) }
func (p Items) Less(i, j int) bool {
	return getItemDate(p[i]).After(*getItemDate(p[j]))
}
func (p Items) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func getItemDate(item *gofeed.Item) *time.Time {
	if item.PublishedParsed != nil {
		return item.PublishedParsed
	}

	return item.UpdatedParsed
}

func GetItemId(item *gofeed.Item) string {
	if len(item.GUID) > 0 {
		return item.GUID
	}

	return item.Link
}

func GetFeedItems(url string) ([]*gofeed.Item, error) {
	feed, err := parser.ParseURL(url)
	if err != nil {
		return nil, err
	}

	feedItems := feed.Items
	sort.Sort(Items(feedItems))

	return feedItems, nil
}
