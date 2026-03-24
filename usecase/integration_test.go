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

type (
	mockPublisher struct{}

	mockFetcher struct {
		feed event.Feed
	}

	recordingFetcher struct {
		calls []string
		feed  event.Feed
	}

	funcFetcher struct {
		fn func(url string) (event.Feed, error)
	}

	mockStorage struct{}

	mockImageStorage struct{}
)

func (m *mockPublisher) Publish(_ context.Context, _ string, _ event.Event) error { return nil }

func (m *mockFetcher) Fetch(_ context.Context, _ string) (event.Feed, error) {
	return m.feed, nil
}

func (f *recordingFetcher) Fetch(_ context.Context, url string) (event.Feed, error) {
	f.calls = append(f.calls, url)
	return f.feed, nil
}

func (f *funcFetcher) Fetch(_ context.Context, url string) (event.Feed, error) {
	return f.fn(url)
}

func (m *mockStorage) Create(_ context.Context, _ string, _ io.Reader) error { return nil }

func (m *mockImageStorage) Presign(_ context.Context, path string, _ time.Duration) (string, error) {
	return "https://cdn.example.com/" + path, nil
}
