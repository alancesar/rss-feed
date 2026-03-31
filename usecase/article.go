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
		CreateArticle(ctx context.Context, article rss.Article) error
	}

	Article struct {
		subscriber Subscriber
		store      ArticleStore
	}
)

func NewArticle(subscriber Subscriber, store ArticleStore) *Article {
	return &Article{
		subscriber: subscriber,
		store:      store,
	}
}

func (uc *Article) Execute(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("starting consume articles")

	deliveries, err := uc.subscriber.Subscribe(ctx, event.TopicFeedArticleFound)
	if err != nil {
		logger.Error().Err(err).Msg("failed to subscribe to articles")
		return err
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("stopping consume articles")
			return ctx.Err()
		case delivery := <-deliveries:
			var e event.Article
			if err := json.Unmarshal(delivery.Payload, &e); err != nil {
				delivery.Nack(false)
				continue
			}

			article := e.ToDomain()
			logger.Info().Str("article_id", article.ID).Msg("saving article to vector database")
			if err := uc.store.CreateArticle(ctx, article); err != nil {
				logger.Error().Err(err).Msg("failed to store article")
				delivery.Nack(true)
				continue
			}

			delivery.Ack()
		}
	}
}
