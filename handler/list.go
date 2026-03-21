package handler

import (
	"net/http"
	"rss-summary/presenter"
	"rss-summary/usecase"
	"time"

	"github.com/go-chi/render"
)

// GetFromDate godoc
//
//	@Summary		List articles by date
//	@Description	Returns all articles published on the given date
//	@Tags			articles
//	@Produce		json
//	@Param			date	query		string						true	"Date in YYYY-MM-DD format"	example(2024-01-15)
//	@Success		200		{object}	presenter.ArticleListResponse
//	@Failure		400		{string}	string	"invalid date param. it must be in YYYY-MM-DD pattern"
//	@Failure		500		{string}	string	"internal server error"
//	@Router			/articles [get]
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

// ListToday godoc
//
//	@Summary		List today's articles
//	@Description	Returns all articles published today
//	@Tags			articles
//	@Produce		json
//	@Success		200	{object}	presenter.ArticleListResponse
//	@Failure		500	{string}	string	"internal server error"
//	@Router			/articles/today [get]
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
