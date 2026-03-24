package queue

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"

	amqp "github.com/rabbitmq/amqp091-go"
)

type (
	RabbitMQ struct {
		conn *amqp.Connection
	}

	Handler func(context.Context, []byte) error

	RabbitMQPublisher struct {
		channel  *amqp.Channel
		exchange string
	}

	RabbitMQConsumer struct {
		channel  *amqp.Channel
		queue    string
		messages chan event.Message
	}
)

func NewRabbitMQ(conn *amqp.Connection) RabbitMQ {
	return RabbitMQ{
		conn: conn,
	}
}

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

func (r RabbitMQ) NewConsumer(queue string) (*RabbitMQConsumer, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQConsumer{
		channel:  ch,
		queue:    queue,
		messages: make(chan event.Message),
	}, nil
}

func (r RabbitMQ) Close() error {
	return r.conn.Close()
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, topic string, e event.Event) error {
	body, err := json.Marshal(e.Payload)
	if err != nil {
		return err
	}

	if err := p.channel.PublishWithContext(ctx, p.exchange, topic, false, false, amqp.Publishing{
		Headers: e.Headers,
		Body:    body,
	}); err != nil {
		return err
	}

	return err
}

func (c RabbitMQConsumer) Subscribe(ctx context.Context, queue string) (<-chan event.Message, error) {
	deliveries, err := c.channel.ConsumeWithContext(ctx, queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	go func() {
		for delivery := range deliveries {
			c.messages <- event.Message{
				Headers: delivery.Headers,
				Payload: delivery.Body,
				Ack: func() error {
					return delivery.Ack(false)
				},
				Nack: func(requeue bool) error {
					return delivery.Nack(false, requeue)
				},
			}
		}
	}()

	return c.messages, nil
}

func (c RabbitMQConsumer) Consume(ctx context.Context, handler Handler) error {
	defer func() {
		_ = c.channel.Close()
	}()

	deliveries, err := c.channel.ConsumeWithContext(ctx, c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		if err := handler(ctx, delivery.Body); err != nil {
			_ = delivery.Nack(false, true)
		} else {
			_ = delivery.Ack(false)
		}
	}

	return nil
}

func (c RabbitMQConsumer) Close() error {
	close(c.messages)
	return c.channel.Close()
}
