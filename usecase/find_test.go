package usecase_test

import (
	"context"
	"testing"
	"time"

	"rss-feed/usecase"
)

func TestFind_Execute(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	now := today()

	saveTestArticle(t, ctx, db, now)

	vectorSearch := &mockVectorSearch{ids: []string{"article-1"}}
	uc := usecase.NewFind(vectorSearch, db)

	articles, err := uc.Execute(ctx, "first post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}
	if articles[0].ID != "article-1" {
		t.Errorf("expected article ID %q, got %q", "article-1", articles[0].ID)
	}
}

func TestFind_Execute_SkipsMissingArticles(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	saveTestArticle(t, ctx, db, time.Now())

	vectorSearch := &mockVectorSearch{ids: []string{"article-1", "article-does-not-exist"}}
	uc := usecase.NewFind(vectorSearch, db)

	articles, err := uc.Execute(ctx, "first post")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}
	if articles[0].ID != "article-1" {
		t.Errorf("expected article ID %q, got %q", "article-1", articles[0].ID)
	}
}

func TestFind_Execute_EmptyResult(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)

	vectorSearch := &mockVectorSearch{ids: []string{}}
	uc := usecase.NewFind(vectorSearch, db)

	articles, err := uc.Execute(ctx, "nothing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(articles) != 0 {
		t.Errorf("expected 0 articles, got %d", len(articles))
	}
}
