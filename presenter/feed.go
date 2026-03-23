package presenter

import (
	"net/http"
)

type (
	AddFeedRequest struct {
		URL string `json:"url" example:"https://example.com/rss/feed"` // RSS feed URL
	}

	FeedResponse struct {
		ID   string `json:"id" example:"abc123"`                        // Unique identifier for the feed
		Name string `json:"name" example:"Some Feed"`                   // Name of the feed
		URL  string `json:"url" example:"https://example.com/rss/feed"` // RSS feed URL
	}
)

func (a FeedResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}
