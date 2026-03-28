package usecase

import (
	"context"
	"fmt"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	FeedStore interface {
		GetFeedByURL(context.Context, string) (rss.Feed, bool, error)
		CreateFeed(context.Context, rss.Feed) error
	}

	SaveFeed struct {
		fetcher   FeedFetcher
		store     FeedStore
		publisher Publisher
	}
)

func NewSaveFeed(fetcher FeedFetcher, store FeedStore, publisher Publisher) *SaveFeed {
	return &SaveFeed{
		publisher: publisher,
		store:     store,
		fetcher:   fetcher,
	}
}

func (uc SaveFeed) Execute(ctx context.Context, url string) (rss.Feed, error) {
	log := zerolog.Ctx(ctx)

	if _, exists, err := uc.store.GetFeedByURL(ctx, url); err != nil {
		return rss.Feed{}, err
	} else if exists {
		log.Info().Str("url", url).Msg("feed already exists, skipping")
		return rss.Feed{}, fmt.Errorf("%w: %s", rss.ErrFeedAlreadyExists, url)
	}

	log.Info().Str("url", url).Msg("fetching feed")
	fetchedFeed, err := uc.fetcher.Fetch(ctx, url)
	if err != nil {
		return rss.Feed{}, err
	}

	feed := fetchedFeed.ToDomain()

	log.Info().Str("feed", fetchedFeed.Name).Msg("saving feed")
	if err := uc.store.CreateFeed(ctx, feed); err != nil {
		return rss.Feed{}, err
	}

	log.Info().Str("feed", fetchedFeed.Name).Int("articles", len(fetchedFeed.Articles)).Msg("publishing feed.article.found event")
	if err := uc.publisher.Publish(ctx, event.NewArticleFoundEvent(fetchedFeed)); err != nil {
		return rss.Feed{}, err
	}

	return feed, nil
}
