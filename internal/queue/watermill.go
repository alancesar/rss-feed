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
	WatermillPublisher struct {
		pubSub *gochannel.GoChannel
	}

	WatermillSubscriber struct {
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

func NewWatermillPublisher(pubSub *gochannel.GoChannel) *WatermillPublisher {
	return &WatermillPublisher{pubSub: pubSub}
}

func NewWatermillSubscriber(pubSub *gochannel.GoChannel) *WatermillSubscriber {
	return &WatermillSubscriber{
		pubSub:     pubSub,
		deliveries: make(chan event.Delivery),
	}
}

func (p *WatermillPublisher) Publish(_ context.Context, topic string, e event.Message) error {
	msg := message.NewMessage(watermill.NewUUID(), []byte(e.Payload))
	for k, v := range e.Headers {
		msg.Metadata.Set(k, fmt.Sprintf("%v", v))
	}

	return p.pubSub.Publish(topic, msg)
}

func (s *WatermillSubscriber) Subscribe(ctx context.Context, queue string) (<-chan event.Delivery, error) {
	subCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	messages, err := s.pubSub.Subscribe(subCtx, queue)
	if err != nil {
		cancel()
		return nil, err
	}

	go func() {
		defer close(s.deliveries)
		for msg := range messages {
			headers := make(map[string]interface{}, len(msg.Metadata))
			for k, v := range msg.Metadata {
				headers[k] = v
			}

			s.deliveries <- event.Delivery{
				Message: event.Message{
					Headers: headers,
					Payload: event.Payload(msg.Payload),
				},
				Ack: func() error {
					msg.Ack()
					return nil
				},
				Nack: func(_ bool) error {
					msg.Nack()
					return nil
				},
			}
		}
	}()

	return s.deliveries, nil
}

func (s *WatermillSubscriber) Close() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}
