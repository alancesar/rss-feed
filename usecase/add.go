package usecase

import (
	"context"
	"errors"
	"fmt"
	"rss-summary/pkg/rss"
)

var (
	ErrFeedAlreadyExists = errors.New("feed already exists")
)

type (
	FeedStore interface {
		SaveFeed(context.Context, rss.Feed) error
		GetFeedByURL(context.Context, string) (rss.Feed, bool, error)
	}

	FeedFetcher interface {
		Fetch(context.Context, string) (rss.Feed, error)
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

func (uc AddFeed) Execute(ctx context.Context, url string) error {
	_, exists, err := uc.store.GetFeedByURL(ctx, url)
	if err != nil {
		return fmt.Errorf("failed retrieve feed from database: %w", err)
	}

	if exists {
		return fmt.Errorf("%w: %s", ErrFeedAlreadyExists, url)
	}

	feed, err := uc.fetcher.Fetch(ctx, url)
	if err != nil {
		return err
	}

	if err := uc.store.SaveFeed(ctx, feed); err != nil {
		return err
	}

	return nil
}
