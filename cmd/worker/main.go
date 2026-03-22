package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"rss-summary/internal/database"
	"rss-summary/internal/queue"
	"rss-summary/internal/storage"
	"rss-summary/pkg/event"
	"rss-summary/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open(os.Getenv("DB_PATH")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalln("while opening sqlite database:", err)
	}
	sqliteDatabase := database.NewGorm(db)

	s3Storage, err := storage.NewS3(os.Getenv("S3_ENDPOINT"), os.Getenv("S3_REGION"), os.Getenv("AWS_BUCKET"))
	if err != nil {
		log.Fatalln("while creating s3 client:", err)
	}

	amqpURL := os.Getenv("AMQP_URL")
	u, err := url.Parse(amqpURL)
	if err != nil {
		log.Fatalln("while parsing AMQP_URL:", err)
	}

	cfg := &tls.Config{
		ServerName: u.Hostname(),
	}

	conn, err := amqp.DialTLS(amqpURL, cfg)
	if err != nil {
		log.Fatalln("while connecting to rabbitmq:", err)
	}

	defer func() {
		_ = conn.Close()
	}()

	rabbitMQ := queue.NewRabbitMQ(conn)
	feedConsumer, err := rabbitMQ.NewConsumer("rss.feed.found")
	if err != nil {
		log.Fatalln("while creating rabbitmq publisher:", err)
	}

	imagesConsumer, err := rabbitMQ.NewConsumer("rss.feed.article.image.found")
	if err != nil {
		log.Fatalln("while creating rabbitmq publisher:", err)
	}

	imagePublisher, err := rabbitMQ.NewPublisher("rss")
	if err != nil {
		log.Fatalln("while creating rabbitmq publisher:", err)
	}

	consumeFeedUseCase := usecase.NewConsumeFeed(sqliteDatabase, imagePublisher)
	consumeImageUseCase := usecase.NewConsumeImage(http.DefaultClient, s3Storage, sqliteDatabase)

	go func() {
		if err := feedConsumer.Consume(ctx, func(ctx context.Context, body []byte) error {
			var e event.Feed
			if err := json.Unmarshal(body, &e); err != nil {
				return err
			}

			return consumeFeedUseCase.Execute(ctx, e)
		}); err != nil {
			log.Fatalln("while consuming feed.found events:", err)
		}
	}()

	if err := imagesConsumer.Consume(ctx, func(ctx context.Context, body []byte) error {
		var e event.Image
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}

		return consumeImageUseCase.Execute(ctx, e)
	}); err != nil {
		log.Fatalln("while consuming feed.article.image.found events:", err)
	}
}
