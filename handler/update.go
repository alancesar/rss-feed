package handler

import (
	"net/http"
	"rss-feed/pkg/event"
	"rss-feed/usecase"
)

// TriggerFeedUpdate godoc
//
//	@Summary		Trigger a feed update
//	@Description	Publishes a job event that instructs the worker to re-fetch all registered feeds immediately
//	@Tags			feeds
//	@Success		202
//	@Failure		500	{string}	string	"internal server error"
//	@Router			/feeds/update [post]
func TriggerFeedUpdate(publisher usecase.Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := publisher.Publish(r.Context(), "feed.jobs", event.Event{
			Payload: event.Job{Command: event.CommandUpdateFeeds},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
