package main

import (
	"context"
	"log"
	"os"
	"rss-summary/internal/database"
	"rss-summary/internal/feed"
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

	sqliteDatabase := database.NewGorm(db)
	addSourceUseCase := usecase.NewAddFeed(sqliteDatabase, feed.NewGoFeed())

	sources := os.Args[1:]
	for _, source := range sources {
		if _, err := addSourceUseCase.Execute(ctx, source); err != nil {
			log.Println(err)
		}
	}
}
