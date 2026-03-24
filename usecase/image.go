package usecase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"rss-feed/pkg/event"
	"rss-feed/pkg/rss"

	"github.com/rs/zerolog"
)

type (
	FileStorage interface {
		Create(ctx context.Context, path string, img io.Reader) error
	}

	ImageStore interface {
		SaveImage(ctx context.Context, articleID string, Image rss.Image) error
	}

	ConsumeImage struct {
		client  *http.Client
		storage FileStorage
		db      ImageStore
	}
)

func NewConsumeImage(client *http.Client, storage FileStorage, db ImageStore) *ConsumeImage {
	return &ConsumeImage{
		client:  client,
		storage: storage,
		db:      db,
	}
}

func (uc ConsumeImage) Execute(ctx context.Context, e event.Image) error {
	log := zerolog.Ctx(ctx)

	log.Info().Str("url", e.URL).Str("article_id", e.ArticleID).Msg("downloading image")
	resp, err := uc.client.Get(e.URL)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	u, err := url.Parse(e.URL)
	if err != nil {
		return err
	}
	ext := filepath.Ext(filepath.Base(u.Path))
	path := fmt.Sprintf("images/original/%s%s", e.ArticleID, ext)

	log.Info().Str("path", path).Msg("uploading image to s3")
	if err := uc.storage.Create(ctx, path, resp.Body); err != nil {
		return err
	}

	img := rss.NewImage(path, rss.OriginalImageType)
	log.Info().Str("article_id", e.ArticleID).Str("path", path).Msg("saving image to database")
	if err := uc.db.SaveImage(ctx, e.ArticleID, img); err != nil {
		return err
	}

	return nil
}
