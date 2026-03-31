// Package main RSS Summary API.
//
//	@title			RSS Summary API
//	@version		1.0
//	@description	Aggregates articles from RSS feeds and exposes them via a REST API.
//
//	@host		localhost:8080
//	@BasePath	/
//
//go:generate swag init -g cmd/api/main.go -o docs
package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"rss-feed/handler"
	"rss-feed/internal/database"
	"rss-feed/internal/feed"
	"rss-feed/internal/queue"
	"rss-feed/internal/storage"
	"rss-feed/usecase"
	"time"

	_ "rss-feed/docs"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
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

	client, err := chroma.NewHTTPClient(
		chroma.WithBaseURL("https://chroma.alancesar.org"),
	)

	defer func() { _ = client.Close() }()

	chromaDatabase, err := database.NewChroma(ctx, client)
	if err != nil {
		log.Fatal().Err(err).Msg("creating chroma database")
	}

	pubSub := queue.NewGoChannel()
	broker := queue.NewWatermillBroker(pubSub)

	readArticlesUseCase := usecase.NewReadArticles(sqliteDatabase, s3Storage)
	saveFeedUseCase := usecase.NewSaveFeed(feed.NewGoFeed(), sqliteDatabase, broker)
	findArticlesUseCase := usecase.NewFind(chromaDatabase, sqliteDatabase)
	handleFeedUseCase := usecase.NewHandleFeed(broker)
	consumeImageUseCase := usecase.NewConsumeImage(http.DefaultClient, broker, s3Storage, sqliteDatabase)
	handleArticleUseCase := usecase.NewArticle(broker, chromaDatabase)
	updateFeedsUseCase := usecase.NewUpdateFeeds(sqliteDatabase, feed.NewGoFeed(), broker)

	updateInterval := 30 * time.Minute
	if v := os.Getenv("UPDATE_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			updateInterval = d
		} else {
			log.Warn().Str("value", v).Msg("invalid UPDATE_INTERVAL, using default 30m")
		}
	}

	go func() {
		if err := handleFeedUseCase.Execute(ctx); err != nil {
			log.Fatal().Err(err).Msg("consuming feed articles")
		}
	}()

	go func() {
		if err := consumeImageUseCase.Execute(ctx); err != nil {
			log.Fatal().Err(err).Msg("consuming article images")
		}
	}()

	go func() {
		if err := handleArticleUseCase.Execute(ctx); err != nil {
			log.Fatal().Err(err).Msg("handling feed articles")
		}
	}()

	go func() {
		if err := updateFeedsUseCase.Execute(ctx, updateInterval); err != nil {
			log.Fatal().Err(err).Msg("updating feeds")
		}
	}()

	r := chi.NewRouter()
	r.Use(render.SetContentType(render.ContentTypeJSON))
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

	r.Route("/articles", func(r chi.Router) {
		r.Get("/", handler.GetFromDate(readArticlesUseCase))
		r.Get("/today", handler.ListToday(readArticlesUseCase))
		r.Get("/search", handler.FindArticles(findArticlesUseCase))
	})

	r.Route("/feeds", func(r chi.Router) {
		r.Post("/", handler.AddFeed(saveFeedUseCase))
		r.Post("/update", handler.TriggerFeedUpdate(broker))
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
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
