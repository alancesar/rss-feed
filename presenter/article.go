package presenter

import (
	"rss-summary/pkg/rss"
	"time"
)

type (
	Article struct {
		Source      string    `json:"source"`
		Title       string    `json:"title"`
		URL         string    `json:"url"`
		PublishedAt time.Time `json:"published_at"`
	}
)

func NewArticleFromDomain(article rss.Article) Article {
	return Article{
		Source:      article.Source,
		Title:       article.Title,
		URL:         article.URL,
		PublishedAt: time.Time(article.PublishedAt),
	}
}
