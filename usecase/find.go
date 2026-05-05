package usecase

import (
	"context"
	"rss-feed/pkg/rss"
)

type (
	ArticleVectorSearch interface {
		FindArticlesIDsByTerm(ctx context.Context, term string) ([]string, error)
	}

	ArticleRepository interface {
		GetArticleByID(ctx context.Context, articleID string) (rss.Article, bool, error)
	}

	Find struct {
		vectorSearch ArticleVectorSearch
		repository   ArticleRepository
	}
)

func NewFind(vectorSearch ArticleVectorSearch, repository ArticleRepository) *Find {
	return &Find{
		vectorSearch: vectorSearch,
		repository:   repository,
	}
}

func (uc Find) Execute(ctx context.Context, term string) ([]rss.Article, error) {
	ids, err := uc.vectorSearch.FindArticlesIDsByTerm(ctx, term)
	if err != nil {
		return nil, err
	}

	articles := make([]rss.Article, 0, len(ids))
	for _, id := range ids {
		article, ok, err := uc.repository.GetArticleByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if ok {
			articles = append(articles, article)
		}
	}

	return articles, nil
}
