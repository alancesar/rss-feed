package usecase

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"
	"time"

	"github.com/rs/zerolog"
)

type (
	// FeedsStore reads and updates the registered feeds.
	FeedsStore interface {
		// GetAllFeeds returns every registered feed.
		GetAllFeeds(ctx context.Context) ([]rss.Feed, error)
		// UpdateFeed persists the changes made to the feed.
		UpdateFeed(ctx context.Context, feed rss.Feed) error
	}

	// UpdateFeeds keeps the registered feeds fresh, refreshing them periodically and on demand.
	UpdateFeeds struct {
		store   FeedsStore
		fetcher FeedFetcher
		broker  Broker
	}
)

// NewUpdateFeeds creates an UpdateFeeds use case that reads feeds from store, fetches them with
// fetcher and publishes and listens for refresh jobs through broker.
func NewUpdateFeeds(store FeedsStore, fetcher FeedFetcher, broker Broker) *UpdateFeeds {
	return &UpdateFeeds{
		store:   store,
		fetcher: fetcher,
		broker:  broker,
	}
}

// Execute refreshes every feed once immediately, then again on every updateInterval and
// whenever a refresh job arrives (which also restarts the interval). A failure in the initial
// refresh is returned; later failures are logged and the loop continues. It runs until ctx is
// cancelled (returning ctx.Err()) or the subscription is closed (returning nil).
func (uc UpdateFeeds) Execute(ctx context.Context, updateInterval time.Duration) error {
	if err := uc.updateFeeds(ctx); err != nil {
		return err
	}

	logger := zerolog.Ctx(ctx)
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	deliveries, err := uc.broker.Subscribe(ctx, event.TopicFeedJobs)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("stopping rss.feed.jobs consumer")
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}
			var j event.Job
			if err := json.Unmarshal(delivery.Payload, &j); err != nil {
				logger.Error().Err(err).Msg("unmarshalling rss.feed.job event")
				delivery.Nack(false)
				continue
			}

			logger.Info().Msg("forced feed update triggered")
			if err := uc.updateFeeds(ctx); err != nil {
				logger.Error().Err(err).Msg("updating feeds")
				delivery.Nack(false)
				continue
			}

			ticker.Reset(updateInterval)
			delivery.Ack()
		case <-ticker.C:
			if err := uc.updateFeeds(ctx); err != nil {
				logger.Error().Err(err).Msg("updating feeds")
			}
		}
	}
}

// updateFeeds fetches every registered feed. Feeds that fail to fetch, or that have not changed
// since they were last checked, are skipped. Changed feeds are announced with a "feed found"
// event and then marked as checked. A publish or persistence failure aborts the run.
func (uc UpdateFeeds) updateFeeds(ctx context.Context) error {
	feeds, err := uc.store.GetAllFeeds(ctx)
	if err != nil {
		return err
	}

	logger := zerolog.Ctx(ctx)
	for _, feed := range feeds {
		logger.Info().Str("url", feed.URL).Msg("fetching feed")
		fetchedFeed, err := uc.fetcher.Fetch(ctx, feed.URL)
		if err != nil {
			logger.Error().Err(err).Str("url", feed.URL).Msg("failed to fetch feed")
			continue
		}
		if fetchedFeed.UpdatedAt != nil && feed.UpdatedAt.After(*fetchedFeed.UpdatedAt) {
			logger.Info().Str("name", fetchedFeed.Name).Str("url", feed.URL).Msg("feed not updated, skipping")
			continue
		}

		logger.Info().Str("feed", fetchedFeed.Name).Int("articles", len(fetchedFeed.Articles)).Msg("publishing rss.feed.found event")
		if err := uc.broker.Publish(ctx, event.NewFeedFoundEvent(fetchedFeed)); err != nil {
			return err
		}

		feed.Touch()
		if err := uc.store.UpdateFeed(ctx, feed); err != nil {
			return err
		}

		logger.Info().Str("name", fetchedFeed.Name).Str("url", feed.URL).Msg("feed fetched")
	}

	return nil
}
