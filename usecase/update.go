package usecase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"rss-summary/pkg/rss"
)

type (
	FeedsStore interface {
		FeedStore
		GetAllFeedURLs(ctx context.Context) ([]string, error)
	}

	FileStorage interface {
		Create(ctx context.Context, path string, img io.Reader) error
	}

	UpdateFeeds struct {
		store   FeedsStore
		fetcher FeedFetcher
		storage FileStorage
		client  *http.Client
	}
)

func NewUpdateFeeds(store FeedsStore, fetcher FeedFetcher, storage FileStorage) *UpdateFeeds {
	return &UpdateFeeds{
		store:   store,
		fetcher: fetcher,
		storage: storage,
		client:  http.DefaultClient,
	}
}

func (f UpdateFeeds) Execute(ctx context.Context) error {
	urls, err := f.store.GetAllFeedURLs(ctx)
	if err != nil {
		return err
	}

	for _, url := range urls {
		fmt.Println("fetching", url)
		feed, err := f.fetcher.Fetch(ctx, url)
		if err != nil {
			return err
		}

		if err := f.handleImages(ctx, &feed); err != nil {
			return err
		}

		if err := f.store.SaveFeed(ctx, feed); err != nil {
			return err
		}
	}

	return nil
}

func (f UpdateFeeds) handleImages(ctx context.Context, feed *rss.Feed) error {
	for _, article := range feed.Articles {
		if article.Image == "" {
			continue
		}

		resp, err := f.client.Get(article.Image)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch image: %s", resp.Status)
		}

		if err := f.storage.Create(ctx, article.ImagePath(), resp.Body); err != nil {
			return err
		}
	}

	return nil
}
