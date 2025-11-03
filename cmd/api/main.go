package main

import (
	"log"
	"net/http"
	"rss-summary/handler"
	"rss-summary/internal/database"
	"rss-summary/internal/feed"
	"rss-summary/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	sqliteDatabase := database.NewGorm(db)
	readUseCase := usecase.NewRead(sqliteDatabase)
	addUseCase := usecase.NewAddFeed(sqliteDatabase, feed.NewGoFeed())

	r.Route("/articles", func(r chi.Router) {
		r.Get("/", handler.GetFromDate(readUseCase))
		r.Get("/today", handler.ListToday(readUseCase))
	})

	r.Route("/feeds", func(r chi.Router) {
		r.Post("/", handler.AddFeed(addUseCase))
	})

	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatalln(err)
	}
}
