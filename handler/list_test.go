package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rss-feed/handler"
	"rss-feed/pkg/rss"
	"rss-feed/presenter"
	"rss-feed/usecase"
)

func TestGetFromDate_InvalidDate(t *testing.T) {
	uc := usecase.NewReadArticles(&mockArticlesStore{}, &mockImageStorage{})

	r := httptest.NewRequest(http.MethodGet, "/articles?date=not-a-date", nil)
	w := httptest.NewRecorder()

	handler.GetFromDate(uc)(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetFromDate_ValidDate(t *testing.T) {
	articles := []rss.Article{
		{ID: "article-1", Title: "First Post", URL: "https://example.com/post-1", PublishedAt: time.Now()},
	}
	uc := usecase.NewReadArticles(&mockArticlesStore{articles: articles}, &mockImageStorage{})

	r := httptest.NewRequest(http.MethodGet, "/articles?date=2025-10-27", nil)
	w := httptest.NewRecorder()

	handler.GetFromDate(uc)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp presenter.ArticleListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(resp.Articles) != 1 {
		t.Errorf("expected 1 article, got %d", len(resp.Articles))
	}
	if resp.Articles[0].Title != "First Post" {
		t.Errorf("expected title %q, got %q", "First Post", resp.Articles[0].Title)
	}
}

func TestListToday(t *testing.T) {
	articles := []rss.Article{
		{ID: "article-1", Title: "Today Post", URL: "https://example.com/today", PublishedAt: time.Now()},
	}
	uc := usecase.NewReadArticles(&mockArticlesStore{articles: articles}, &mockImageStorage{})

	r := httptest.NewRequest(http.MethodGet, "/articles/today", nil)
	w := httptest.NewRecorder()

	handler.ListToday(uc)(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp presenter.ArticleListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(resp.Articles) != 1 {
		t.Errorf("expected 1 article, got %d", len(resp.Articles))
	}
}
