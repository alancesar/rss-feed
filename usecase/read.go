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

	ReadUseCase struct {
		store ArticlesStore
	}
)

func NewRead(store ArticlesStore) *ReadUseCase {
	return &ReadUseCase{
		store: store,
	}
}

func (uc ReadUseCase) Execute(ctx context.Context, date time.Time) ([]rss.Article, error) {
	articles, err := uc.store.GetArticlesFromDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("while getting articles from database: %w", err)
	}

	return articles, nil
}
