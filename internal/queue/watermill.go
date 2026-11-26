package queue

import (
	"context"
	"fmt"
	"rss-feed/pkg/event"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type (
	// WatermillBroker is an in-process message broker (publisher and subscriber) built on a
	// Watermill GoChannel. Messages live in memory only.
	WatermillBroker struct {
		pubSub *gochannel.GoChannel
	}
)

// NewGoChannel creates a Watermill GoChannel with an output buffer of 64 messages per
// subscriber and persistence enabled, so messages published before a subscription are
// kept for it. Watermill logging is disabled.
func NewGoChannel() *gochannel.GoChannel {
	return gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 64,
			Persistent:          true,
		},
		watermill.NopLogger{},
	)
}

// NewWatermillBroker creates a WatermillBroker on top of pubSub.
func NewWatermillBroker(pubSub *gochannel.GoChannel) *WatermillBroker {
	return &WatermillBroker{
		pubSub: pubSub,
	}
}

// Publish sends the message payload to its topic as a new Watermill message with a fresh
// UUID. Message headers are copied to the Watermill metadata, formatted as strings.
func (b *WatermillBroker) Publish(ctx context.Context, e event.Message) error {
	msg := message.NewMessageWithContext(ctx, watermill.NewUUID(), []byte(e.Payload))
	for k, v := range e.Headers {
		msg.Metadata.Set(k, fmt.Sprintf("%v", v))
	}

	return b.pubSub.Publish(string(e.Topic), msg)
}

// Subscribe subscribes to the topic and returns a channel of deliveries. Ack maps to
// Watermill's Ack. Nack(true) maps to Watermill's Nack, which redelivers the message at once;
// Nack(false) acks the message instead, so it is dropped rather than redelivered (GoChannel
// has no dead-letter queue). The channel is closed when ctx is cancelled or the underlying
// subscription ends.
func (b *WatermillBroker) Subscribe(ctx context.Context, topic event.Topic) (<-chan event.Delivery, error) {
	messages, err := b.pubSub.Subscribe(ctx, string(topic))
	if err != nil {
		return nil, err
	}

	deliveries := make(chan event.Delivery)
	go func() {
		defer close(deliveries)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-messages:
				if !ok {
					return
				}

				headers := make(map[string]interface{}, len(msg.Metadata))
				for k, v := range msg.Metadata {
					headers[k] = v
				}

				delivery := event.Delivery{
					Message: event.Message{
						Topic:   topic,
						Headers: headers,
						Payload: event.Payload(msg.Payload),
					},
					Ack: func() {
						msg.Ack()
					},
					Nack: func(requeue bool) {
						if !requeue {
							msg.Ack()
							return
						}

						msg.Nack()
					},
				}

				select {
				case deliveries <- delivery:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return deliveries, nil
}

// Close closes the underlying GoChannel, ending every subscription.
func (b *WatermillBroker) Close() error {
	return b.pubSub.Close()
}
