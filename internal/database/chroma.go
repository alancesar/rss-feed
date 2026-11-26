package database

import (
	"context"
	"rss-feed/internal/ml"
	"rss-feed/pkg/rss"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
)

const (
	// maxDistance is the largest vector distance a search hit may have to still be
	// considered relevant; hits farther away than this are discarded.
	maxDistance = 0.8
)

type (
	// EmbeddingProvider turns text into an embedding vector.
	EmbeddingProvider interface {
		// GetEmbeddings returns the embedding of the given text.
		GetEmbeddings(ctx context.Context, text string) (ml.Embedding, error)
	}

	// ArticleParser renders an article as the plain text that gets embedded.
	ArticleParser interface {
		// Parse returns the text representation of the article.
		Parse(article rss.Article) (string, error)
	}

	// Chroma is a Chroma vector store holding article embeddings in a single collection.
	// Embeddings are computed through an EmbeddingProvider.
	Chroma struct {
		collection chroma.Collection
		embedding  EmbeddingProvider
		parser     ArticleParser
	}
)

// NewChroma creates a Chroma store that reads and writes the given collection, embedding
// text with provider and rendering articles into text with parser.
func NewChroma(collection chroma.Collection, provider EmbeddingProvider, parser ArticleParser) *Chroma {
	return &Chroma{
		collection: collection,
		embedding:  provider,
		parser:     parser,
	}
}

// CreateArticle renders the article as text, embeds it and adds it to the collection using
// the article ID as document ID. The title, description (as "summary"), publication time
// (Unix seconds), feed name, URL and categories are stored as metadata.
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

// FindArticlesIDsByTerm embeds the term and queries the collection for its 5 nearest
// articles, returning their IDs from closest to farthest. Hits beyond maxDistance are dropped.
// It returns no IDs when the query yields no results.
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
	hasDistances := len(distanceGroups) > 0 && len(distanceGroups[0]) == len(idGroups[0])
	ids := make([]string, 0, len(idGroups[0]))
	for i, id := range idGroups[0] {
		if hasDistances && float64(distanceGroups[0][i]) > maxDistance {
			break
		}
		ids = append(ids, string(id))
	}

	return ids, nil
}
