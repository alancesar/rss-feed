package rss

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"regexp"
	"time"
)

var (
	rfc3339regex  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$`)
	dateOnlyRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

type (
	DateTime time.Time

	Feed struct {
		ID       string
		Name     string
		URL      string
		Articles []Article
	}

	Article struct {
		ID          string
		Source      string
		Title       string
		URL         string
		PublishedAt DateTime
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

func NewArticle(source, title, url, publishedAt string) (Article, error) {
	dateTime, err := NewDateTime(publishedAt)
	if err != nil {
		return Article{}, fmt.Errorf("could not parse published date: %v", err)
	}

	return Article{
		ID:          hashFromString(url),
		Source:      source,
		Title:       title,
		URL:         url,
		PublishedAt: dateTime,
	}, nil
}

func (a Article) ToMarkdown() string {
	return fmt.Sprintf(`[%s](%s)`, html.UnescapeString(a.Title), a.URL)
}

func NewDateTime(raw string) (DateTime, error) {
	var (
		dt  time.Time
		err error
	)

	switch {
	case rfc3339regex.MatchString(raw):
		dt, err = time.Parse(time.RFC3339, raw)
	case dateOnlyRegex.MatchString(raw):
		dt, err = time.Parse(time.DateOnly, raw)
	default:
		dt, err = time.Parse(time.RFC1123Z, raw)
	}

	if err != nil {
		return DateTime{}, fmt.Errorf("could not parse date: %v", err)
	}

	return DateTime(dt), nil
}

func hashFromString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
