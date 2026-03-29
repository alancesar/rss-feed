package event_test

import (
	"testing"
	"time"

	"rss-feed/pkg/event"
)

func TestArticle_ToDomain(t *testing.T) {
	publishedAt := time.Date(2025, 10, 27, 14, 0, 0, 0, time.UTC)
	a := event.Article{
		ArticleID:   "article-1",
		FeedID:      "feed-abc",
		Title:       "First Post",
		Description: "A short description",
		Content:     "Full content here",
		Categories:  []string{"tech", "go"},
		URL:         "https://example.com/post-1",
		PublishedAt: publishedAt,
	}

	got := a.ToDomain()

	if got.ID != a.ArticleID {
		t.Errorf("expected ID %q, got %q", a.ArticleID, got.ID)
	}
	if got.Title != a.Title {
		t.Errorf("expected Title %q, got %q", a.Title, got.Title)
	}
	if got.Description != a.Description {
		t.Errorf("expected Description %q, got %q", a.Description, got.Description)
	}
	if got.Content != a.Content {
		t.Errorf("expected Content %q, got %q", a.Content, got.Content)
	}
	if len(got.Categories) != len(a.Categories) {
		t.Fatalf("expected %d categories, got %d", len(a.Categories), len(got.Categories))
	}
	for i, c := range a.Categories {
		if got.Categories[i] != c {
			t.Errorf("categories[%d]: expected %q, got %q", i, c, got.Categories[i])
		}
	}
	if got.URL != a.URL {
		t.Errorf("expected URL %q, got %q", a.URL, got.URL)
	}
	if !got.PublishedAt.Equal(publishedAt) {
		t.Errorf("expected PublishedAt %v, got %v", publishedAt, got.PublishedAt)
	}
}

func TestFeed_ToDomain(t *testing.T) {
	publishedAt := time.Date(2025, 10, 27, 14, 0, 0, 0, time.UTC)
	f := event.Feed{
		FeedID: "feed-abc",
		Name:   "Test Blog",
		URL:    "https://example.com/feed.xml",
		Articles: []event.Article{
			{ArticleID: "article-1", Title: "First Post", URL: "https://example.com/post-1", PublishedAt: publishedAt},
			{ArticleID: "article-2", Title: "Second Post", URL: "https://example.com/post-2", PublishedAt: publishedAt},
		},
	}

	got := f.ToDomain()

	if got.ID != f.FeedID {
		t.Errorf("expected ID %q, got %q", f.FeedID, got.ID)
	}
	if got.Name != f.Name {
		t.Errorf("expected Name %q, got %q", f.Name, got.Name)
	}
	if got.URL != f.URL {
		t.Errorf("expected URL %q, got %q", f.URL, got.URL)
	}
	if len(got.Articles) != len(f.Articles) {
		t.Fatalf("expected %d articles, got %d", len(f.Articles), len(got.Articles))
	}
	for i, a := range f.Articles {
		if got.Articles[i].ID != a.ArticleID {
			t.Errorf("article[%d]: expected ID %q, got %q", i, a.ArticleID, got.Articles[i].ID)
		}
	}
}

func TestFeed_ToDomain_Empty(t *testing.T) {
	f := event.Feed{
		FeedID: "feed-abc",
		Name:   "Test Blog",
		URL:    "https://example.com/feed.xml",
	}

	got := f.ToDomain()

	if got.Articles == nil {
		t.Error("expected non-nil Articles slice")
	}
	if len(got.Articles) != 0 {
		t.Errorf("expected 0 articles, got %d", len(got.Articles))
	}
}
