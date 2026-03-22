package presenter

import (
	"net/http"
)

type (
	AddFeedRequest struct {
		URL string `json:"url" example:"https://example.com/rss/feed"` // RSS feed URL
	}

	AddFeedResponse struct {
		ID   string `json:"id" example:"abc123"`                        // Unique identifier for the feed
		Name string `json:"name" example:"Some Feed"`                   // Name of the feed
		URL  string `json:"url" example:"https://example.com/rss/feed"` // RSS feed URL
	}
)

func (a AddFeedResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}
