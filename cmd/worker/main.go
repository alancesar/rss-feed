package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"rss-feed/internal/database"
	"rss-feed/internal/feed"
	"rss-feed/internal/queue"
	"rss-feed/internal/storage"
	"rss-feed/pkg/event"
	"rss-feed/usecase"
	"time"

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

	s3Storage, err := storage.NewS3(os.Getenv("S3_ENDPOINT"), os.Getenv("S3_REGION"), os.Getenv("AWS_BUCKET"))
	if err != nil {
		log.Fatal().Err(err).Msg("creating s3 client")
	}

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
	feedSubscriber, err := rabbitMQ.NewSubscriber()
	if err != nil {
		log.Fatal().Err(err).Msg("creating rabbitmq subscriber")
	}

	imageSubscriber, err := rabbitMQ.NewSubscriber()
	if err != nil {
		log.Fatal().Err(err).Msg("creating rabbitmq subscriber")
	}

	jobsConsumer, err := rabbitMQ.NewSubscriber()
	if err != nil {
		log.Fatal().Err(err).Msg("creating rss.feed.jobs consumer")
	}

	imagePublisher, err := rabbitMQ.NewPublisher("rss")
	if err != nil {
		log.Fatal().Err(err).Msg("creating rabbitmq publisher")
	}

	consumeFeedUseCase := usecase.NewConsumeFeed(sqliteDatabase, feedSubscriber, imagePublisher)
	consumeImageUseCase := usecase.NewConsumeImage(http.DefaultClient, imageSubscriber, s3Storage, sqliteDatabase)
	updateFeedsUseCase := usecase.NewUpdateFeeds(sqliteDatabase, feed.NewGoFeed(), imagePublisher)

	updateInterval := 30 * time.Minute
	if v := os.Getenv("UPDATE_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			updateInterval = d
		} else {
			log.Warn().Str("value", v).Msg("invalid UPDATE_INTERVAL, using default 30m")
		}
	}

	forceUpdate := make(chan struct{}, 1)

	log.Info().Msg("worker started")

	go func() {
		if err := consumeFeedUseCase.Execute(ctx); err != nil {
			log.Fatal().Err(err).Msg("consuming rss.feed.article.found events")
		}
	}()

	go func() {
		if err := consumeImageUseCase.Execute(ctx); err != nil {
			log.Fatal().Err(err).Msg("consuming rss.feed.article.image events")
		}
	}()

	go func() {
		deliveries, err := jobsConsumer.Subscribe(ctx, "rss.feed.jobs")
		if err != nil {
			log.Fatal().Err(err).Msg("consuming rss.feed.jobs events")
		}

		for delivery := range deliveries {
			var j event.Job
			if err := json.Unmarshal(delivery.Payload, &j); err != nil {
				log.Fatal().Err(err).Msg("unmarshalling rss.feed.job event")
				if err := delivery.Nack(false); err != nil {
					log.Fatal().Err(err).Msg("acknowledging delivery")
				}
				continue
			}

			switch j.Command {
			case event.CommandUpdateFeeds:
				select {
				case forceUpdate <- struct{}{}:
				default:
				}
			}
		}

	}()

	log.Info().Dur("interval", updateInterval).Msg("starting feed update ticker")
	if err := updateFeedsUseCase.Execute(ctx); err != nil {
		log.Error().Err(err).Msg("updating feeds")
	}

	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := updateFeedsUseCase.Execute(ctx); err != nil {
				log.Error().Err(err).Msg("updating feeds")
			}
		case <-forceUpdate:
			log.Info().Msg("forced feed update triggered")
			if err := updateFeedsUseCase.Execute(ctx); err != nil {
				log.Error().Err(err).Msg("updating feeds")
			}
			ticker.Reset(updateInterval)
		}
	}
}
