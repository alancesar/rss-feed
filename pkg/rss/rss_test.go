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
