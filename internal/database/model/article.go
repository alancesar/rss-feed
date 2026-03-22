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
		URL         string `gorm:"uniqueIndex"`
		Images      []Image
		CreatedAt   time.Time
		PublishedAt time.Time `gorm:"index"`
	}

	Image struct {
		ID        string `gorm:"primarykey"`
		ArticleID string
		Path      string
		Type      string
		CreatedAt time.Time
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
		PublishedAt: time.Time(article.PublishedAt),
	}
}

func NewImageFromDomain(image rss.Image) Image {
	return Image{
		ID:        image.ID,
		ArticleID: image.ArticleID,
		Path:      image.Path,
		Type:      string(image.Type),
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
	images := make([]rss.Image, len(a.Images))
	for i, img := range a.Images {
		images[i] = img.ToDomain()
	}

	return rss.Article{
		ID:          a.ID,
		Source:      a.Source,
		Title:       a.Title,
		URL:         a.URL,
		Images:      images,
		PublishedAt: rss.DateTime(a.PublishedAt),
	}
}

func (i Image) ToDomain() rss.Image {
	return rss.Image{
		ID:        i.ID,
		ArticleID: i.ArticleID,
		Path:      i.Path,
		Type:      rss.ImageType(i.Type),
	}
}
