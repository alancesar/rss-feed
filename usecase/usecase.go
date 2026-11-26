// Package usecase holds the application's business logic. Each use case
// orchestrates domain entities through the small ports (interfaces) declared
// alongside it, keeping the logic independent from any concrete infrastructure.
package usecase

import (
	"context"
	"rss-feed/pkg/event"
)

type (
	// Subscriber receives messages published to a topic.
	Subscriber interface {
		// Subscribe starts listening to a topic and returns a channel of deliveries.
		// Each delivery must be acknowledged or rejected by the consumer.
		Subscribe(ctx context.Context, topic event.Topic) (<-chan event.Delivery, error)
		// Close stops the subscription and releases its resources.
		Close() error
	}

	// Publisher sends messages to a topic so interested consumers can react to them.
	Publisher interface {
		// Publish delivers a message to the topic it is addressed to.
		Publish(ctx context.Context, e event.Message) error
	}

	// Broker is both a Publisher and a Subscriber.
	Broker interface {
		Publisher
		Subscriber
	}

	// FeedFetcher retrieves and parses a remote feed.
	FeedFetcher interface {
		// Fetch downloads the feed at the given URL and returns it along with its articles.
		Fetch(context.Context, string) (event.Feed, error)
	}
)
