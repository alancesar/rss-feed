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
		PublishedAt time.Time `json:"published_at"`
	}

	ArticleListResponse struct {
		Articles []ArticleResponse `json:"articles"`
	}
)

func NewArticleListResponse(articles []rss.Article) *ArticleListResponse {
	output := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		output[i] = NewArticleResponseFromDomain(article)
	}

	return &ArticleListResponse{
		Articles: output,
	}
}

func (a ArticleListResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func NewArticleResponseFromDomain(article rss.Article) ArticleResponse {
	return ArticleResponse{
		Source:      article.Source,
		Title:       article.Title,
		URL:         article.URL,
		PublishedAt: time.Time(article.PublishedAt),
	}
}
