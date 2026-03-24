package usecase_test

import (
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"testing"
	"time"

	"rss-feed/internal/database"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

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

func createTestFeed(t *testing.T, ctx context.Context, db *database.Gorm) {
	t.Helper()
	if err := db.CreateFeed(ctx, rss.Feed{ID: "feed-abc", Name: "Test Blog", URL: "https://example.com/feed.xml"}); err != nil {
		t.Fatalf("creating feed: %v", err)
	}
}

func saveTestArticle(t *testing.T, ctx context.Context, db *database.Gorm, publishedAt time.Time) {
	t.Helper()
	if err := db.SaveArticle(ctx, "feed-abc", rss.Article{
		ID:          "article-1",
		Title:       "First Post",
		URL:         "https://example.com/post-1",
		PublishedAt: publishedAt,
	}); err != nil {
		t.Fatalf("saving article: %v", err)
	}
}

// — mocks —

type (
	mockPublisher struct{}

	mockSubscriber struct {
		messages []any
	}

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

func (m *mockPublisher) Publish(_ context.Context, _ string, _ event.Message) error { return nil }

func (m *mockSubscriber) Subscribe(_ context.Context, _ string) (<-chan event.Delivery, error) {
	ch := make(chan event.Delivery)
	go func() {
		defer close(ch)
		for _, msg := range m.messages {
			payload, _ := json.Marshal(msg)
			ch <- event.Delivery{
				Message: event.Message{Payload: payload},
				Ack:     func() error { return nil },
				Nack:    func(bool) error { return nil },
			}
		}
	}()
	return ch, nil
}

func (m *mockSubscriber) Close() error { return nil }

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
