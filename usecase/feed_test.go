package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"
	"rss-feed/usecase"
)

type (
	failingPublisherBroker struct {
		deliveries chan event.Delivery
	}

	noopArticleStore struct{}
)

func (s *noopArticleStore) SaveArticle(_ context.Context, _ rss.Article) error {
	return nil
}

func (b *failingPublisherBroker) Publish(_ context.Context, _ event.Message) error {
	return errors.New("publish failed")
}

func (b *failingPublisherBroker) Subscribe(_ context.Context, _ event.Topic) (<-chan event.Delivery, error) {
	return b.deliveries, nil
}

func (b *failingPublisherBroker) Close() error {
	return nil
}

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

	broker := newTestBroker(t, event.TopicFeedFound, feedEvent)

	articleDeliveries, err := broker.Subscribe(ctx, event.TopicFeedArticleFound)
	if err != nil {
		t.Fatalf("subscribing to article topic: %v", err)
	}

	uc := usecase.NewHandleFeed(broker, &noopArticleStore{})
	go func() { _ = uc.Execute(ctx) }()

	for {
		select {
		case d := <-articleDeliveries:
			d.Ack()
			var got event.Article
			if err := json.Unmarshal(d.Payload, &got); err != nil {
				t.Fatalf("unmarshaling article event: %v", err)
			}
			if got.Image.ImageID == "" {
				continue
			}
			if got.Image.ImageID != "img-1" {
				t.Errorf("expected ImageID %q, got %q", "img-1", got.Image.ImageID)
			}
			if got.ArticleID != "article-2" {
				t.Errorf("expected ArticleID %q, got %q", "article-2", got.ArticleID)
			}
			return
		case <-ctx.Done():
			t.Fatal("timed out waiting for image event")
		}
	}
}

func TestConsumeFeed_Execute_PublishFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	feedEvent := event.Feed{
		FeedID: "feed-abc",
		Articles: []event.Article{
			{
				ArticleID:   "article-1",
				Title:       "First Post",
				URL:         "https://example.com/post-1",
				PublishedAt: today(),
			},
		},
	}

	nacked := make(chan bool, 1)
	acked := make(chan struct{}, 1)

	deliveryCh := make(chan event.Delivery, 1)
	deliveryCh <- event.Delivery{
		Message: event.Message{
			Topic:   event.TopicFeedFound,
			Payload: event.NewPayload(feedEvent),
		},
		Ack:  func() { acked <- struct{}{} },
		Nack: func(requeue bool) { nacked <- requeue },
	}

	uc := usecase.NewHandleFeed(&failingPublisherBroker{deliveries: deliveryCh}, &noopArticleStore{})
	go func() { _ = uc.Execute(ctx) }()

	select {
	case requeue := <-nacked:
		if !requeue {
			t.Errorf("expected Nack(true), got Nack(false)")
		}
	case <-acked:
		t.Error("expected Nack but got Ack")
	case <-ctx.Done():
		t.Fatal("timed out waiting for Ack/Nack")
	}
}
