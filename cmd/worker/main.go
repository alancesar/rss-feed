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
	feedConsumer, err := rabbitMQ.NewConsumer("rss.feed.article.found")
	if err != nil {
		log.Fatal().Err(err).Msg("creating rss.feed.article.found consumer")
	}

	imagesConsumer, err := rabbitMQ.NewConsumer("rss.feed.article.image.found")
	if err != nil {
		log.Fatal().Err(err).Msg("creating rss.feed.article.image.found consumer")
	}

	imagePublisher, err := rabbitMQ.NewPublisher("rss")
	if err != nil {
		log.Fatal().Err(err).Msg("creating rabbitmq publisher")
	}

	consumeFeedUseCase := usecase.NewConsumeFeed(sqliteDatabase, imagePublisher)
	consumeImageUseCase := usecase.NewConsumeImage(http.DefaultClient, s3Storage, sqliteDatabase)
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

	jobsConsumer, err := rabbitMQ.NewConsumer("rss.feed.jobs")
	if err != nil {
		log.Fatal().Err(err).Msg("creating rss.feed.jobs consumer")
	}

	log.Info().Msg("worker started")

	go func() {
		if err := feedConsumer.Consume(ctx, func(ctx context.Context, body []byte) error {
			var e event.Feed
			if err := json.Unmarshal(body, &e); err != nil {
				return err
			}

			return consumeFeedUseCase.Execute(ctx, e)
		}); err != nil {
			log.Fatal().Err(err).Msg("consuming rss.feed.article.found events")
		}
	}()

	go func() {
		if err := imagesConsumer.Consume(ctx, func(ctx context.Context, body []byte) error {
			var e event.Image
			if err := json.Unmarshal(body, &e); err != nil {
				return err
			}

			return consumeImageUseCase.Execute(ctx, e)
		}); err != nil {
			log.Fatal().Err(err).Msg("consuming rss.feed.article.image.found events")
		}
	}()

	go func() {
		if err := jobsConsumer.Consume(ctx, func(_ context.Context, body []byte) error {
			var j event.Job
			if err := json.Unmarshal(body, &j); err != nil {
				return err
			}

			if j.Command == event.CommandUpdateFeeds {
				select {
				case forceUpdate <- struct{}{}:
				default:
				}
			}

			return nil
		}); err != nil {
			log.Fatal().Err(err).Msg("consuming rss.feed.jobs events")
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
