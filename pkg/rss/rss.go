package rss

import (
	"sort"

	"github.com/mmcdole/gofeed"
)

var parser = gofeed.NewParser()

// sort items by date desc.
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
