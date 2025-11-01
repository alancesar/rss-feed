package usecase

import (
	"context"
	"rss-summary/pkg/rss"
)

type (
	FeedStore interface {
		SaveFeed(context.Context, rss.Feed) error
	}

	AddFeed struct {
		store   FeedStore
		fetcher FeedFetcher
	}
)

func NewAddFeed(store FeedStore, fetcher FeedFetcher) *AddFeed {
	return &AddFeed{
		store:   store,
		fetcher: fetcher,
	}
}

func (uc AddFeed) Execute(ctx context.Context, url string) (rss.Feed, error) {
	feed, err := uc.fetcher.Fetch(ctx, url)
	if err != nil {
		return rss.Feed{}, err
	}

	if err := uc.store.SaveFeed(ctx, feed); err != nil {
		return rss.Feed{}, err
	}

	return feed, nil
}
