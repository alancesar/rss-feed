package presenter

import (
	"net/http"
	"rss-summary/pkg/rss"
	"time"
)

type (
	ArticleResponse struct {
		Source      string    `json:"source"`
		Title       string    `json:"title"`
		URL         string    `json:"url"`
		ImageURL    string    `json:"image_url"`
		PublishedAt time.Time `json:"published_at" example:"2006-01-02T15:04:05-07:00" format:"date-time"` // RFC3339 timestamp with timezone offset
	}

	ArticleListResponse struct {
		Articles []ArticleResponse `json:"articles"`
	}
)

func NewArticleListResponse(articles []rss.Article) *ArticleListResponse {
	output := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		output[i] = NewArticleResponseFromDomain(article, article.ImagePath())
	}

	return &ArticleListResponse{
		Articles: output,
	}
}

func (a ArticleListResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func NewArticleResponseFromDomain(article rss.Article, imageURL string) ArticleResponse {
	return ArticleResponse{
		Source:      article.Source,
		Title:       article.Title,
		URL:         article.URL,
		ImageURL:    imageURL,
		PublishedAt: time.Time(article.PublishedAt),
	}
}
