package usecase

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	// ArticleStore indexes articles so they can later be found by meaning.
	ArticleStore interface {
		// CreateArticle vectorizes the article text and stores it in the search index.
		CreateArticle(ctx context.Context, article rss.Article) error
	}

	// Article is a consumer that indexes every discovered article for semantic search.
	Article struct {
		subscriber Subscriber
		store      ArticleStore
	}
)

// NewArticle creates an Article use case that listens through subscriber and indexes into store.
func NewArticle(subscriber Subscriber, store ArticleStore) *Article {
	return &Article{
		subscriber: subscriber,
		store:      store,
	}
}

// Execute consumes "article found" events and indexes each article for semantic search.
// Messages that cannot be decoded are rejected with a request not to redeliver them, and
// indexing failures are rejected with a request to redeliver them; whether the broker honours
// that request is up to the broker. It blocks until ctx is cancelled (returning ctx.Err())
// or the subscription is closed (returning nil), and closes the subscriber on exit.
func (uc *Article) Execute(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("starting consume articles")

	defer func() {
		_ = uc.subscriber.Close()
		logger.Info().Msg("finished consume articles")
	}()

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
		case delivery, ok := <-deliveries:
			if !ok {
				logger.Info().Msg("articles subscription closed")
				return nil
			}

			var e event.Article
			if err := json.Unmarshal(delivery.Payload, &e); err != nil {
				logger.Error().Err(err).Msg("failed to unmarshal article")
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
