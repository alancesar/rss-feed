// Package presenter defines the JSON request and response DTOs of the HTTP API and the
// conversions from domain entities to them.
package presenter

import (
	"net/http"
	"rss-feed/pkg/rss"
)

type (
	// AddFeedRequest is the JSON body for registering a feed.
	AddFeedRequest struct {
		URL string `json:"url" example:"https://example.com/rss/feed"` // RSS feed URL
	}

	// FeedResponse is the JSON representation of a feed.
	FeedResponse struct {
		ID   string `json:"id" example:"abc123"`                        // Unique identifier for the feed
		Name string `json:"name" example:"Some Feed"`                   // Name of the feed
		URL  string `json:"url" example:"https://example.com/rss/feed"` // RSS feed URL
	}
)

// NewFeedResponse converts a domain feed into a FeedResponse.
func NewFeedResponse(feed rss.Feed) FeedResponse {
	return FeedResponse{
		ID:   feed.ID,
		Name: feed.Name,
		URL:  feed.URL,
	}
}

// Render implements the go-chi render.Renderer interface. It performs no post-processing.
func (a FeedResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}
