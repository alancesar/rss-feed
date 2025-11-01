package handler

import (
	"encoding/json"
	"net/http"
	"rss-summary/presenter"
	"rss-summary/usecase"
	"time"
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

		output := make([]presenter.ArticleResponse, len(articles))
		for i, article := range articles {
			output[i] = presenter.NewArticleResponseFromDomain(article)
		}

		if err := json.NewEncoder(w).Encode(output); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func ListToday(uc *usecase.ReadArticle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		articles, err := uc.Execute(r.Context(), time.Now())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		output := make([]presenter.ArticleResponse, len(articles))
		for i, article := range articles {
			output[i] = presenter.NewArticleResponseFromDomain(article)
		}

		if err := json.NewEncoder(w).Encode(output); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
