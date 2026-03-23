package usecase

import (
	"context"
	"rss-feed/pkg/event"
)

type (
	Publisher interface {
		Publish(ctx context.Context, topic string, e event.Event) error
	}

	FeedFetcher interface {
		Fetch(context.Context, string) (event.Feed, error)
	}
)
