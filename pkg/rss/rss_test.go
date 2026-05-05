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

func TestArticle_ToMarkdown_DescriptionOnly(t *testing.T) {
	article := rss.Article{
		Title:       "Hello World",
		Description: "A short summary.",
	}

	got := article.ToMarkdown()
	want := "# Hello World\n\nA short summary."

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestArticle_ToMarkdown_DescriptionAndContent(t *testing.T) {
	article := rss.Article{
		Title:       "Hello World",
		Description: "A short summary.",
		Content:     "Full article content.",
	}

	got := article.ToMarkdown()
	want := "# Hello World\n\nA short summary.\n\n---\n\nFull article content."

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestArticle_ToMarkdown_ConvertsHTMLEntities(t *testing.T) {
	article := rss.Article{
		Title:       "Rust &amp; Go: A Comparison",
		Description: "A comparison article.",
	}

	got := article.ToMarkdown()
	want := "# Rust &amp; Go: A Comparison\n\nA comparison article."

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
