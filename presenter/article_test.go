package presenter_test

import (
	"testing"
	"time"

	"rss-feed/pkg/rss"
	"rss-feed/presenter"
)

func TestNewImageResponseFromDomain(t *testing.T) {
	image := rss.Image{
		Path: "images/original/article-1.jpg",
		Type: rss.OriginalImageType,
	}

	got := presenter.NewImageResponseFromDomain(image)

	if got.URL != image.Path {
		t.Errorf("expected URL %q, got %q", image.Path, got.URL)
	}
	if got.Type != string(image.Type) {
		t.Errorf("expected Type %q, got %q", string(image.Type), got.Type)
	}
}

func TestNewArticleResponseFromDomain(t *testing.T) {
	publishedAt := time.Date(2025, 10, 27, 14, 0, 0, 0, time.UTC)
	article := rss.Article{
		ID:    "article-1",
		Title: "First Post",
		URL:   "https://example.com/post-1",
		Feed: rss.Feed{
			ID:   "feed-abc",
			Name: "Test Blog",
			URL:  "https://example.com/feed.xml",
		},
		Images: []rss.Image{
			{Path: "images/original/article-1.jpg", Type: rss.OriginalImageType},
			{Path: "images/thumbnail/article-1.jpg", Type: rss.ThumbnailImageType},
		},
		PublishedAt: publishedAt,
	}

	got := presenter.NewArticleResponseFromDomain(article)

	if got.Title != article.Title {
		t.Errorf("expected Title %q, got %q", article.Title, got.Title)
	}
	if got.URL != article.URL {
		t.Errorf("expected URL %q, got %q", article.URL, got.URL)
	}
	if got.PublishedAt != publishedAt {
		t.Errorf("expected PublishedAt %v, got %v", publishedAt, got.PublishedAt)
	}
	if got.Feed.ID != article.Feed.ID {
		t.Errorf("expected Feed.ID %q, got %q", article.Feed.ID, got.Feed.ID)
	}
	if got.Feed.Name != article.Feed.Name {
		t.Errorf("expected Feed.Name %q, got %q", article.Feed.Name, got.Feed.Name)
	}
	if got.Feed.URL != article.Feed.URL {
		t.Errorf("expected Feed.URL %q, got %q", article.Feed.URL, got.Feed.URL)
	}
	if len(got.Images) != len(article.Images) {
		t.Fatalf("expected %d images, got %d", len(article.Images), len(got.Images))
	}
	for i, img := range article.Images {
		if got.Images[i].URL != img.Path {
			t.Errorf("image[%d]: expected URL %q, got %q", i, img.Path, got.Images[i].URL)
		}
		if got.Images[i].Type != string(img.Type) {
			t.Errorf("image[%d]: expected Type %q, got %q", i, string(img.Type), got.Images[i].Type)
		}
	}
}

func TestNewArticleListResponse(t *testing.T) {
	articles := []rss.Article{
		{
			ID:          "article-1",
			Title:       "First Post",
			URL:         "https://example.com/post-1",
			PublishedAt: time.Date(2025, 10, 27, 14, 0, 0, 0, time.UTC),
		},
		{
			ID:          "article-2",
			Title:       "Second Post",
			URL:         "https://example.com/post-2",
			PublishedAt: time.Date(2025, 10, 27, 15, 0, 0, 0, time.UTC),
		},
	}

	got := presenter.NewArticleListResponse(articles)

	if len(got.Articles) != len(articles) {
		t.Fatalf("expected %d articles, got %d", len(articles), len(got.Articles))
	}
	for i, a := range articles {
		if got.Articles[i].Title != a.Title {
			t.Errorf("article[%d]: expected Title %q, got %q", i, a.Title, got.Articles[i].Title)
		}
		if got.Articles[i].URL != a.URL {
			t.Errorf("article[%d]: expected URL %q, got %q", i, a.URL, got.Articles[i].URL)
		}
	}
}

func TestNewArticleListResponse_Empty(t *testing.T) {
	got := presenter.NewArticleListResponse([]rss.Article{})

	if got.Articles == nil {
		t.Error("expected non-nil Articles slice")
	}
	if len(got.Articles) != 0 {
		t.Errorf("expected 0 articles, got %d", len(got.Articles))
	}
}
