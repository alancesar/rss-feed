package usecase

import (
	"context"
	"fmt"
	"rss-summary/pkg/rss"
)

type (
	FeedsStore interface {
		GetAllFeedURLs(ctx context.Context) ([]string, error)
	}

	FeedHandlerFn func(context.Context, string) (rss.Feed, error)

	UpdateFeeds struct {
		store   FeedsStore
		handler FeedHandlerFn
	}
)

func NewUpdateFeeds(store FeedsStore, fn FeedHandlerFn) *UpdateFeeds {
	return &UpdateFeeds{
		store:   store,
		handler: fn,
	}
}

func (uc UpdateFeeds) Execute(ctx context.Context) error {
	urls, err := uc.store.GetAllFeedURLs(ctx)
	if err != nil {
		return err
	}

	for _, url := range urls {
		fmt.Println("fetching", url)
		f, err := uc.handler(ctx, url)
		if err != nil {
			return err
		}

		fmt.Println("fetched", f.Name, "from", url)
	}

	return nil
}
