// Package feed provides clients that read RSS and Atom feeds.
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
	// ErrEmptyFeed is returned when a feed is parsed successfully but contains no items.
	ErrEmptyFeed = errors.New("feed is empty")
)

type (
	// GoFeed reads feeds with the gofeed parser, which handles RSS, Atom and JSON feeds.
	GoFeed struct {
		parser *gofeed.Parser
	}
)

// NewGoFeed creates a GoFeed with a default gofeed parser.
func NewGoFeed() *GoFeed {
	return &GoFeed{
		parser: gofeed.NewParser(),
	}
}

// Fetch downloads and parses the feed at url and converts it into an event.Feed. Feed and
// article IDs are hashes of the feed link and of each item link; an article's image ID is a
// hash of the article ID and the image URL, and is only set when the item has an image.
// It fails with ErrEmptyFeed when the feed has no items.
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
		articleID := hash.Hash(item.Link)
		article := event.Article{
			ArticleID:   articleID,
			FeedID:      feedID,
			FeedName:    feed.Title,
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
