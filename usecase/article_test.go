package usecase_test

import (
	"context"
	"testing"
	"time"

	"rss-feed/pkg/event"
	"rss-feed/usecase"
)

func TestArticle_Execute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := today()
	articleEvent := event.Article{
		ArticleID:   "article-1",
		Title:       "First Post",
		URL:         "https://example.com/post-1",
		PublishedAt: now,
	}

	broker := newTestBroker(t, event.TopicFeedArticleFound, articleEvent)
	store := &mockArticleStore{}
	uc := usecase.NewArticle(broker, store)
	go func() { _ = uc.Execute(ctx) }()

	awaitCondition(t, ctx, func() bool {
		return len(store.stored()) == 1
	})

	articles := store.stored()
	if articles[0].ID != "article-1" {
		t.Errorf("expected article ID %q, got %q", "article-1", articles[0].ID)
	}
	if articles[0].Title != "First Post" {
		t.Errorf("expected title %q, got %q", "First Post", articles[0].Title)
	}
}

func TestArticle_Execute_SkipsInvalidPayload(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := today()
	validEvent := event.Article{
		ArticleID:   "article-2",
		Title:       "Second Post",
		URL:         "https://example.com/post-2",
		PublishedAt: now,
	}

	broker := newTestBroker(t, event.TopicFeedArticleFound, "not valid json", validEvent)
	store := &mockArticleStore{}
	uc := usecase.NewArticle(broker, store)
	go func() { _ = uc.Execute(ctx) }()

	awaitCondition(t, ctx, func() bool {
		return len(store.stored()) == 1
	})

	articles := store.stored()
	if len(articles) != 1 {
		t.Fatalf("expected 1 stored article, got %d", len(articles))
	}
	if articles[0].ID != "article-2" {
		t.Errorf("expected article ID %q, got %q", "article-2", articles[0].ID)
	}
}
