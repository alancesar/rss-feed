package queue

import (
	"context"
	"encoding/json"
	"rss-feed/pkg/event"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
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
		channel *amqp.Channel
		queue   string
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
		channel: ch,
		queue:   queue,
	}, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, topic string, e event.Event) error {
	body, err := json.Marshal(e.Payload)
	if err != nil {
		return err
	}

	zerolog.Ctx(ctx).Info().Str("exchange", p.exchange).Str("topic", topic).Msg("publishing event")
	if err := p.channel.PublishWithContext(ctx, p.exchange, topic, false, false, amqp.Publishing{
		Headers: e.Headers,
		Body:    body,
	}); err != nil {
		return err
	}

	return err
}

func (c RabbitMQConsumer) Consume(ctx context.Context, handler Handler) error {
	defer func() {
		_ = c.channel.Close()
	}()

	deliveries, err := c.channel.ConsumeWithContext(ctx, c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log := zerolog.Ctx(ctx)
	log.Info().Str("queue", c.queue).Msg("starting consumer")
	for delivery := range deliveries {
		if err := handler(ctx, delivery.Body); err != nil {
			log.Error().Err(err).Str("queue", c.queue).Msg("message processing failed, nacking")
			_ = delivery.Nack(false, true)
		} else {
			_ = delivery.Ack(false)
		}
	}

	return nil
}
