package usecase

import (
	"context"
	"rss-summary/pkg/event"
)

type (
	Publisher interface {
		Publish(ctx context.Context, topic string, e event.Event) error
	}
)
