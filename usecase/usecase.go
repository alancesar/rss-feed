package usecase

import (
	"context"
	"rss-summary/pkg/rss"
)

type (
	FeedFetcher interface {
		Fetch(context.Context, string) (rss.Feed, error)
	}
)
