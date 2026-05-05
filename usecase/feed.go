package usecase

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	FeedArticleStore interface {
		SaveArticle(ctx context.Context, article rss.Article) error
	}

	HandleFeed struct {
		broker Broker
		store  FeedArticleStore
	}
)

func NewHandleFeed(broker Broker, store FeedArticleStore) *HandleFeed {
	return &HandleFeed{
		broker: broker,
		store:  store,
	}
}

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
			logger.Info().Str("article_id", article.ArticleID).Msg("publishing article event")

			if err := uc.store.SaveArticle(ctx, article.ToDomain()); err != nil {
				logger.Error().Err(err).Msg("failed to save article")
				failed = true
				break
			}

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
