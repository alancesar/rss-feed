package model

import (
	"encoding/json"
	"rss-feed/pkg/rss"
	"time"

	"gorm.io/datatypes"
)

type (
	Feed struct {
		ID        string `gorm:"primarykey"`
		Name      string
		URL       string
		Articles  []Article `gorm:"foreignKey:FeedID;references:ID"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	Article struct {
		ID          string `gorm:"primarykey"`
		FeedID      string `gorm:"index"`
		Feed        Feed   `gorm:"foreignKey:FeedID;references:ID"`
		Title       string
		Description string
		Content     string
		Categories  datatypes.JSON
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
		ID:        feed.ID,
		Name:      feed.Name,
		URL:       feed.URL,
		Articles:  articles,
		UpdatedAt: feed.UpdatedAt,
	}
}

func NewArticleFromDomain(article rss.Article, feedID string) Article {
	categories, _ := json.Marshal(article.Categories)

	return Article{
		ID:          article.ID,
		FeedID:      feedID,
		Title:       article.Title,
		Description: article.Description,
		Content:     article.Content,
		Categories:  categories,
		URL:         article.URL,
		CreatedAt:   time.Time{},
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
		ID:        s.ID,
		Name:      s.Name,
		URL:       s.URL,
		UpdatedAt: s.UpdatedAt,
	}
}

func (a Article) ToDomain() rss.Article {
	images := make([]rss.Image, len(a.Images))
	for i, img := range a.Images {
		images[i] = img.ToDomain()
	}

	var categories []string
	_ = json.Unmarshal(a.Categories, &categories)

	return rss.Article{
		ID:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		Content:     a.Content,
		Categories:  categories,
		URL:         a.URL,
		Feed:        a.Feed.ToDomain(),
		Images:      images,
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
