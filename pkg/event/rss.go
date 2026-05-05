package event

import (
	"rss-feed/pkg/rss"
	"time"
)

type (
	Feed struct {
		FeedID    string     `json:"feed_id"`
		Name      string     `json:"name"`
		URL       string     `json:"url"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
		Articles  []Article  `json:"articles"`
	}

	Image struct {
		ImageID   string `json:"image_id"`
		ArticleID string `json:"article_id"`
		URL       string `json:"path"`
	}

	Article struct {
		ArticleID   string    `json:"article_id"`
		FeedID      string    `json:"feed_id"`
		FeedName    string    `json:"feed_name"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Content     string    `json:"content"`
		Categories  []string  `json:"categories"`
		URL         string    `json:"url"`
		Image       Image     `json:"image"`
		PublishedAt time.Time `json:"published_at"`
	}
)

func (a Article) ToDomain() rss.Article {
	return rss.Article{
		ID:          a.ArticleID,
		Title:       a.Title,
		Description: a.Description,
		Content:     a.Content,
		Categories:  a.Categories,
		URL:         a.URL,
		PublishedAt: a.PublishedAt,
		Feed: rss.Feed{
			ID:   a.FeedID,
			Name: a.FeedName,
		},
	}
}

func (f Feed) ToDomain() rss.Feed {
	articles := make([]rss.Article, len(f.Articles))
	for i, article := range f.Articles {
		articles[i] = article.ToDomain()
	}

	return rss.Feed{
		ID:       f.FeedID,
		Name:     f.Name,
		URL:      f.URL,
		Articles: articles,
	}
}
