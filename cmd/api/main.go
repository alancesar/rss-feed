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
	"log"
	"net/http"
	"rss-summary/handler"
	"rss-summary/internal/database"
	"rss-summary/internal/feed"
	"rss-summary/internal/storage"
	"rss-summary/usecase"

	_ "rss-summary/docs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db, err := gorm.Open(sqlite.Open("rss.sqlite"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalln("while opening sqlite database:", err)
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
	)

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://s3.alancesar.org")
		o.UsePathStyle = true
	})

	s3Storage, err := storage.NewS3(s3Client)
	if err != nil {
		log.Fatalln("while creating s3 client:", err)
	}

	sqliteDatabase := database.NewGorm(db)
	readUseCase := usecase.NewRead(sqliteDatabase, s3Storage)
	addUseCase := usecase.NewAddFeed(sqliteDatabase, feed.NewGoFeed(s3Storage, http.DefaultClient))

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
		r.Get("/", handler.GetFromDate(readUseCase))
		r.Get("/today", handler.ListToday(readUseCase))
	})

	r.Route("/feeds", func(r chi.Router) {
		r.Post("/", handler.AddFeed(addUseCase))
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatalln(err)
	}
}
