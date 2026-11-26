package usecase

import (
	"context"
	"rss-feed/pkg/rss"
)

type (
	// ArticleVectorSearch looks articles up by meaning rather than exact text.
	ArticleVectorSearch interface {
		// FindArticlesIDsByTerm returns the IDs of the articles most relevant to the term, best match first.
		FindArticlesIDsByTerm(ctx context.Context, term string) ([]string, error)
	}

	// ArticleRepository loads stored articles.
	ArticleRepository interface {
		// GetArticleByID returns the article and whether it was found.
		GetArticleByID(ctx context.Context, articleID string) (rss.Article, bool, error)
	}

	// Find performs a semantic search over the articles.
	Find struct {
		vectorSearch ArticleVectorSearch
		repository   ArticleRepository
	}
)

// NewFind creates a Find use case that ranks through vectorSearch and loads articles from repository.
func NewFind(vectorSearch ArticleVectorSearch, repository ArticleRepository) *Find {
	return &Find{
		vectorSearch: vectorSearch,
		repository:   repository,
	}
}

// Execute returns the articles most relevant to term, preserving the relevance order.
// IDs returned by the search that no longer exist in the repository are skipped.
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
