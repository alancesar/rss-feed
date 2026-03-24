package usecase

import (
	"context"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	FeedsStore interface {
		GetAllFeeds(ctx context.Context) ([]rss.Feed, error)
		UpdateFeed(ctx context.Context, feed rss.Feed) error
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
	feeds, err := uc.store.GetAllFeeds(ctx)
	if err != nil {
		return err
	}

	log := zerolog.Ctx(ctx)
	for _, feed := range feeds {
		log.Info().Str("url", feed.URL).Msg("fetching feed")
		fetchedFeed, err := uc.fetcher.Fetch(ctx, feed.URL)
		if err != nil {
			log.Error().Err(err).Str("url", feed.URL).Msg("failed to fetch feed")
			continue
		}

		log.Info().Str("feed", fetchedFeed.Name).Int("articles", len(fetchedFeed.Articles)).Msg("publishing feed.article.found event")
		if err := uc.publisher.Publish(ctx, "feed.article.found", event.Message{
			Payload: fetchedFeed,
		}); err != nil {
			return err
		}

		feed.Touch()
		if err := uc.store.UpdateFeed(ctx, feed); err != nil {
			return err
		}

		log.Info().Str("name", fetchedFeed.Name).Str("url", feed.URL).Msg("feed fetched")
	}

	return nil
}
