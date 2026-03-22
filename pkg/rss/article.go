package rss

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"time"
)

const (
	OriginalImageType  ImageType = "original"
	ThumbnailImageType ImageType = "thumbnail"
)

type (
	ImageType string

	Feed struct {
		ID       string
		Name     string
		URL      string
		Articles []Article
	}

	Image struct {
		ID        string
		ArticleID string
		Path      string
		Type      ImageType
	}

	Article struct {
		ID          string
		Source      string
		Title       string
		URL         string
		Images      []Image
		PublishedAt time.Time
	}
)

func NewFeed(name, url string, articles []Article) Feed {
	return Feed{
		ID:       hashFromString(url),
		Name:     name,
		URL:      url,
		Articles: articles,
	}
}

func NewImage(articleID, path string, imgType ImageType) Image {
	return Image{
		ID:        hashFromString(articleID + string(imgType)),
		ArticleID: articleID,
		Path:      path,
		Type:      imgType,
	}
}

func NewArticle(source, title, url string, publishedAt *time.Time) (Article, error) {
	article := Article{
		ID:     hashFromString(url),
		Source: source,
		Title:  title,
		URL:    url,
	}

	if publishedAt != nil {
		article.PublishedAt = *publishedAt
	}

	return article, nil
}

func (a Article) ToMarkdown() string {
	return fmt.Sprintf(`[%s](%s)`, html.UnescapeString(a.Title), a.URL)
}

func hashFromString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
