package usecase_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"rss-feed/pkg/event"
	"rss-feed/usecase"
)

func TestConsumeFeed_Execute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := today()
	feedEvent := event.Feed{
		FeedID: "feed-abc",
		Name:   "Test Blog",
		URL:    "https://example.com/feed.xml",
		Articles: []event.Article{
			{
				ArticleID:   "article-1",
				Title:       "First Post",
				URL:         "https://example.com/post-1",
				PublishedAt: now,
			},
			{
				ArticleID:   "article-2",
				Title:       "Second Post",
				URL:         "https://example.com/post-2",
				PublishedAt: now,
				Image:       event.Image{ImageID: "img-1", ArticleID: "article-2", URL: "https://example.com/img.jpg"},
			},
		},
	}

	broker := newTestBroker(t, event.TopicFeedArticleFound, feedEvent)

	imageDeliveries, err := broker.Subscribe(ctx, event.TopicFeedArticleImageFound)
	if err != nil {
		t.Fatalf("subscribing to image topic: %v", err)
	}

	uc := usecase.NewHandleArticles(broker)
	go func() { _ = uc.Execute(ctx) }()

	select {
	case d := <-imageDeliveries:
		d.Ack()
		var got event.Image
		if err := json.Unmarshal(d.Payload, &got); err != nil {
			t.Fatalf("unmarshaling image event: %v", err)
		}
		if got.ImageID != "img-1" {
			t.Errorf("expected ImageID %q, got %q", "img-1", got.ImageID)
		}
		if got.ArticleID != "article-2" {
			t.Errorf("expected ArticleID %q, got %q", "article-2", got.ArticleID)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for image event")
	}
}
