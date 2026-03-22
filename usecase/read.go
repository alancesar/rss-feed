package usecase

import (
	"context"
	"fmt"
	"rss-summary/pkg/rss"
	"time"
)

type (
	ArticlesStore interface {
		GetArticlesFromDate(ctx context.Context, time time.Time) ([]rss.Article, error)
	}

	ImageStorage interface {
		Presign(ctx context.Context, path string, ttl time.Duration) (string, error)
	}

	ReadArticle struct {
		store   ArticlesStore
		storage ImageStorage
	}
)

func NewRead(store ArticlesStore, storage ImageStorage) *ReadArticle {
	return &ReadArticle{
		store:   store,
		storage: storage,
	}
}

func (uc ReadArticle) Execute(ctx context.Context, date time.Time) ([]rss.Article, error) {
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
