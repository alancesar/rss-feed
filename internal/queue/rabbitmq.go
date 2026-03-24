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

	RabbitMQSubscriber struct {
		channel    *amqp.Channel
		deliveries chan event.Delivery
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

func (r RabbitMQ) Close() error {
	return r.conn.Close()
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, topic string, e event.Message) error {
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

func (c RabbitMQSubscriber) Subscribe(ctx context.Context, queue string) (<-chan event.Delivery, error) {
	deliveries, err := c.channel.ConsumeWithContext(ctx, queue, "", false, false, false, false, nil)
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
				Ack: func() error {
					return delivery.Ack(false)
				},
				Nack: func(requeue bool) error {
					return delivery.Nack(false, requeue)
				},
			}
		}
	}()

	return c.deliveries, nil
}

func (c RabbitMQSubscriber) Close() error {
	close(c.deliveries)
	return c.channel.Close()
}
