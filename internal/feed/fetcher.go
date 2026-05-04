package feed

import (
	"context"
	"errors"
	"fmt"
	"rss-feed/pkg/event"
	"rss-feed/pkg/hash"

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

func (r GoFeed) Fetch(ctx context.Context, url string) (event.Feed, error) {
	feed, err := r.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return event.Feed{}, fmt.Errorf("failed to parse feed: %w", err)
	}

	if len(feed.Items) == 0 {
		return event.Feed{}, ErrEmptyFeed
	}

	feedID := hash.Hash(feed.FeedLink)
	articles := make([]event.Article, len(feed.Items))
	for i, item := range feed.Items {
		articleID := hash.Hash(item.GUID)
		article := event.Article{
			ArticleID:   articleID,
			FeedID:      feedID,
			Title:       item.Title,
			Description: item.Description,
			Content:     item.Content,
			Categories:  item.Categories,
			URL:         item.Link,
		}

		if item.PublishedParsed != nil {
			article.PublishedAt = *item.PublishedParsed
		}

		if item.Image != nil {
			article.Image = event.Image{
				ImageID:   hash.Hash(articleID, item.Image.URL),
				ArticleID: articleID,
				URL:       item.Image.URL,
			}
		}

		articles[i] = article
	}

	return event.Feed{
		FeedID:    feedID,
		Name:      feed.Title,
		URL:       feed.FeedLink,
		UpdatedAt: feed.UpdatedParsed,
		Articles:  articles,
	}, nil
}
