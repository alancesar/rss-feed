package usecase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"rss-summary/pkg/rss"
)

type (
	FileStorage interface {
		Create(ctx context.Context, path string, img io.Reader) error
	}

	ImageDatabase interface {
		SaveImage(ctx context.Context, Image rss.Image) error
	}

	AddImageRequest struct {
		ArticleID string `json:"article_id"`
		SourceURL string `json:"source_url"`
	}

	AddImages struct {
		client  *http.Client
		storage FileStorage
		db      ImageDatabase
	}
)

func NewAddImages(client *http.Client, storage FileStorage, db ImageDatabase) *AddImages {
	return &AddImages{
		client:  client,
		storage: storage,
		db:      db,
	}
}

func (uc AddImages) Execute(ctx context.Context, req AddImageRequest) error {
	fmt.Println("Adding image for article", req.ArticleID, "from source", req.SourceURL)
	resp, err := uc.client.Get(req.SourceURL)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	u, err := url.Parse(req.SourceURL)
	if err != nil {
		return err
	}
	ext := filepath.Ext(filepath.Base(u.Path))
	path := fmt.Sprintf("images/original/%s%s", req.ArticleID, ext)
	if err := uc.storage.Create(ctx, path, resp.Body); err != nil {
		return err
	}

	img := rss.NewImage(req.ArticleID, path, rss.OriginalImageType)
	if err := uc.db.SaveImage(ctx, img); err != nil {
		return err
	}

	return nil
}
