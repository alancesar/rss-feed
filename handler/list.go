package handler

import (
	"net/http"
	"rss-summary/presenter"
	"rss-summary/usecase"
	"time"

	"github.com/go-chi/render"
)

func GetFromDate(uc *usecase.ReadArticle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawDate := r.URL.Query().Get("date")
		date, err := time.Parse(time.DateOnly, rawDate)
		if err != nil {
			http.Error(w, "invalid date param. it must be in YYYY-MM-DD pattern", http.StatusInternalServerError)
			return
		}

		articles, err := uc.Execute(r.Context(), date)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		render.Status(r, http.StatusOK)
		if err := render.Render(w, r, presenter.NewArticleListResponse(articles)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func ListToday(uc *usecase.ReadArticle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		articles, err := uc.Execute(r.Context(), time.Now())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		render.Status(r, http.StatusOK)
		if err := render.Render(w, r, presenter.NewArticleListResponse(articles)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
