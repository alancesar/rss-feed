package queue_test

import (
	"context"
	"testing"
	"time"

	"rss-feed/internal/queue"
	"rss-feed/pkg/event"
)

func TestWatermillBroker_Nack_NoRequeue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	broker, deliveries := subscribeTestBroker(t, ctx)
	publishTestMessage(t, ctx, broker, "invalid")

	receiveTestDelivery(t, ctx, deliveries).Nack(false)

	select {
	case d := <-deliveries:
		t.Errorf("expected the message to be dropped, got redelivery of %q", d.Payload)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestWatermillBroker_Nack_Requeue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	broker, deliveries := subscribeTestBroker(t, ctx)
	publishTestMessage(t, ctx, broker, "retry")

	receiveTestDelivery(t, ctx, deliveries).Nack(true)

	d := receiveTestDelivery(t, ctx, deliveries)
	if string(d.Payload) != "retry" {
		t.Errorf("expected %q to be redelivered, got %q", "retry", d.Payload)
	}
	d.Ack()
}

// subscribeTestBroker creates a WatermillBroker subscribed to TopicFeedFound and closes it
// when the test ends.
func subscribeTestBroker(t *testing.T, ctx context.Context) (*queue.WatermillBroker, <-chan event.Delivery) {
	t.Helper()

	broker := queue.NewWatermillBroker(queue.NewGoChannel())
	t.Cleanup(func() {
		_ = broker.Close()
	})

	deliveries, err := broker.Subscribe(ctx, event.TopicFeedFound)
	if err != nil {
		t.Fatalf("subscribing: %v", err)
	}

	return broker, deliveries
}

// publishTestMessage publishes payload to TopicFeedFound.
func publishTestMessage(t *testing.T, ctx context.Context, broker *queue.WatermillBroker, payload string) {
	t.Helper()

	if err := broker.Publish(ctx, event.Message{Topic: event.TopicFeedFound, Payload: event.Payload(payload)}); err != nil {
		t.Fatalf("publishing %q: %v", payload, err)
	}
}

// receiveTestDelivery waits for the next delivery, failing the test on timeout.
func receiveTestDelivery(t *testing.T, ctx context.Context, deliveries <-chan event.Delivery) event.Delivery {
	t.Helper()

	select {
	case d, ok := <-deliveries:
		if !ok {
			t.Fatal("deliveries channel closed")
		}
		return d
	case <-ctx.Done():
		t.Fatal("timed out waiting for delivery")
	}

	return event.Delivery{}
}
