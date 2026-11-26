package usecase

import (
	"context"
	"fmt"
	"rss-feed/pkg/rss"
	"time"
)

type (
	// ArticlesStore queries stored articles.
	ArticlesStore interface {
		// GetArticlesFromDate returns the articles published on the given date.
		GetArticlesFromDate(ctx context.Context, time time.Time) ([]rss.Article, error)
	}

	// ImageStorage grants access to stored images.
	ImageStorage interface {
		// Presign returns a temporary URL that gives access to the file at path for the given duration.
		Presign(ctx context.Context, path string, ttl time.Duration) (string, error)
	}

	// ReadArticles lists the articles of a given date.
	ReadArticles struct {
		store   ArticlesStore
		storage ImageStorage
	}
)

// NewReadArticles creates a ReadArticles use case that reads from store and signs image links through storage.
func NewReadArticles(store ArticlesStore, storage ImageStorage) *ReadArticles {
	return &ReadArticles{
		store:   store,
		storage: storage,
	}
}

// Execute returns the articles published on date. Each image path is replaced by a
// temporary access URL valid for one hour. Any failure aborts the whole operation.
func (uc ReadArticles) Execute(ctx context.Context, date time.Time) ([]rss.Article, error) {
	articles, err := uc.store.GetArticlesFromDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("while getting articles from database: %w", err)
	}

	for i, article := range articles {
		for j, image := range article.Images {
			signedURL, err := uc.storage.Presign(ctx, image.Path, time.Hour)
			if err != nil {
				return nil, err
			}

			articles[i].Images[j].Path = signedURL
		}
	}

	return articles, nil
}
