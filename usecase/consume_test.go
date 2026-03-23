package usecase_test

import (
	"context"
	"testing"

	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"
	"rss-feed/usecase"
)

func TestConsumeFeed_Execute(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	if err := db.CreateFeed(ctx, rss.Feed{ID: "feed-abc", Name: "Test Blog", URL: "https://example.com/feed.xml"}); err != nil {
		t.Fatalf("creating feed: %v", err)
	}

	now := today()
	feedEvent := event.Feed{
		FeedID: "feed-abc",
		Name:   "Test Blog",
		URL:    "https://example.com/feed.xml",
		Articles: []event.Article{
			{
				ArticleID:   "article-1",
				Title:       "First Post",
				URL:         "https://example.com/post-1",
				PublishedAt: now,
			},
			{
				ArticleID:   "article-2",
				Title:       "Second Post",
				URL:         "https://example.com/post-2",
				PublishedAt: now,
				Image:       event.Image{ImageID: "img-1", ArticleID: "article-2", URL: "https://example.com/img.jpg"},
			},
		},
	}

	uc := usecase.NewConsumeFeed(db, &mockPublisher{})
	if err := uc.Execute(ctx, feedEvent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	articles, err := db.GetArticlesFromDate(ctx, now)
	if err != nil {
		t.Fatalf("getting articles: %v", err)
	}
	if len(articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(articles))
	}
}
