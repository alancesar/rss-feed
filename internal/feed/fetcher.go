package feed

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"rss-feed/pkg/event"

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

	feedID := hashFromString(feed.FeedLink)
	articles := make([]event.Article, len(feed.Items))
	for i, item := range feed.Items {
		articleID := hashFromString(item.Link)
		article := event.Article{
			ArticleID: articleID,
			FeedID:    feedID,
			Title:     item.Title,
			URL:       item.Link,
		}

		if item.PublishedParsed != nil {
			article.PublishedAt = *item.PublishedParsed
		}

		if item.Image != nil {
			article.Image = event.Image{
				ImageID:   hashFromString(item.Image.URL),
				ArticleID: articleID,
				URL:       item.Image.URL,
			}
		}

		articles[i] = article
	}

	return event.Feed{
		FeedID:    feedID,
		Name:      feed.Title,
		URL:       feed.Link,
		UpdatedAt: feed.UpdatedParsed,
		Articles:  articles,
	}, nil
}

func hashFromString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
