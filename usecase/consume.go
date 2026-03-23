package usecase

import (
	"context"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	ArticleStore interface {
		SaveArticle(context.Context, string, rss.Article) error
	}

	ConsumeFeed struct {
		store     ArticleStore
		publisher Publisher
	}
)

func NewConsumeFeed(store ArticleStore, publisher Publisher) *ConsumeFeed {
	return &ConsumeFeed{
		store:     store,
		publisher: publisher,
	}
}

func (uc ConsumeFeed) Execute(ctx context.Context, e event.Feed) error {
	log := zerolog.Ctx(ctx)
	feed := e.ToDomain()

	log.Info().Str("feed", feed.Name).Int("articles", len(feed.Articles)).Msg("saving feed")
	for _, article := range e.Articles {
		if err := uc.store.SaveArticle(ctx, feed.ID, article.ToDomain()); err != nil {
			return err
		}

		if article.Image.ImageID == "" {
			continue
		}

		log.Info().Str("article_id", article.ArticleID).Msg("publishing image event")
		if err := uc.publisher.Publish(ctx, "feed.article.image.found", event.Event{
			Payload: article.Image,
		}); err != nil {
			return err
		}
	}

	return nil
}
