package main

import (
	"context"
	"crypto/tls"
	"log"
	"rss-summary/internal/database"
	"rss-summary/internal/feed"
	"rss-summary/internal/queue"
	"rss-summary/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open("rss.sqlite"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalln("while opening sqlite database:", err)
	}

	sqliteDatabase := database.NewGorm(db)

	cfg := &tls.Config{
		ServerName: "amqp.alancesar.org",
	}

	conn, err := amqp.DialTLS("amqps://rabbitmq:Pa55w0rd@amqp.alancesar.org", cfg)
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
