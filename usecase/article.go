package usecase

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"

	"github.com/rs/zerolog"
)

type (
	HandleArticles struct {
		broker Broker
	}
)

func NewHandleArticles(broker Broker) *HandleArticles {
	return &HandleArticles{
		broker: broker,
	}
}

func (uc HandleArticles) Execute(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("starting consume articles")

	defer func() {
		_ = uc.broker.Close()
		logger.Info().Msg("finished consume articles")
	}()

	deliveries, err := uc.broker.Subscribe(ctx, event.TopicFeedArticleFound)
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
			if article.Image.ImageID == "" {
				continue
			}

			logger.Info().Str("article_id", article.ArticleID).Msg("publishing image event")
			if err := uc.broker.Publish(ctx, event.NewImageFoundEvent(article.Image)); err != nil {
				continue
			}
		}

		delivery.Ack()
	}

	return nil
}
