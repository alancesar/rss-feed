package usecase

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	ArticleStore interface {
		SaveArticle(context.Context, string, rss.Article) error
	}

	ConsumeFeed struct {
		subscriber Subscriber
		store      ArticleStore
		publisher  Publisher
	}
)

func NewConsumeFeed(store ArticleStore, subscriber Subscriber, publisher Publisher) *ConsumeFeed {
	return &ConsumeFeed{
		subscriber: subscriber,
		store:      store,
		publisher:  publisher,
	}
}

func (uc ConsumeFeed) Execute(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("starting consume articles")

	defer func() {
		_ = uc.subscriber.Close()
		logger.Info().Msg("finished consume articles")
	}()

	deliveries, err := uc.subscriber.Subscribe(ctx, "rss.feed.article.found")
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

		feed := e.ToDomain()

		logger.Info().Str("feed", feed.Name).Int("articles", len(feed.Articles)).Msg("saving feed")
		for _, article := range e.Articles {
			if err := uc.store.SaveArticle(ctx, feed.ID, article.ToDomain()); err != nil {
				logger.Error().Err(err).Msg("failed to save article")
				continue
			}

			if article.Image.ImageID == "" {
				continue
			}

			logger.Info().Str("article_id", article.ArticleID).Msg("publishing image event")
			if err := uc.publisher.Publish(ctx, "feed.article.image.found", event.Message{
				Payload: event.NewPayload(article.Image),
			}); err != nil {
				continue
			}
		}

		delivery.Ack()
	}

	return nil
}
