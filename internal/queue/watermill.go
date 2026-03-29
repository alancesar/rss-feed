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
	WatermillBroker struct {
		pubSub *gochannel.GoChannel
	}
)

func NewGoChannel() *gochannel.GoChannel {
	return gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 64,
			Persistent:          true,
		},
		watermill.NopLogger{},
	)
}

func NewWatermillBroker(pubSub *gochannel.GoChannel) *WatermillBroker {
	return &WatermillBroker{
		pubSub: pubSub,
	}
}

func (b *WatermillBroker) Publish(ctx context.Context, e event.Message) error {
	msg := message.NewMessageWithContext(ctx, watermill.NewUUID(), []byte(e.Payload))
	for k, v := range e.Headers {
		msg.Metadata.Set(k, fmt.Sprintf("%v", v))
	}

	return b.pubSub.Publish(string(e.Topic), msg)
}

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
					Nack: func(_ bool) {
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

func (b *WatermillBroker) Close() error {
	return b.pubSub.Close()
}
