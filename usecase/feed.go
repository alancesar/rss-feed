package usecase

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	// FeedArticleStore persists articles.
	FeedArticleStore interface {
		// SaveArticle persists the article.
		SaveArticle(ctx context.Context, article rss.Article) error
	}

	// HandleFeed is a consumer that persists the articles of every discovered feed
	// and announces each one so further processing can happen independently.
	HandleFeed struct {
		broker Broker
		store  FeedArticleStore
	}
)

// NewHandleFeed creates a HandleFeed use case that consumes and publishes through broker and persists into store.
func NewHandleFeed(broker Broker, store FeedArticleStore) *HandleFeed {
	return &HandleFeed{
		broker: broker,
		store:  store,
	}
}

// Execute consumes "feed found" events. For each article in the feed it persists the
// article and then publishes an "article found" event. On the first failure the whole
// feed message is rejected with a request to redeliver it; messages that cannot be decoded are
// rejected with a request not to redeliver them. Whether the broker honours that request is up
// to the broker. It runs until the subscription is closed and closes the broker on exit.
func (uc HandleFeed) Execute(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("starting consume feed")

	defer func() {
		_ = uc.broker.Close()
		logger.Info().Msg("finished consume feed")
	}()

	deliveries, err := uc.broker.Subscribe(ctx, event.TopicFeedFound)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		var e event.Feed
		if err := json.Unmarshal(delivery.Payload, &e); err != nil {
			logger.Error().Err(err).Msg("failed to unmarshal feed")
			delivery.Nack(false)
			continue
		}

		failed := false
		for _, article := range e.Articles {
			if err := uc.store.SaveArticle(ctx, article.ToDomain()); err != nil {
				logger.Error().Err(err).Msg("failed to save article")
				failed = true
				break
			}

			logger.Info().Str("article_id", article.ArticleID).Msg("publishing article event")

			if err := uc.broker.Publish(ctx, event.NewArticleFound(article)); err != nil {
				logger.Error().Err(err).Msg("failed to publish article")
				failed = true
				break
			}
		}

		if failed {
			delivery.Nack(true)
		} else {
			delivery.Ack()
		}
	}

	return nil
}
