package main

import (
	"context"
	"crypto/tls"
	"net/url"
	"os"
	"rss-feed/internal/database"
	"rss-feed/internal/feed"
	"rss-feed/internal/queue"
	"rss-feed/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	ctx := log.WithContext(context.Background())

	db, err := gorm.Open(sqlite.Open(os.Getenv("DB_PATH")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("opening sqlite database")
	}

	sqliteDatabase := database.NewGorm(db)

	amqpURL := os.Getenv("AMQP_URL")
	u, err := url.Parse(amqpURL)
	if err != nil {
		log.Fatal().Err(err).Msg("parsing AMQP_URL")
	}

	cfg := &tls.Config{
		ServerName: u.Hostname(),
	}

	conn, err := amqp.DialTLS(amqpURL, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("connecting to rabbitmq")
	}

	defer func() {
		_ = conn.Close()
	}()

	rabbitMQ := queue.NewRabbitMQ(conn)
	publisher, err := rabbitMQ.NewPublisher("rss")
	if err != nil {
		log.Fatal().Err(err).Msg("creating rabbitmq publisher")
	}

	updateFeedUseCase := usecase.NewUpdateFeeds(sqliteDatabase, feed.NewGoFeed(), publisher)
	if err := updateFeedUseCase.Execute(ctx); err != nil {
		log.Error().Err(err).Msg("updating feeds")
	}
}
