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

	ReadArticle struct {
		store ArticlesStore
	}
)

func NewRead(store ArticlesStore) *ReadArticle {
	return &ReadArticle{
		store: store,
	}
}

func (uc ReadArticle) Execute(ctx context.Context, date time.Time) ([]rss.Article, error) {
	articles, err := uc.store.GetArticlesFromDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("while getting articles from database: %w", err)
	}

	return articles, nil
}
