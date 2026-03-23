package model

import (
	"rss-feed/pkg/rss"
	"time"
)

type (
	Feed struct {
		ID        string `gorm:"primarykey"`
		Name      string
		URL       string
		Articles  []Article `gorm:"foreignKey:FeedID;references:ID"`
		CreatedAt time.Time
	}

	Article struct {
		ID          string `gorm:"primarykey"`
		FeedID      string `gorm:"index"`
		Feed        Feed   `gorm:"foreignKey:FeedID;references:ID"`
		Title       string
		URL         string  `gorm:"uniqueIndex"`
		Images      []Image `gorm:"foreignKey:ArticleID;references:ID"`
		CreatedAt   time.Time
		PublishedAt time.Time `gorm:"index"`
	}

	Image struct {
		ID        string `gorm:"primarykey"`
		ArticleID string `gorm:"index"`
		Path      string
		Type      string
		CreatedAt time.Time
	}
)

func NewFeedFromDomain(feed rss.Feed) Feed {
	articles := make([]Article, len(feed.Articles))
	for i, article := range feed.Articles {
		articles[i] = NewArticleFromDomain(article, feed.ID)
	}

	return Feed{
		ID:       feed.ID,
		Name:     feed.Name,
		URL:      feed.URL,
		Articles: articles,
	}
}

func NewArticleFromDomain(article rss.Article, feedID string) Article {
	return Article{
		ID:          article.ID,
		FeedID:      feedID,
		Title:       article.Title,
		URL:         article.URL,
		PublishedAt: article.PublishedAt,
	}
}

func NewImageFromDomain(image rss.Image, articleID string) Image {
	return Image{
		ID:        image.ID,
		ArticleID: articleID,
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
		Title:       a.Title,
		URL:         a.URL,
		Images:      images,
		Feed:        a.Feed.ToDomain(),
		PublishedAt: a.PublishedAt,
	}
}

func (i Image) ToDomain() rss.Image {
	return rss.Image{
		ID:   i.ID,
		Path: i.Path,
		Type: rss.ImageType(i.Type),
	}
}
