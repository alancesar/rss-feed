package usecase

import (
	"context"
	"rss-feed/pkg/event"

	"github.com/rs/zerolog"
)

type (
	FeedsStore interface {
		GetAllFeedURLs(ctx context.Context) ([]string, error)
	}

	UpdateFeeds struct {
		store     FeedsStore
		fetcher   FeedFetcher
		publisher Publisher
	}
)

func NewUpdateFeeds(store FeedsStore, fetcher FeedFetcher, publisher Publisher) *UpdateFeeds {
	return &UpdateFeeds{
		store:     store,
		fetcher:   fetcher,
		publisher: publisher,
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
		feed, err := uc.fetcher.Fetch(ctx, url)
		if err != nil {
			log.Error().Err(err).Str("url", url).Msg("failed to fetch feed")
			continue
		}

		log.Info().Str("feed", feed.Name).Int("articles", len(feed.Articles)).Msg("publishing feed.article.found event")
		if err := uc.publisher.Publish(ctx, "feed.article.found", event.Event{
			Payload: feed,
		}); err != nil {
			return err
		}

		log.Info().Str("name", feed.Name).Str("url", url).Msg("feed fetched")
	}

	return nil
}
