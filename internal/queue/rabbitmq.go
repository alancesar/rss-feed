// Package queue provides message broker adapters: RabbitMQ over AMQP 0-9-1 and an
// in-memory Watermill broker.
package queue

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"

	amqp "github.com/rabbitmq/amqp091-go"
)

type (
	// RabbitMQ wraps an AMQP connection and creates publishers and subscribers on it.
	RabbitMQ struct {
		conn *amqp.Connection
	}

	// RabbitMQPublisher publishes messages to an exchange through its own AMQP channel.
	RabbitMQPublisher struct {
		channel  *amqp.Channel
		exchange string
	}

	// RabbitMQSubscriber consumes queues through its own AMQP channel and exposes the messages
	// as deliveries.
	RabbitMQSubscriber struct {
		channel    *amqp.Channel
		deliveries chan event.Delivery
	}
)

// NewRabbitMQ creates a RabbitMQ on top of an established AMQP connection.
func NewRabbitMQ(conn *amqp.Connection) RabbitMQ {
	return RabbitMQ{
		conn: conn,
	}
}

// NewPublisher opens a channel and returns a publisher that sends to the given exchange.
func (r RabbitMQ) NewPublisher(exchange string) (*RabbitMQPublisher, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQPublisher{
		exchange: exchange,
		channel:  ch,
	}, nil
}

// NewSubscriber opens a channel and returns a subscriber on it.
func (r RabbitMQ) NewSubscriber() (*RabbitMQSubscriber, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQSubscriber{
		channel:    ch,
		deliveries: make(chan event.Delivery),
	}, nil
}

// Close closes the underlying AMQP connection.
func (r RabbitMQ) Close() error {
	return r.conn.Close()
}

// Publish JSON-encodes the message payload and publishes it to the exchange, using the
// message topic as routing key and its headers as AMQP headers.
func (p *RabbitMQPublisher) Publish(ctx context.Context, e event.Message) error {
	body, err := json.Marshal(e.Payload)
	if err != nil {
		return err
	}

	if err := p.channel.PublishWithContext(ctx, p.exchange, string(e.Topic), false, false, amqp.Publishing{
		Headers: e.Headers,
		Body:    body,
	}); err != nil {
		return err
	}

	return err
}

// Subscribe consumes the queue named after the topic, with manual acknowledgements, and
// forwards each message as a delivery. Ack acknowledges the message; Nack rejects it and
// optionally requeues it. Every call shares the subscriber's single deliveries channel.
func (c RabbitMQSubscriber) Subscribe(ctx context.Context, topic event.Topic) (<-chan event.Delivery, error) {
	deliveries, err := c.channel.ConsumeWithContext(ctx, string(topic), "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		for delivery := range deliveries {
			c.deliveries <- event.Delivery{
				Message: event.Message{
					Headers: delivery.Headers,
					Payload: delivery.Body,
				},
				Ack: func() {
					_ = delivery.Ack(false)
				},
				Nack: func(requeue bool) {
					_ = delivery.Nack(false, requeue)
				},
			}
		}
	}()

	return c.deliveries, nil
}

// Close closes the deliveries channel and the AMQP channel.
func (c RabbitMQSubscriber) Close() error {
	close(c.deliveries)
	return c.channel.Close()
}
