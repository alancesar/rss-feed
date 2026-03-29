package usecase

import (
	"context"
	"rss-feed/pkg/event"
)

type (
	Subscriber interface {
		Subscribe(ctx context.Context, topic event.Topic) (<-chan event.Delivery, error)
		Close() error
	}

	Publisher interface {
		Publish(ctx context.Context, e event.Message) error
	}

	Broker interface {
		Publisher
		Subscriber
	}

	FeedFetcher interface {
		Fetch(context.Context, string) (event.Feed, error)
	}
)
