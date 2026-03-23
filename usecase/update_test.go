package usecase_test

import (
	"context"
	"errors"
	"testing"

	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"
	"rss-feed/usecase"
)

func TestUpdateFeeds_Execute(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	for _, f := range []rss.Feed{
		{ID: "feed-1", Name: "Blog One", URL: "https://blog1.com/feed.xml"},
		{ID: "feed-2", Name: "Blog Two", URL: "https://blog2.com/feed.xml"},
	} {
		if err := db.CreateFeed(ctx, f); err != nil {
			t.Fatalf("creating feed: %v", err)
		}
	}

	fetcher := &recordingFetcher{
		feed: event.Feed{FeedID: "x", Name: "Feed"},
	}

	uc := usecase.NewUpdateFeeds(db, fetcher, &mockPublisher{})
	if err := uc.Execute(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fetcher.calls) != 2 {
		t.Errorf("expected fetcher to be called 2 times, got %d", len(fetcher.calls))
	}
}

func TestUpdateFeeds_Execute_SkipsFailedFeeds(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	for _, f := range []rss.Feed{
		{ID: "feed-1", Name: "Blog One", URL: "https://blog1.com/feed.xml"},
		{ID: "feed-2", Name: "Blog Two", URL: "https://blog2.com/feed.xml"},
	} {
		if err := db.CreateFeed(ctx, f); err != nil {
			t.Fatalf("creating feed: %v", err)
		}
	}

	var successfulURLs []string
	fetcher := &funcFetcher{fn: func(url string) (event.Feed, error) {
		if url == "https://blog1.com/feed.xml" {
			return event.Feed{}, errors.New("connection refused")
		}
		successfulURLs = append(successfulURLs, url)
		return event.Feed{FeedID: "x", Name: "Feed", URL: url}, nil
	}}

	uc := usecase.NewUpdateFeeds(db, fetcher, &mockPublisher{})
	if err := uc.Execute(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(successfulURLs) != 1 {
		t.Errorf("expected 1 successful fetch, got %d", len(successfulURLs))
	}
	if successfulURLs[0] != "https://blog2.com/feed.xml" {
		t.Errorf("expected successful URL %q, got %q", "https://blog2.com/feed.xml", successfulURLs[0])
	}
}
