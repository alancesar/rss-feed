package usecase

import (
	"context"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	FeedFetcher interface {
		Fetch(context.Context, string) (event.Feed, error)
	}

	FeedStore interface {
		SaveFeed(context.Context, rss.Feed) error
	}

	PublishFeed struct {
		fetcher   FeedFetcher
		store     FeedStore
		publisher Publisher
	}
)

func NewPublishFeed(fetcher FeedFetcher, store FeedStore, publisher Publisher) *PublishFeed {
	return &PublishFeed{
		publisher: publisher,
		store:     store,
		fetcher:   fetcher,
	}
}

func (uc PublishFeed) Execute(ctx context.Context, url string) (rss.Feed, error) {
	log := zerolog.Ctx(ctx)

	log.Info().Str("url", url).Msg("fetching feed")
	fetchedFeed, err := uc.fetcher.Fetch(ctx, url)
	if err != nil {
		return rss.Feed{}, err
	}

	feed := rss.Feed{
		ID:   fetchedFeed.FeedID,
		Name: fetchedFeed.Name,
		URL:  fetchedFeed.URL,
	}

	log.Info().Str("feed", fetchedFeed.Name).Msg("saving feed")
	if err := uc.store.SaveFeed(ctx, feed); err != nil {
		return rss.Feed{}, err
	}

	log.Info().Str("feed", fetchedFeed.Name).Int("articles", len(fetchedFeed.Articles)).Msg("publishing feed.found event")
	if err := uc.publisher.Publish(ctx, "feed.found", event.Event{
		Payload: fetchedFeed,
	}); err != nil {
		return rss.Feed{}, err
	}

	return feed, nil
}
