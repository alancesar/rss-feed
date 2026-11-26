package presenter

import (
	"net/http"
	"rss-feed/pkg/rss"
	"time"
)

type (
	// ArticleResponse is the JSON representation of an article, with its feed and images.
	ArticleResponse struct {
		Title       string          `json:"title" example:"Some Post"`                                           // Article title
		URL         string          `json:"url" example:"https://example.com/posts/some-post"`                   // Link to the original article
		Feed        FeedResponse    `json:"feed"`                                                                // Feed the article belongs to
		Images      []ImageResponse `json:"images"`                                                              // Images attached to the article
		PublishedAt time.Time       `json:"published_at" example:"2006-01-02T15:04:05-07:00" format:"date-time"` // RFC3339 timestamp with timezone offset
	}

	// ImageResponse is the JSON representation of an article image: its location and type.
	ImageResponse struct {
		URL  string `json:"url" example:"https://cdn.example.com/images/original/abc123.jpg"` // Location of the image
		Type string `json:"type" example:"original" enums:"original,thumbnail"`               // Image variant
	}

	// ArticleListResponse is the JSON envelope for a list of articles.
	ArticleListResponse struct {
		Articles []ArticleResponse `json:"articles"` // Articles in the result, in order
	}
)

// NewArticleListResponse converts domain articles into an ArticleListResponse, keeping their order.
// The list is empty, never nil, when there are no articles.
func NewArticleListResponse(articles []rss.Article) *ArticleListResponse {
	output := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		output[i] = NewArticleResponseFromDomain(article)
	}

	return &ArticleListResponse{
		Articles: output,
	}
}

// NewImageResponseFromDomain converts a domain image into an ImageResponse, exposing the
// image path as its URL.
func NewImageResponseFromDomain(image rss.Image) ImageResponse {
	return ImageResponse{
		URL:  image.Path,
		Type: string(image.Type),
	}
}

// Render implements the go-chi render.Renderer interface. It performs no post-processing.
func (a ArticleListResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// NewArticleResponseFromDomain converts a domain article into an ArticleResponse, including
// its feed and images.
func NewArticleResponseFromDomain(article rss.Article) ArticleResponse {
	images := make([]ImageResponse, len(article.Images))
	for i, img := range article.Images {
		images[i] = NewImageResponseFromDomain(img)
	}

	return ArticleResponse{
		Title:       article.Title,
		URL:         article.URL,
		Feed:        NewFeedResponse(article.Feed),
		Images:      images,
		PublishedAt: article.PublishedAt,
	}
}
