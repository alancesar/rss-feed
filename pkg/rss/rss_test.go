package rss_test

import (
	"testing"

	"rss-feed/pkg/rss"
)

func TestNewImage(t *testing.T) {
	img := rss.NewImage("images/original/article-1.jpg", rss.OriginalImageType)

	if img.Path != "images/original/article-1.jpg" {
		t.Errorf("expected Path %q, got %q", "images/original/article-1.jpg", img.Path)
	}
	if img.Type != rss.OriginalImageType {
		t.Errorf("expected Type %q, got %q", rss.OriginalImageType, img.Type)
	}
	if img.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestNewImage_IDDerivedFromPath(t *testing.T) {
	a := rss.NewImage("images/original/article-1.jpg", rss.OriginalImageType)
	b := rss.NewImage("images/original/article-1.jpg", rss.OriginalImageType)
	c := rss.NewImage("images/original/article-2.jpg", rss.OriginalImageType)

	if a.ID != b.ID {
		t.Error("expected same path to produce same ID")
	}
	if a.ID == c.ID {
		t.Error("expected different paths to produce different IDs")
	}
}

func TestArticle_ToMarkdown(t *testing.T) {
	article := rss.Article{
		Title: "Hello World",
		URL:   "https://example.com/post-1",
	}

	got := article.ToMarkdown()
	want := "[Hello World](https://example.com/post-1)"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestArticle_ToMarkdown_UnescapesHTMLEntities(t *testing.T) {
	article := rss.Article{
		Title: "Rust &amp; Go: A Comparison",
		URL:   "https://example.com/post-1",
	}

	got := article.ToMarkdown()
	want := "[Rust & Go: A Comparison](https://example.com/post-1)"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
