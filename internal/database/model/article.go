// Package model holds the GORM models that map the domain entities to database tables,
// and the conversions between them.
package model

import (
	"encoding/json"
	"rss-feed/pkg/rss"
	"time"

	"gorm.io/datatypes"
)

type (
	// Feed is the GORM model of an RSS feed. It owns many Articles.
	Feed struct {
		ID        string `gorm:"primarykey"`
		Name      string
		URL       string
		Articles  []Article `gorm:"foreignKey:FeedID;references:ID"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	// Article is the GORM model of an article. It belongs to a Feed, owns many Images, has a
	// unique URL and stores its categories as a JSON array.
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

	// Image is the GORM model of an image file belonging to an article, identified by its
	// storage path and its type.
	Image struct {
		ID        string `gorm:"primarykey"`
		ArticleID string `gorm:"index"`
		Path      string
		Type      string
		CreatedAt time.Time
	}
)

// NewFeedFromDomain converts a domain feed, and its articles, into a Feed model.
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

// NewArticleFromDomain converts a domain article into an Article model owned by feedID.
// Categories are encoded as JSON; images are not carried over.
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

// NewImageFromDomain converts a domain image into an Image model owned by articleID.
func NewImageFromDomain(image rss.Image, articleID string) Image {
	return Image{
		ID:        image.ID,
		ArticleID: articleID,
		Path:      image.Path,
		Type:      string(image.Type),
	}
}

// ToDomain converts the model into a domain feed. Articles are not carried over.
func (s Feed) ToDomain() rss.Feed {
	return rss.Feed{
		ID:        s.ID,
		Name:      s.Name,
		URL:       s.URL,
		UpdatedAt: s.UpdatedAt,
	}
}

// ToDomain converts the model into a domain article, including its feed and images.
// Categories that cannot be decoded from JSON are left empty.
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

// ToDomain converts the model into a domain image.
func (i Image) ToDomain() rss.Image {
	return rss.Image{
		ID:   i.ID,
		Path: i.Path,
		Type: rss.ImageType(i.Type),
	}
}
