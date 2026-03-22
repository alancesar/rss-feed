package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/url"
	"os"
	"rss-feed/internal/database"
	"rss-feed/internal/feed"
	"rss-feed/internal/queue"
	"rss-feed/usecase"

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
	publisher, err := rabbitMQ.NewPublisher("rss")
	if err != nil {
		log.Fatalln("while creating rabbitmq publisher:", err)
	}

	feedPublisherUseCase := usecase.NewPublishFeed(publisher, feed.NewGoFeed())
	updateFeedUseCase := usecase.NewUpdateFeeds(sqliteDatabase, feedPublisherUseCase.Execute)
	if err := updateFeedUseCase.Execute(ctx); err != nil {
		log.Println(err)
	}
}
