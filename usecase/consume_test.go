package usecase_test

import (
	"context"
	"testing"
	"time"

	"rss-feed/pkg/event"
	"rss-feed/usecase"
)

func TestConsumeFeed_Execute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := newTestDB(t)
	createTestFeed(t, ctx, db)

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

	uc := usecase.NewConsumeFeed(db, newTestBroker(t, "rss.feed.article.found", feedEvent))
	go func() { _ = uc.Execute(ctx) }()

	awaitCondition(t, ctx, func() bool {
		articles, _ := db.GetArticlesFromDate(ctx, now)
		return len(articles) == 2
	})

	articles, err := db.GetArticlesFromDate(ctx, now)
	if err != nil {
		t.Fatalf("getting articles: %v", err)
	}
	if len(articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(articles))
	}
}
