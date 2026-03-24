package usecase

import (
	"context"
	"rss-feed/pkg/event"
)

type (
	Subscriber interface {
		Subscribe(ctx context.Context, queue string) (<-chan event.Delivery, error)
		Close() error
	}

	Publisher interface {
		Publish(ctx context.Context, topic string, e event.Message) error
	}

	FeedFetcher interface {
		Fetch(context.Context, string) (event.Feed, error)
	}
)
