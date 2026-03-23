package usecase_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"
	"rss-feed/usecase"
)

func TestConsumeImage_Execute(t *testing.T) {
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9}) // minimal JPEG bytes
	}))
	defer srv.Close()

	imgEvent := event.Image{
		ImageID:   "img-1",
		ArticleID: "article-1",
		URL:       srv.URL + "/image.jpg",
	}

	uc := usecase.NewConsumeImage(http.DefaultClient, &mockStorage{}, db)
	if err := uc.Execute(ctx, imgEvent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	articles, err := db.GetArticlesFromDate(ctx, now)
	if err != nil {
		t.Fatalf("getting articles: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}
	if len(articles[0].Images) != 1 {
		t.Errorf("expected 1 image on article, got %d", len(articles[0].Images))
	}
}
