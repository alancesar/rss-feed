package usecase

import (
	"context"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
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

	log := zerolog.Ctx(ctx)
	for _, url := range urls {
		log.Info().Str("url", url).Msg("fetching feed")
		f, err := uc.handler(ctx, url)
		if err != nil {
			return err
		}

		log.Info().Str("name", f.Name).Str("url", url).Msg("feed fetched")
	}

	return nil
}
