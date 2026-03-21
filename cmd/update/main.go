package main

import (
	"context"
	"log"
	"rss-summary/internal/database"
	"rss-summary/internal/feed"
	"rss-summary/internal/storage"
	"rss-summary/usecase"

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

	s3, err := storage.NewS3(ctx)
	if err != nil {
		log.Fatalln("while creating s3 client:", err)
	}

	sqliteDatabase := database.NewGorm(db)
	addSourceUseCase := usecase.NewUpdateFeeds(sqliteDatabase, feed.NewGoFeed(), s3)

	if err := addSourceUseCase.Execute(ctx); err != nil {
		log.Println(err)
	}
}
