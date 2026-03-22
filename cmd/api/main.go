// Package main RSS Summary API.
//
//	@title			RSS Summary API
//	@version		1.0
//	@description	Aggregates articles from RSS feeds and exposes them via a REST API.
//
//	@host		localhost:3000
//	@BasePath	/
//
//go:generate swag init -g cmd/api/main.go -o docs
package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"rss-feed/handler"
	"rss-feed/internal/database"
	"rss-feed/internal/feed"
	"rss-feed/internal/queue"
	"rss-feed/internal/storage"
	"rss-feed/usecase"

	_ "rss-feed/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

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

	dial, err := amqp.Dial(os.Getenv("AMQP_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("connecting to rabbitmq")
	}

	rabbitMQ := queue.NewRabbitMQ(dial)
	publisher, err := rabbitMQ.NewPublisher("rss")
	if err != nil {
		log.Fatal().Err(err).Msg("creating rabbitmq publisher")
	}

	readArticlesUseCase := usecase.NewReadArticles(sqliteDatabase, s3Storage)
	publishFeedUseCase := usecase.NewPublishFeed(publisher, feed.NewGoFeed())

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			zerolog.Ctx(r.Context()).Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Msg("request")
		})
	})
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost*"},
	}))

	render.Respond = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(v)
	}

	r.Route("/articles", func(r chi.Router) {
		r.Get("/", handler.GetFromDate(readArticlesUseCase))
		r.Get("/today", handler.ListToday(readArticlesUseCase))
	})

	r.Route("/feeds", func(r chi.Router) {
		r.Post("/", handler.AddFeed(publishFeedUseCase))
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	port := os.Getenv("PORT")
	log.Info().Str("port", port).Msg("server listening")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
		BaseContext: func(_ net.Listener) context.Context {
			return log.WithContext(context.Background())
		},
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
