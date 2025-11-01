package presenter

import (
	"rss-summary/pkg/rss"
	"time"
)

type (
	ArticleResponse struct {
		Source      string    `json:"source"`
		Title       string    `json:"title"`
		URL         string    `json:"url"`
		PublishedAt time.Time `json:"published_at"`
	}
)

func NewArticleResponseFromDomain(article rss.Article) ArticleResponse {
	return ArticleResponse{
		Source:      article.Source,
		Title:       article.Title,
		URL:         article.URL,
		PublishedAt: time.Time(article.PublishedAt),
	}
}
