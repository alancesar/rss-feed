package handler

import (
	"net/http"
	"rss-feed/presenter"
	"rss-feed/usecase"

	"github.com/go-chi/render"
	"github.com/rs/zerolog"
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

		logger := zerolog.Ctx(r.Context())

		articles, err := uc.Execute(r.Context(), term)
		if err != nil {
			logger.Error().Err(err).Msg("failed to find articles")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		render.Status(r, http.StatusOK)
		if err := render.Render(w, r, presenter.NewArticleListResponse(articles)); err != nil {
			logger.Error().Err(err).Msg("failed to render articles response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
