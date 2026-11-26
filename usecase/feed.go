package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

var (
	// ErrInvalidFeed is returned when a feed event cannot be decoded.
	ErrInvalidFeed = errors.New("invalid feed")
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
// article and then publishes an "article found" event. A message that cannot be decoded
// (ErrInvalidFeed) is rejected with a request not to redeliver it. Any other failure, such as
// an article that cannot be saved or published, rejects the whole feed message with a request
// to redeliver it. Whether the broker honours either request is up to the broker. It runs
// until the subscription is closed and closes the broker on exit.
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
		if err := uc.handleFeed(ctx, delivery.Payload); err != nil {
			if errors.Is(err, ErrInvalidFeed) {
				logger.Error().Err(err).Msg("discarding invalid feed")
				delivery.Nack(false)
				continue
			}

			logger.Error().Err(err).Msg("failed to handle feed")
			delivery.Nack(true)
			continue
		}

		delivery.Ack()
	}

	return nil
}

// handleFeed decodes the feed event and handles each of its articles, stopping at the
// first failure. A payload that cannot be decoded yields ErrInvalidFeed.
func (uc HandleFeed) handleFeed(ctx context.Context, payload event.Payload) error {
	var e event.Feed
	if err := json.Unmarshal(payload, &e); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidFeed, err)
	}

	for _, article := range e.Articles {
		if err := uc.handleArticle(ctx, article); err != nil {
			return err
		}
	}

	return nil
}

// handleArticle persists the article and publishes an "article found" event for it.
func (uc HandleFeed) handleArticle(ctx context.Context, article event.Article) error {
	if err := uc.store.SaveArticle(ctx, article.ToDomain()); err != nil {
		return fmt.Errorf("save article %s: %w", article.ArticleID, err)
	}

	zerolog.Ctx(ctx).Info().Str("article_id", article.ArticleID).Msg("publishing article event")

	if err := uc.broker.Publish(ctx, event.NewArticleFound(article)); err != nil {
		return fmt.Errorf("publish article %s: %w", article.ArticleID, err)
	}

	return nil
}
