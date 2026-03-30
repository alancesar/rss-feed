package rss

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

var (
	ErrFeedAlreadyExists = errors.New("feed already exists")
)

const (
	OriginalImageType  ImageType = "original"
	ThumbnailImageType ImageType = "thumbnail"
)

type (
	ImageType string

	Feed struct {
		ID        string
		Name      string
		URL       string
		Articles  []Article
		UpdatedAt time.Time
	}

	Image struct {
		ID   string
		Path string
		Type ImageType
	}

	Article struct {
		ID          string
		Title       string
		Description string
		Content     string
		Categories  []string
		URL         string
		Feed        Feed
		Images      []Image
		PublishedAt time.Time
	}
)

func (f *Feed) Touch() {
	f.UpdatedAt = time.Now().UTC()
}

func NewImage(path string, imgType ImageType) Image {
	return Image{
		ID:   hashFromString(path),
		Path: path,
		Type: imgType,
	}
}

func (a Article) ToMarkdown() string {
	description, err := htmltomarkdown.ConvertString(a.Description)
	if err != nil {
		description = a.Description
	}

	content, err := htmltomarkdown.ConvertString(a.Content)
	if err != nil {
		content = a.Content
	}

	body := description
	if content != "" {
		body = body + "\n\n---\n\n" + content
	}

	return fmt.Sprintf("# %s\n\n%s", a.Title, body)
}

func hashFromString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
