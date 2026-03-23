package presenter

import (
	"net/http"
	"rss-feed/pkg/rss"
	"time"
)

type (
	ArticleResponse struct {
		Title       string          `json:"title"`
		URL         string          `json:"url"`
		Feed        FeedResponse    `json:"feed"`
		Images      []ImageResponse `json:"images"`
		PublishedAt time.Time       `json:"published_at" example:"2006-01-02T15:04:05-07:00" format:"date-time"` // RFC3339 timestamp with timezone offset
	}

	ImageResponse struct {
		URL  string `json:"url"`
		Type string `json:"type"`
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

func NewImageResponseFromDomain(image rss.Image) ImageResponse {
	return ImageResponse{
		URL:  image.Path,
		Type: string(image.Type),
	}
}

func (a ArticleListResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func NewArticleResponseFromDomain(article rss.Article) ArticleResponse {
	images := make([]ImageResponse, len(article.Images))
	for i, img := range article.Images {
		images[i] = NewImageResponseFromDomain(img)
	}

	return ArticleResponse{
		Title: article.Title,
		URL:   article.URL,
		Feed: FeedResponse{
			ID:   article.Feed.ID,
			Name: article.Feed.Name,
			URL:  article.Feed.URL,
		},
		Images:      images,
		PublishedAt: article.PublishedAt,
	}
}
