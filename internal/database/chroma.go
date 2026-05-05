package database

import (
	"context"
	"rss-feed/internal/ml"
	"rss-feed/pkg/rss"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
)

const (
	maxDistance = 0.8
)

type (
	EmbeddingProvider interface {
		GetEmbeddings(ctx context.Context, text string) (ml.Embedding, error)
	}

	ArticleParser interface {
		Parse(article rss.Article) (string, error)
	}

	Chroma struct {
		collection chroma.Collection
		embedding  EmbeddingProvider
		parser     ArticleParser
	}
)

func NewChroma(collection chroma.Collection, provider EmbeddingProvider, parser ArticleParser) *Chroma {
	return &Chroma{
		collection: collection,
		embedding:  provider,
		parser:     parser,
	}
}

func (c *Chroma) CreateArticle(ctx context.Context, article rss.Article) error {
	document, err := c.parser.Parse(article)
	if err != nil {
		return err
	}

	embedding, err := c.embedding.GetEmbeddings(ctx, document)
	if err != nil {
		return err
	}

	return c.collection.Add(ctx,
		chroma.WithEmbeddings(embeddings.NewEmbeddingFromFloat32(embedding.Vector)),
		chroma.WithIDs(chroma.DocumentID(article.ID)),
		chroma.WithTexts(document),
		chroma.WithMetadatas(chroma.NewDocumentMetadata(
			chroma.NewStringAttribute("title", article.Title),
			chroma.NewStringAttribute("summary", article.Description),
			chroma.NewIntAttribute("published_at", article.PublishedAt.Unix()),
			chroma.NewStringAttribute("feed", article.Feed.Name),
			chroma.NewStringAttribute("url", article.URL),
			chroma.NewStringArrayAttribute("categories", article.Categories),
		)),
	)
}

func (c *Chroma) FindArticlesIDsByTerm(ctx context.Context, term string) ([]string, error) {
	embedding, err := c.embedding.GetEmbeddings(ctx, term)
	if err != nil {
		return nil, err
	}

	results, err := c.collection.Query(ctx,
		chroma.WithQueryEmbeddings(embeddings.NewEmbeddingFromFloat32(embedding.Vector)),
		chroma.WithNResults(5),
	)
	if err != nil {
		return nil, err
	}

	idGroups := results.GetIDGroups()
	distanceGroups := results.GetDistancesGroups()
	if len(idGroups) == 0 {
		return nil, nil
	}

	// Chroma returns results in ascending distance order, so we can stop at the first exceeded threshold.
	hasDistances := len(distanceGroups) > 0
	ids := make([]string, 0, len(idGroups[0]))
	for i, id := range idGroups[0] {
		if hasDistances && float64(distanceGroups[0][i]) > maxDistance {
			break
		}
		ids = append(ids, string(id))
	}

	return ids, nil
}
