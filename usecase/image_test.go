package usecase_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"rss-feed/pkg/event"
	"rss-feed/usecase"
)

func TestConsumeImage_Execute(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	now := today()

	createTestFeed(t, ctx, db)
	saveTestArticle(t, ctx, db, now)

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

	subscriber := &mockSubscriber{messages: []any{imgEvent}}
	uc := usecase.NewConsumeImage(http.DefaultClient, subscriber, &mockStorage{}, db)
	if err := uc.Execute(ctx); err != nil {
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
