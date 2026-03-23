package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"rss-feed/handler"
	"rss-feed/pkg/event"
	"rss-feed/presenter"
	"rss-feed/usecase"
)

func TestAddFeed_InvalidBody(t *testing.T) {
	uc := usecase.NewSaveFeed(&mockFeedFetcher{}, &mockFeedStore{}, &mockPublisher{})

	r := httptest.NewRequest(http.MethodPost, "/feeds", strings.NewReader("not-json"))
	w := httptest.NewRecorder()

	handler.AddFeed(uc)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAddFeed_DuplicateFeed(t *testing.T) {
	store := &mockFeedStore{existingURL: "https://example.com/feed.xml"}
	uc := usecase.NewSaveFeed(&mockFeedFetcher{}, store, &mockPublisher{})

	body := `{"url":"https://example.com/feed.xml"}`
	r := httptest.NewRequest(http.MethodPost, "/feeds", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.AddFeed(uc)(w, r)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestAddFeed_Success(t *testing.T) {
	fetcher := &mockFeedFetcher{
		feed: event.Feed{
			FeedID: "feed-abc",
			Name:   "Test Blog",
			URL:    "https://example.com/feed.xml",
		},
	}
	uc := usecase.NewSaveFeed(fetcher, &mockFeedStore{}, &mockPublisher{})

	body := `{"url":"https://example.com/feed.xml"}`
	r := httptest.NewRequest(http.MethodPost, "/feeds", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.AddFeed(uc)(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp presenter.FeedResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.ID != "feed-abc" {
		t.Errorf("expected ID %q, got %q", "feed-abc", resp.ID)
	}
	if resp.Name != "Test Blog" {
		t.Errorf("expected Name %q, got %q", "Test Blog", resp.Name)
	}
	if resp.URL != "https://example.com/feed.xml" {
		t.Errorf("expected URL %q, got %q", "https://example.com/feed.xml", resp.URL)
	}
}
