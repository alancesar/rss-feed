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
		ID   string
		Path string
		Type ImageType
	}

	Article struct {
		ID          string
		Title       string
		URL         string
		Feed        Feed
		Images      []Image
		PublishedAt time.Time
	}
)

func NewImage(path string, imgType ImageType) Image {
	return Image{
		ID:   hashFromString(path),
		Path: path,
		Type: imgType,
	}
}

func (a Article) ToMarkdown() string {
	return fmt.Sprintf(`[%s](%s)`, html.UnescapeString(a.Title), a.URL)
}

func hashFromString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
