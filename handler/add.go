package handler

import (
	"encoding/json"
	"net/http"
	"rss-summary/presenter"
	"rss-summary/usecase"

	"github.com/go-chi/render"
)

func AddFeed(uc *usecase.AddFeed) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request presenter.AddFeedRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		feed, err := uc.Execute(r.Context(), request.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		render.Status(r, http.StatusCreated)
		if err := render.Render(w, r, presenter.NewAddFeedResponseFromDomain(feed)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
