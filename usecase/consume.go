package usecase

import (
	"context"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	FeedStore interface {
		SaveFeed(context.Context, rss.Feed) error
	}

	ConsumeFeed struct {
		store     FeedStore
		publisher Publisher
	}
)

func NewConsumeFeed(store FeedStore, publisher Publisher) *ConsumeFeed {
	return &ConsumeFeed{
		store:     store,
		publisher: publisher,
	}
}

func (uc ConsumeFeed) Execute(ctx context.Context, e event.Feed) error {
	log := zerolog.Ctx(ctx)

	articles := make([]rss.Article, len(e.Articles))
	for i, article := range e.Articles {
		articles[i] = rss.Article{
			ID:          article.ArticleID,
			Title:       article.Title,
			URL:         article.URL,
			PublishedAt: article.PublishedAt,
		}
	}

	feed := rss.Feed{
		ID:       e.FeedID,
		Name:     e.Name,
		URL:      e.URL,
		Articles: articles,
	}

	log.Info().Str("feed", feed.Name).Int("articles", len(articles)).Msg("saving feed")
	if err := uc.store.SaveFeed(ctx, feed); err != nil {
		return err
	}

	for _, article := range e.Articles {
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
