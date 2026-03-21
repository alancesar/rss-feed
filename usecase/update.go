package usecase

import (
	"context"
	"fmt"
)

type (
	FeedsStore interface {
		FeedStore
		GetAllFeedURLs(ctx context.Context) ([]string, error)
	}

	UpdateFeeds struct {
		store   FeedsStore
		fetcher FeedFetcher
	}
)

func NewUpdateFeeds(store FeedsStore, fetcher FeedFetcher) *UpdateFeeds {
	return &UpdateFeeds{
		store:   store,
		fetcher: fetcher,
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

		if err := f.store.SaveFeed(ctx, feed); err != nil {
			return err
		}
	}

	return nil
}
