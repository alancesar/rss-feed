package usecase

import (
	"context"
	"errors"
	"fmt"
	"rss-summary/pkg/rss"
)

var (
	ErrSourceAlreadyExists = errors.New("source already exists")
)

type (
	FeedStore interface {
		SaveFeed(context.Context, rss.Feed) error
		GetFeedByURL(context.Context, string) (rss.Feed, bool, error)
	}

	FeedFetcher interface {
		Fetch(context.Context, string) (rss.Feed, error)
	}
	AddSource struct {
		store   FeedStore
		fetcher FeedFetcher
	}
)

func NewAddSource(store FeedStore, fetcher FeedFetcher) *AddSource {
	return &AddSource{
		store:   store,
		fetcher: fetcher,
	}
}

func (uc AddSource) Execute(ctx context.Context, url string) error {
	_, exists, err := uc.store.GetFeedByURL(ctx, url)
	if err != nil {
		return fmt.Errorf("failed retrieve source from database: %w", err)
	}

	if exists {
		return fmt.Errorf("%w: %s", ErrSourceAlreadyExists, url)
	}

	source, err := uc.fetcher.Fetch(ctx, url)
	if err != nil {
		return err
	}

	if err := uc.store.SaveFeed(ctx, source); err != nil {
		return err
	}

	return nil
}
