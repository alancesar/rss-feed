package usecase

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"

	"github.com/rs/zerolog"
)

type (
	HandleFeed struct {
		broker Broker
	}
)

func NewHandleFeed(broker Broker) *HandleFeed {
	return &HandleFeed{
		broker: broker,
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

		for _, article := range e.Articles {
			logger.Info().Str("article_id", article.ArticleID).Msg("publishing article event")

			if err := uc.broker.Publish(ctx, event.NewArticleFound(article)); err != nil {
				logger.Error().Err(err).Msg("failed to publish article")
			}
		}

		delivery.Ack()
	}

	return nil
}
