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
		pubSub     *gochannel.GoChannel
		deliveries chan event.Delivery
		cancel     context.CancelFunc
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
		pubSub:     pubSub,
		deliveries: make(chan event.Delivery),
	}
}

func (b *WatermillBroker) Publish(_ context.Context, topic string, e event.Message) error {
	msg := message.NewMessage(watermill.NewUUID(), []byte(e.Payload))
	for k, v := range e.Headers {
		msg.Metadata.Set(k, fmt.Sprintf("%v", v))
	}

	return b.pubSub.Publish(topic, msg)
}

func (b *WatermillBroker) Subscribe(ctx context.Context, queue string) (<-chan event.Delivery, error) {
	subCtx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	messages, err := b.pubSub.Subscribe(subCtx, queue)
	if err != nil {
		cancel()
		return nil, err
	}

	go func() {
		defer close(b.deliveries)
		for msg := range messages {
			headers := make(map[string]interface{}, len(msg.Metadata))
			for k, v := range msg.Metadata {
				headers[k] = v
			}

			b.deliveries <- event.Delivery{
				Message: event.Message{
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
		}
	}()

	return b.deliveries, nil
}

func (b *WatermillBroker) Close() error {
	if b.cancel != nil {
		b.cancel()
	}
	return nil
}
