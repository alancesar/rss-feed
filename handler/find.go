package handler

import (
	"net/http"
	"rss-feed/presenter"
	"rss-feed/usecase"

	"github.com/go-chi/render"
)

// FindArticles godoc
//
//	@Summary		Find articles by term
//	@Description	Returns articles matching the given search term via vector search
//	@Tags			articles
//	@Produce		json
//	@Param			q	query		string						true	"Search term"
//	@Success		200	{object}	presenter.ArticleListResponse
//	@Failure		400	{string}	string	"missing q param"
//	@Failure		500	{string}	string	"internal server error"
//	@Router			/articles/search [get]
func FindArticles(uc *usecase.Find) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		term := r.URL.Query().Get("q")
		if term == "" {
			http.Error(w, "missing q param", http.StatusBadRequest)
			return
		}

		articles, err := uc.Execute(r.Context(), term)
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
