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
)

type (
	FileStorage interface {
		Create(ctx context.Context, path string, img io.Reader) error
	}

	ImageStore interface {
		SaveImage(ctx context.Context, Image rss.Image) error
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
	resp, err := uc.client.Get(e.URL)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	u, err := url.Parse(e.URL)
	if err != nil {
		return err
	}
	ext := filepath.Ext(filepath.Base(u.Path))
	path := fmt.Sprintf("images/original/%s%s", e.ArticleID, ext)
	if err := uc.storage.Create(ctx, path, resp.Body); err != nil {
		return err
	}

	img := rss.NewImage(e.ArticleID, path, rss.OriginalImageType)
	if err := uc.db.SaveImage(ctx, img); err != nil {
		return err
	}

	return nil
}
