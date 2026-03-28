package usecase_test

import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"rss-feed/internal/database"
	"rss-feed/internal/queue"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"
	"rss-feed/usecase"

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

// — test pub/sub helpers —

func newTestPublisher(t *testing.T) usecase.Publisher {
	t.Helper()
	return queue.NewWatermillBroker(queue.NewGoChannel())
}

func newTestBroker(t *testing.T, topic string, messages ...any) usecase.Broker {
	t.Helper()
	broker := queue.NewWatermillBroker(queue.NewGoChannel())
	for _, msg := range messages {
		if err := broker.Publish(context.Background(), topic, event.Message{Payload: event.NewPayload(msg)}); err != nil {
			t.Fatalf("publishing test message: %v", err)
		}
	}
	return broker
}

func awaitCondition(t *testing.T, ctx context.Context, check func() bool) {
	t.Helper()
	for {
		if check() {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("timed out waiting for condition")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// — stubs —

type (
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
