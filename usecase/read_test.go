package usecase_test

import (
	"context"
	"testing"

	"rss-feed/pkg/rss"
	"rss-feed/usecase"
)

func TestReadArticles_Execute(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	now := today()

	if err := db.CreateFeed(ctx, rss.Feed{ID: "feed-abc", Name: "Test Blog", URL: "https://example.com/feed.xml"}); err != nil {
		t.Fatalf("creating feed: %v", err)
	}
	if err := db.SaveArticle(ctx, "feed-abc", rss.Article{
		ID:          "article-1",
		Title:       "First Post",
		URL:         "https://example.com/post-1",
		PublishedAt: now,
	}); err != nil {
		t.Fatalf("saving article: %v", err)
	}
	img := rss.NewImage("images/original/article-1.jpg", rss.OriginalImageType)
	if err := db.SaveImage(ctx, "article-1", img); err != nil {
		t.Fatalf("saving image: %v", err)
	}

	uc := usecase.NewReadArticles(db, &mockImageStorage{})
	articles, err := uc.Execute(ctx, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}
	if len(articles[0].Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(articles[0].Images))
	}

	wantURL := "https://cdn.example.com/images/original/article-1.jpg"
	if articles[0].Images[0].Path != wantURL {
		t.Errorf("expected presigned URL %q, got %q", wantURL, articles[0].Images[0].Path)
	}
}

func TestReadArticles_Execute_NoImages(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	now := today()

	if err := db.CreateFeed(ctx, rss.Feed{ID: "feed-abc", Name: "Test Blog", URL: "https://example.com/feed.xml"}); err != nil {
		t.Fatalf("creating feed: %v", err)
	}
	if err := db.SaveArticle(ctx, "feed-abc", rss.Article{
		ID:          "article-1",
		Title:       "First Post",
		URL:         "https://example.com/post-1",
		PublishedAt: now,
	}); err != nil {
		t.Fatalf("saving article: %v", err)
	}

	uc := usecase.NewReadArticles(db, &mockImageStorage{})
	articles, err := uc.Execute(ctx, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}
	if len(articles[0].Images) != 0 {
		t.Errorf("expected 0 images, got %d", len(articles[0].Images))
	}
}

func TestReadArticles_Execute_DateFiltering(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	now := today()
	yesterday := now.AddDate(0, 0, -1)

	if err := db.CreateFeed(ctx, rss.Feed{ID: "feed-abc", Name: "Test Blog", URL: "https://example.com/feed.xml"}); err != nil {
		t.Fatalf("creating feed: %v", err)
	}
	if err := db.SaveArticle(ctx, "feed-abc", rss.Article{
		ID:          "article-today",
		Title:       "Today Post",
		URL:         "https://example.com/today",
		PublishedAt: now,
	}); err != nil {
		t.Fatalf("saving article: %v", err)
	}
	if err := db.SaveArticle(ctx, "feed-abc", rss.Article{
		ID:          "article-yesterday",
		Title:       "Yesterday Post",
		URL:         "https://example.com/yesterday",
		PublishedAt: yesterday,
	}); err != nil {
		t.Fatalf("saving article: %v", err)
	}

	uc := usecase.NewReadArticles(db, &mockImageStorage{})
	articles, err := uc.Execute(ctx, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article for today, got %d", len(articles))
	}
	if articles[0].ID != "article-today" {
		t.Errorf("expected article ID %q, got %q", "article-today", articles[0].ID)
	}
}
