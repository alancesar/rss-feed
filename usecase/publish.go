package usecase

import (
	"context"
	"rss-summary/pkg/event"
	"rss-summary/pkg/rss"
)

type (
	FeedFetcher interface {
		Fetch(context.Context, string) (event.Feed, error)
	}

	PublishFeed struct {
		publisher Publisher
		fetcher   FeedFetcher
	}
)

func NewPublishFeed(publisher Publisher, fetcher FeedFetcher) *PublishFeed {
	return &PublishFeed{
		publisher: publisher,
		fetcher:   fetcher,
	}
}

func (uc PublishFeed) Execute(ctx context.Context, url string) (rss.Feed, error) {
	feed, err := uc.fetcher.Fetch(ctx, url)
	if err != nil {
		return rss.Feed{}, err
	}

	if err := uc.publisher.Publish(ctx, "feed.found", event.Event{
		Payload: feed,
	}); err != nil {
		return rss.Feed{}, err
	}

	return rss.Feed{
		ID:   feed.FeedID,
		Name: feed.Name,
		URL:  feed.URL,
	}, nil
}
