package feed

import (
	"context"
	"errors"
	"fmt"
	"rss-summary/pkg/rss"

	"github.com/mmcdole/gofeed"
)

var (
	ErrEmptyFeed = errors.New("feed is empty")
)

type (
	GoFeed struct {
		parser *gofeed.Parser
	}
)

func NewGoFeed() *GoFeed {
	return &GoFeed{
		parser: gofeed.NewParser(),
	}
}

func (r GoFeed) Fetch(_ context.Context, url string) (rss.Feed, error) {
	feed, err := r.parser.ParseURL(url)
	if err != nil {
		return rss.Feed{}, fmt.Errorf("faield to parse feed: %w", err)
	}

	if len(feed.Items) == 0 {
		return rss.Feed{}, ErrEmptyFeed
	}

	articles := make([]rss.Article, len(feed.Items))
	for i, item := range feed.Items {
		imageURL := ""
		if item.Image != nil {
			imageURL = item.Image.URL
		}

		article, err := rss.NewArticle(feed.Title, item.Title, item.Link, imageURL, item.Published)
		if err != nil {
			return rss.Feed{}, err
		}

		articles[i] = article
	}

	return rss.NewFeed(feed.Title, url, articles), nil
}
