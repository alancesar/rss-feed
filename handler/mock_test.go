package handler_test

import (
	"context"
	"time"

	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"
)

type mockArticlesStore struct {
	articles []rss.Article
}

func (m *mockArticlesStore) GetArticlesFromDate(_ context.Context, _ time.Time) ([]rss.Article, error) {
	return m.articles, nil
}

type mockImageStorage struct{}

func (m *mockImageStorage) Presign(_ context.Context, path string, _ time.Duration) (string, error) {
	return "https://cdn.example.com/" + path, nil
}

type mockFeedFetcher struct {
	feed event.Feed
}

func (m *mockFeedFetcher) Fetch(_ context.Context, _ string) (event.Feed, error) {
	return m.feed, nil
}

type mockFeedStore struct {
	existingURL string
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

type mockPublisher struct{}

func (m *mockPublisher) Publish(_ context.Context, _ string, _ event.Event) error { return nil }
