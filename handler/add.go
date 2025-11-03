package handler

import (
	"encoding/json"
	"net/http"
	"rss-summary/presenter"
	"rss-summary/usecase"
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

		response := presenter.NewAddFeedResponseFromDomain(feed)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	}
}
