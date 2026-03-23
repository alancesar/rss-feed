package usecase_test

import (
	"context"
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
