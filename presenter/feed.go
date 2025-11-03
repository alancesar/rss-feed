package presenter

import "rss-summary/pkg/rss"

type (
	AddFeedRequest struct {
		URL string `json:"url"`
	}

	AddFeedResponse struct {
		Name     string            `json:"name"`
		URL      string            `json:"url"`
		Articles []ArticleResponse `json:"articles"`
	}
)

func NewAddFeedResponseFromDomain(feed rss.Feed) AddFeedResponse {
	articles := make([]ArticleResponse, len(feed.Articles))
	for i, article := range feed.Articles {
		articles[i] = NewArticleResponseFromDomain(article)
	}

	return AddFeedResponse{
		Name:     feed.Name,
		URL:      feed.URL,
		Articles: articles,
	}
}
