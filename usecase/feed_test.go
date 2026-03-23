package usecase_test

import (
	"context"
	"testing"
	"time"

	"rss-feed/pkg/event"
	"rss-feed/usecase"
)

func TestSaveFeed_Execute(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	fetcher := &mockFetcher{
		feed: event.Feed{
			FeedID: "feed-abc",
			Name:   "Test Blog",
			URL:    "https://example.com/feed.xml",
			Articles: []event.Article{
				{ArticleID: "article-1", Title: "First Post", URL: "https://example.com/post-1", PublishedAt: time.Now()},
				{ArticleID: "article-2", Title: "Second Post", URL: "https://example.com/post-2", PublishedAt: time.Now()},
			},
		},
	}

	uc := usecase.NewSaveFeed(fetcher, db, &mockPublisher{})
	feed, err := uc.Execute(ctx, "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if feed.Name != "Test Blog" {
		t.Errorf("expected feed name %q, got %q", "Test Blog", feed.Name)
	}

	saved, exists, err := db.GetFeedByURL(ctx, "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("getting feed from db: %v", err)
	}
	if !exists {
		t.Fatal("expected feed to be saved in database")
	}
	if saved.ID != "feed-abc" {
		t.Errorf("expected feed ID %q, got %q", "feed-abc", saved.ID)
	}
}
