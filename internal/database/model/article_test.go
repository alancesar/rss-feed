package model_test

import (
	"testing"
	"time"

	"rss-feed/internal/database/model"
	"rss-feed/pkg/rss"
)

func TestNewFeedFromDomain(t *testing.T) {
	publishedAt := time.Date(2025, 10, 27, 14, 0, 0, 0, time.UTC)
	feed := rss.Feed{
		ID:   "feed-abc",
		Name: "Test Blog",
		URL:  "https://example.com/feed.xml",
		Articles: []rss.Article{
			{ID: "article-1", Title: "First Post", URL: "https://example.com/post-1", PublishedAt: publishedAt},
			{ID: "article-2", Title: "Second Post", URL: "https://example.com/post-2", PublishedAt: publishedAt},
		},
	}

	got := model.NewFeedFromDomain(feed)

	if got.ID != feed.ID {
		t.Errorf("expected ID %q, got %q", feed.ID, got.ID)
	}
	if got.Name != feed.Name {
		t.Errorf("expected Name %q, got %q", feed.Name, got.Name)
	}
	if got.URL != feed.URL {
		t.Errorf("expected URL %q, got %q", feed.URL, got.URL)
	}
	if len(got.Articles) != len(feed.Articles) {
		t.Fatalf("expected %d articles, got %d", len(feed.Articles), len(got.Articles))
	}
	for i, a := range feed.Articles {
		if got.Articles[i].ID != a.ID {
			t.Errorf("article[%d]: expected ID %q, got %q", i, a.ID, got.Articles[i].ID)
		}
		if got.Articles[i].FeedID != feed.ID {
			t.Errorf("article[%d]: expected FeedID %q, got %q", i, feed.ID, got.Articles[i].FeedID)
		}
	}
}

func TestNewArticleFromDomain(t *testing.T) {
	publishedAt := time.Date(2025, 10, 27, 14, 0, 0, 0, time.UTC)
	article := rss.Article{
		ID:          "article-1",
		Title:       "First Post",
		Description: "A short description",
		Content:     "Full content here",
		Categories:  []string{"tech", "go"},
		URL:         "https://example.com/post-1",
		PublishedAt: publishedAt,
	}

	got := model.NewArticleFromDomain(article, "feed-abc")

	if got.ID != article.ID {
		t.Errorf("expected ID %q, got %q", article.ID, got.ID)
	}
	if got.FeedID != "feed-abc" {
		t.Errorf("expected FeedID %q, got %q", "feed-abc", got.FeedID)
	}
	if got.Title != article.Title {
		t.Errorf("expected Title %q, got %q", article.Title, got.Title)
	}
	if got.Description != article.Description {
		t.Errorf("expected Description %q, got %q", article.Description, got.Description)
	}
	if got.Content != article.Content {
		t.Errorf("expected Content %q, got %q", article.Content, got.Content)
	}
	if got.URL != article.URL {
		t.Errorf("expected URL %q, got %q", article.URL, got.URL)
	}
	if !got.PublishedAt.Equal(publishedAt) {
		t.Errorf("expected PublishedAt %v, got %v", publishedAt, got.PublishedAt)
	}
}

func TestNewImageFromDomain(t *testing.T) {
	img := rss.NewImage("images/original/article-1.jpg", rss.OriginalImageType)

	got := model.NewImageFromDomain(img, "article-1")

	if got.ID != img.ID {
		t.Errorf("expected ID %q, got %q", img.ID, got.ID)
	}
	if got.ArticleID != "article-1" {
		t.Errorf("expected ArticleID %q, got %q", "article-1", got.ArticleID)
	}
	if got.Path != img.Path {
		t.Errorf("expected Path %q, got %q", img.Path, got.Path)
	}
	if got.Type != string(img.Type) {
		t.Errorf("expected Type %q, got %q", string(img.Type), got.Type)
	}
}

func TestFeed_ToDomain(t *testing.T) {
	m := model.Feed{
		ID:   "feed-abc",
		Name: "Test Blog",
		URL:  "https://example.com/feed.xml",
	}

	got := m.ToDomain()

	if got.ID != m.ID {
		t.Errorf("expected ID %q, got %q", m.ID, got.ID)
	}
	if got.Name != m.Name {
		t.Errorf("expected Name %q, got %q", m.Name, got.Name)
	}
	if got.URL != m.URL {
		t.Errorf("expected URL %q, got %q", m.URL, got.URL)
	}
}

func TestArticle_ToDomain(t *testing.T) {
	publishedAt := time.Date(2025, 10, 27, 14, 0, 0, 0, time.UTC)
	m := model.Article{
		ID:          "article-1",
		Title:       "First Post",
		Description: "A short description",
		Content:     "Full content here",
		Categories:  []byte(`["tech","go"]`),
		URL:         "https://example.com/post-1",
		PublishedAt: publishedAt,
		Feed:        model.Feed{ID: "feed-abc", Name: "Test Blog", URL: "https://example.com/feed.xml"},
		Images: []model.Image{
			{ID: "img-1", Path: "images/original/article-1.jpg", Type: string(rss.OriginalImageType)},
		},
	}

	got := m.ToDomain()

	if got.ID != m.ID {
		t.Errorf("expected ID %q, got %q", m.ID, got.ID)
	}
	if got.Title != m.Title {
		t.Errorf("expected Title %q, got %q", m.Title, got.Title)
	}
	if got.Description != m.Description {
		t.Errorf("expected Description %q, got %q", m.Description, got.Description)
	}
	if got.Content != m.Content {
		t.Errorf("expected Content %q, got %q", m.Content, got.Content)
	}
	if len(got.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(got.Categories))
	}
	if got.Categories[0] != "tech" || got.Categories[1] != "go" {
		t.Errorf("expected categories [tech go], got %v", got.Categories)
	}
	if got.URL != m.URL {
		t.Errorf("expected URL %q, got %q", m.URL, got.URL)
	}
	if !got.PublishedAt.Equal(publishedAt) {
		t.Errorf("expected PublishedAt %v, got %v", publishedAt, got.PublishedAt)
	}
	if got.Feed.ID != m.Feed.ID {
		t.Errorf("expected Feed.ID %q, got %q", m.Feed.ID, got.Feed.ID)
	}
	if len(got.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(got.Images))
	}
	if got.Images[0].Path != m.Images[0].Path {
		t.Errorf("expected image Path %q, got %q", m.Images[0].Path, got.Images[0].Path)
	}
}

func TestImage_ToDomain(t *testing.T) {
	m := model.Image{
		ID:   "img-1",
		Path: "images/original/article-1.jpg",
		Type: string(rss.OriginalImageType),
	}

	got := m.ToDomain()

	if got.ID != m.ID {
		t.Errorf("expected ID %q, got %q", m.ID, got.ID)
	}
	if got.Path != m.Path {
		t.Errorf("expected Path %q, got %q", m.Path, got.Path)
	}
	if got.Type != rss.OriginalImageType {
		t.Errorf("expected Type %q, got %q", rss.OriginalImageType, got.Type)
	}
}
