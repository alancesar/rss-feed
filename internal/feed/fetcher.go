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
	PublishImageFn func(ctx context.Context, articleID, sourceURL string) error

	GoFeed struct {
		parser  *gofeed.Parser
		publish PublishImageFn
	}
)

func NewGoFeed(publisher PublishImageFn) *GoFeed {
	return &GoFeed{
		parser:  gofeed.NewParser(),
		publish: publisher,
	}
}

func (r GoFeed) Fetch(ctx context.Context, url string) (rss.Feed, error) {
	feed, err := r.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return rss.Feed{}, fmt.Errorf("faield to parse feed: %w", err)
	}

	if len(feed.Items) == 0 {
		return rss.Feed{}, ErrEmptyFeed
	}

	articles := make([]rss.Article, len(feed.Items))
	for i, item := range feed.Items {
		article, err := rss.NewArticle(feed.Title, item.Title, item.Link, item.PublishedParsed)
		if err != nil {
			return rss.Feed{}, err
		}

		if item.Image != nil {
			if err := r.publish(ctx, article.ID, item.Image.URL); err != nil {
				return rss.Feed{}, fmt.Errorf("failed to publish image: %w", err)
			}
		}

		articles[i] = article
	}

	return rss.NewFeed(feed.Title, url, articles), nil
}
