package usecase_test

import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"rss-feed/internal/database"
	"rss-feed/pkg/event"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// — helpers —

func newTestDB(t *testing.T) *database.Gorm {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.sqlite")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	return database.NewGorm(db)
}

func today() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

// — mocks —

type mockPublisher struct{}

func (m *mockPublisher) Publish(_ context.Context, _ string, _ event.Event) error { return nil }

type mockFetcher struct {
	feed event.Feed
}

func (m *mockFetcher) Fetch(_ context.Context, _ string) (event.Feed, error) {
	return m.feed, nil
}

type recordingFetcher struct {
	calls []string
	feed  event.Feed
}

func (f *recordingFetcher) Fetch(_ context.Context, url string) (event.Feed, error) {
	f.calls = append(f.calls, url)
	return f.feed, nil
}

type funcFetcher struct {
	fn func(url string) (event.Feed, error)
}

func (f *funcFetcher) Fetch(_ context.Context, url string) (event.Feed, error) {
	return f.fn(url)
}

type mockStorage struct{}

func (m *mockStorage) Create(_ context.Context, _ string, _ io.Reader) error { return nil }

type mockImageStorage struct{}

func (m *mockImageStorage) Presign(_ context.Context, path string, _ time.Duration) (string, error) {
	return "https://cdn.example.com/" + path, nil
}
