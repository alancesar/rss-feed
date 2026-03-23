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
