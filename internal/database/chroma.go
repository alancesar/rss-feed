package database

import (
	"context"
	"rss-feed/pkg/rss"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
)

type (
	Chroma struct {
		collection chroma.Collection
	}
)

func NewChroma(ctx context.Context, client chroma.Client) (*Chroma, error) {
	collection, err := client.GetOrCreateCollection(ctx, "articles")
	if err != nil {
		return nil, err
	}

	return &Chroma{
		collection: collection,
	}, nil
}

func (c Chroma) CreateArticle(ctx context.Context, article rss.Article) error {
	if err := c.collection.Add(ctx,
		chroma.WithIDs(chroma.DocumentID(article.ID)),
		chroma.WithTexts(article.ToMarkdown()),
		chroma.WithMetadatas(chroma.NewDocumentMetadata(
			chroma.NewStringAttribute("title", article.Title),
			chroma.NewStringAttribute("summary", article.Description),
			chroma.NewIntAttribute("published_at", article.PublishedAt.Unix()),
			chroma.NewStringAttribute("feed", article.Feed.Name),
			chroma.NewStringAttribute("url", article.URL),
			chroma.NewStringArrayAttribute("categories", article.Categories),
		))); err != nil {
		return err
	}

	return nil
}

func (c Chroma) FindArticlesIDsByTerm(ctx context.Context, term string) ([]string, error) {
	results, err := c.collection.Query(ctx,
		chroma.WithQueryTexts(term),
		chroma.WithNResults(5),
	)
	if err != nil {
		return nil, err
	}

	groups := results.GetIDGroups()
	if len(groups) == 0 {
		return nil, nil
	}

	ids := make([]string, 0, len(groups[0]))
	for _, id := range groups[0] {
		ids = append(ids, string(id))
	}

	return ids, nil
}
