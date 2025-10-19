package rss

import (
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
		ID       uint
		Name     string
		URL      string
		Articles []Article
	}

	Article struct {
		ID          uint
		Source      string
		Title       string
		URL         string
		PublishedAt DateTime
	}
)

func NewArticle(source, title, url, publishedAt string) (Article, error) {
	dateTime, err := NewDateTime(publishedAt)
	if err != nil {
		return Article{}, fmt.Errorf("could not parse published date: %v", err)
	}

	return Article{
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
