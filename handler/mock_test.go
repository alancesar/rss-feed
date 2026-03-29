package handler_test

import (
	"context"
	"time"

	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"
)

type (
	mockArticlesStore struct {
		articles []rss.Article
	}

	mockImageStorage struct{}

	mockFeedFetcher struct {
		feed event.Feed
	}

	mockFeedStore struct {
		existingURL string
	}
)

func (m *mockArticlesStore) GetArticlesFromDate(_ context.Context, _ time.Time) ([]rss.Article, error) {
	return m.articles, nil
}

func (m *mockImageStorage) Presign(_ context.Context, path string, _ time.Duration) (string, error) {
	return "https://cdn.example.com/" + path, nil
}

func (m *mockFeedFetcher) Fetch(_ context.Context, _ string) (event.Feed, error) {
	return m.feed, nil
}

func (m *mockFeedStore) GetFeedByURL(_ context.Context, url string) (rss.Feed, bool, error) {
	if url == m.existingURL {
		return rss.Feed{}, true, nil
	}
	return rss.Feed{}, false, nil
}

func (m *mockFeedStore) CreateFeed(_ context.Context, _ rss.Feed) error {
	return nil
}
