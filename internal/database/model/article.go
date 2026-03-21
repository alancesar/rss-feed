package model

import (
	"rss-summary/pkg/rss"
	"time"
)

type (
	Feed struct {
		ID        string `gorm:"primarykey"`
		Name      string
		URL       string
		Articles  []Article
		CreatedAt time.Time
	}

	Article struct {
		ID          string `gorm:"primarykey"`
		FeedID      string
		Source      string
		Title       string
		Image       string
		URL         string `gorm:"uniqueIndex"`
		CreatedAt   time.Time
		PublishedAt time.Time `gorm:"index"`
	}
)

func NewFeedFromDomain(feed rss.Feed) Feed {
	articles := make([]Article, len(feed.Articles))
	for i, article := range feed.Articles {
		articles[i] = NewArticleFromDomain(article, feed)
	}

	return Feed{
		ID:       feed.ID,
		Name:     feed.Name,
		URL:      feed.URL,
		Articles: articles,
	}
}

func NewArticleFromDomain(article rss.Article, feed rss.Feed) Article {
	return Article{
		ID:          article.ID,
		FeedID:      feed.ID,
		Source:      feed.Name,
		Title:       article.Title,
		URL:         article.URL,
		Image:       article.Image,
		PublishedAt: time.Time(article.PublishedAt),
	}
}

func (s Feed) ToDomain() rss.Feed {
	return rss.Feed{
		ID:   s.ID,
		Name: s.Name,
		URL:  s.URL,
	}
}

func (a Article) ToDomain() rss.Article {
	return rss.Article{
		ID:          a.ID,
		Source:      a.Source,
		Title:       a.Title,
		URL:         a.URL,
		Image:       a.Image,
		PublishedAt: rss.DateTime(a.PublishedAt),
	}
}
