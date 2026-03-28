package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"rss-feed/handler"
	"rss-feed/internal/database"
	"rss-feed/pkg/event"
	"rss-feed/presenter"
	"rss-feed/usecase"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// — in-memory broker —

type (
	inMemoryBroker struct {
		queues map[string]chan event.Delivery
	}

	inMemorySubscriber struct {
		broker *inMemoryBroker
		out    chan event.Delivery
	}

	fakeFetcher struct {
		feed event.Feed
	}

	mockFileStorage struct{}

	mockImageStorage struct{}
)

var routingToQueue = map[string]string{
	"feed.article.found":       "rss.feed.article.found",
	"feed.article.image.found": "rss.feed.article.image.found",
	"feed.jobs":                "rss.feed.jobs",
}

func newInMemoryBroker() *inMemoryBroker {
	return &inMemoryBroker{
		queues: map[string]chan event.Delivery{
			"rss.feed.article.found":       make(chan event.Delivery, 10),
			"rss.feed.article.image.found": make(chan event.Delivery, 10),
			"rss.feed.jobs":                make(chan event.Delivery, 10),
		},
	}
}

func (b *inMemoryBroker) Publish(_ context.Context, topic string, msg event.Message) error {
	queue, ok := routingToQueue[topic]
	if !ok {
		return fmt.Errorf("unknown routing key: %s", topic)
	}
	b.queues[queue] <- event.Delivery{
		Message: msg,
		Ack:     func() error { return nil },
		Nack:    func(bool) error { return nil },
	}
	return nil
}

func (b *inMemoryBroker) newSubscriber() *inMemorySubscriber {
	return &inMemorySubscriber{
		broker: b,
		out:    make(chan event.Delivery),
	}
}

func (s *inMemorySubscriber) Subscribe(ctx context.Context, queue string) (<-chan event.Delivery, error) {
	src, ok := s.broker.queues[queue]
	if !ok {
		return nil, fmt.Errorf("unknown queue: %s", queue)
	}
	go func() {
		defer close(s.out)
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-src:
				if !ok {
					return
				}
				select {
				case s.out <- d:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return s.out, nil
}

func (s *inMemorySubscriber) Close() error { return nil }

func (f *fakeFetcher) Fetch(_ context.Context, _ string) (event.Feed, error) {
	return f.feed, nil
}

func (m *mockFileStorage) Create(_ context.Context, _ string, _ io.Reader) error { return nil }

func (m *mockImageStorage) Presign(_ context.Context, path string, _ time.Duration) (string, error) {
	return "https://cdn.example.com/" + path, nil
}

// — test —

func TestEndToEnd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	gormDB, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "e2e.sqlite")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	db := database.NewGorm(gormDB)

	imageSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9})
	}))
	t.Cleanup(imageSrv.Close)

	broker := newInMemoryBroker()

	fetcher := &fakeFetcher{
		feed: event.Feed{
			FeedID: "feed-e2e",
			Name:   "E2E Blog",
			URL:    "https://example.com/feed.xml",
			Articles: []event.Article{
				{
					ArticleID:   "article-e2e",
					Title:       "E2E Article",
					URL:         "https://example.com/article-1",
					PublishedAt: time.Now().UTC(),
					Image: event.Image{
						ImageID:   "image-e2e",
						ArticleID: "article-e2e",
						URL:       imageSrv.URL + "/image.jpg",
					},
				},
			},
		},
	}

	saveFeedUC := usecase.NewSaveFeed(fetcher, db, broker)
	consumeFeedUC := usecase.NewConsumeFeed(db, broker.newSubscriber(), broker)
	consumeImageUC := usecase.NewConsumeImage(http.DefaultClient, broker.newSubscriber(), &mockFileStorage{}, db)
	readArticlesUC := usecase.NewReadArticles(db, &mockImageStorage{})

	go func() {
		_ = consumeFeedUC.Execute(ctx)
	}()
	go func() {
		_ = consumeImageUC.Execute(ctx)
	}()

	r := chi.NewRouter()
	r.Post("/feeds", handler.AddFeed(saveFeedUC))
	r.Get("/articles/today", handler.ListToday(readArticlesUC))
	apiSrv := httptest.NewServer(r)
	t.Cleanup(apiSrv.Close)

	resp, err := http.Post(apiSrv.URL+"/feeds", "application/json", strings.NewReader(`{"url":"https://example.com/feed.xml"}`))
	if err != nil {
		t.Fatalf("POST /feeds: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 from POST /feeds, got %d", resp.StatusCode)
	}

	var result presenter.ArticleListResponse
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err = http.Get(apiSrv.URL + "/articles/today")
		if err != nil {
			t.Fatalf("GET /articles/today: %v", err)
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("decoding response: %v", err)
		}
		resp.Body.Close()
		if len(result.Articles) > 0 && len(result.Articles[0].Images) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if len(result.Articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(result.Articles))
	}
	article := result.Articles[0]
	if article.Title != "E2E Article" {
		t.Errorf("title: want %q, got %q", "E2E Article", article.Title)
	}
	if article.URL != "https://example.com/article-1" {
		t.Errorf("URL: want %q, got %q", "https://example.com/article-1", article.URL)
	}
	if article.Feed.ID != "feed-e2e" {
		t.Errorf("feed.id: want %q, got %q", "feed-e2e", article.Feed.ID)
	}
	if article.Feed.Name != "E2E Blog" {
		t.Errorf("feed.name: want %q, got %q", "E2E Blog", article.Feed.Name)
	}
	if len(article.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(article.Images))
	}
	if article.Images[0].Type != "original" {
		t.Errorf("image.type: want %q, got %q", "original", article.Images[0].Type)
	}
	wantImageURL := "https://cdn.example.com/images/original/article-e2e.jpg"
	if article.Images[0].URL != wantImageURL {
		t.Errorf("image.url: want %q, got %q", wantImageURL, article.Images[0].URL)
	}
}
