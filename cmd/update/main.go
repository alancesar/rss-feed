package main

import (
	"context"
	"log"
	"net/http"
	"rss-summary/internal/database"
	"rss-summary/internal/feed"
	"rss-summary/internal/storage"
	"rss-summary/usecase"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	addSourceUseCase := usecase.NewUpdateFeeds(sqliteDatabase, feed.NewGoFeed(s3Storage, http.DefaultClient))

	if err := addSourceUseCase.Execute(ctx); err != nil {
		log.Println(err)
	}
}
