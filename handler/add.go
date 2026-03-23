package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"rss-feed/pkg/rss"
	"rss-feed/presenter"
	"rss-feed/usecase"

	"github.com/go-chi/render"
)

// AddFeed godoc
//
//	@Summary		Add a new RSS feed
//	@Description	Fetches the RSS feed at the given URL and persists it along with its articles
//	@Tags			feeds
//	@Accept			json
//	@Produce		json
//	@Param			request	body		presenter.AddFeedRequest	true	"Feed URL"
//	@Success		201		{object}	presenter.FeedResponse
//	@Failure		400		{string}	string	"invalid request body"
//	@Failure		500		{string}	string	"internal server error"
//	@Router			/feeds [post]
func AddFeed(uc *usecase.SaveFeed) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request presenter.AddFeedRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		feed, err := uc.Execute(r.Context(), request.URL)
		if err != nil {
			if errors.Is(err, rss.ErrFeedAlreadyExists) {
				http.Error(w, "feed already exists", http.StatusConflict)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		render.Status(r, http.StatusCreated)
		if err := render.Render(w, r, presenter.FeedResponse{
			ID:   feed.ID,
			Name: feed.Name,
			URL:  feed.URL,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
