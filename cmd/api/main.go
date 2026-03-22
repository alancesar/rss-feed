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
	"encoding/json"
	"log"
	"net/http"
	"os"
	"rss-summary/handler"
	"rss-summary/internal/database"
	"rss-summary/internal/feed"
	"rss-summary/internal/queue"
	"rss-summary/internal/storage"
	"rss-summary/usecase"

	_ "rss-summary/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	amqp "github.com/rabbitmq/amqp091-go"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
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

	dial, err := amqp.Dial(os.Getenv("AMQP_URL"))
	if err != nil {
		log.Fatalln("while connecting to rabbitmq:", err)
	}

	rabbitMQ := queue.NewRabbitMQ(dial)
	publisher, err := rabbitMQ.NewPublisher("rss")
	if err != nil {
		log.Fatalln("while creating rabbitmq publisher:", err)
	}

	readArticlesUseCase := usecase.NewReadArticles(sqliteDatabase, s3Storage)
	publishFeedUseCase := usecase.NewPublishFeed(publisher, feed.NewGoFeed())

	r := chi.NewRouter()
	r.Use(middleware.Logger)
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

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), r); err != nil {
		log.Fatalln(err)
	}
}
