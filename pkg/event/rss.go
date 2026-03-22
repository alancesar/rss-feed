package event

import "time"

type (
	Feed struct {
		FeedID   string    `json:"feed_id"`
		Name     string    `json:"name"`
		URL      string    `json:"url"`
		Articles []Article `json:"articles"`
	}

	Image struct {
		ImageID   string `json:"image_id"`
		ArticleID string `json:"article_id"`
		URL       string `json:"path"`
	}

	Article struct {
		ArticleID   string    `json:"article_id"`
		FeedID      string    `json:"feed_id"`
		Title       string    `json:"title"`
		URL         string    `json:"url"`
		Image       Image     `json:"image"`
		PublishedAt time.Time `json:"published_at"`
	}
)
