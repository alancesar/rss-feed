package feed

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"rss-summary/pkg/rss"

	"github.com/mmcdole/gofeed"
)

var (
	ErrEmptyFeed = errors.New("feed is empty")
)

type (
	FileStorage interface {
		Create(ctx context.Context, path string, img io.Reader) error
	}

	GoFeed struct {
		parser  *gofeed.Parser
		storage FileStorage
		client  *http.Client
	}
)

func NewGoFeed(storage FileStorage, client *http.Client) *GoFeed {
	return &GoFeed{
		parser:  gofeed.NewParser(),
		storage: storage,
		client:  client,
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
		article, err := rss.NewArticle(feed.Title, item.Title, item.Link, item.Published)
		if err != nil {
			return rss.Feed{}, err
		}

		if item.Image != nil {
			if err := r.handleImage(ctx, item.Image.URL, article.ImagePath()); err != nil {
				return rss.Feed{}, err
			}
		}

		articles[i] = article
	}

	return rss.NewFeed(feed.Title, url, articles), nil
}

func (r GoFeed) handleImage(ctx context.Context, source, target string) error {
	resp, err := r.client.Get(source)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch image: %s", resp.Status)
	}

	if err := r.storage.Create(ctx, target, resp.Body); err != nil {
		return err
	}

	return nil
}
