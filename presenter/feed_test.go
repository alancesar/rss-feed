package presenter_test

import (
	"testing"

	"rss-feed/pkg/rss"
	"rss-feed/presenter"
)

func TestNewFeedResponse(t *testing.T) {
	feed := rss.Feed{
		ID:   "feed-abc",
		Name: "Test Blog",
		URL:  "https://example.com/feed.xml",
	}

	got := presenter.NewFeedResponse(feed)

	if got.ID != feed.ID {
		t.Errorf("expected ID %q, got %q", feed.ID, got.ID)
	}
	if got.Name != feed.Name {
		t.Errorf("expected Name %q, got %q", feed.Name, got.Name)
	}
	if got.URL != feed.URL {
		t.Errorf("expected URL %q, got %q", feed.URL, got.URL)
	}
}
