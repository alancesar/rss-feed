package usecase

import (
	"context"
	"encoding/json"
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
	// FileStorage stores files.
	FileStorage interface {
		// Create writes the content of img at the given path.
		Create(ctx context.Context, path string, img io.Reader) error
	}

	// ImageStore records images belonging to articles.
	ImageStore interface {
		// SaveImage associates the image with the article.
		SaveImage(ctx context.Context, articleID string, image rss.Image) error
	}

	// ConsumeImage is a consumer that downloads the image of every discovered article,
	// stores the file and records it for the article.
	ConsumeImage struct {
		client     *http.Client
		subscriber Subscriber
		storage    FileStorage
		db         ImageStore
	}
)

// NewConsumeImage creates a ConsumeImage use case that downloads with client, listens through
// subscriber, stores files in storage and records them in db.
func NewConsumeImage(client *http.Client, subscriber Subscriber, storage FileStorage, db ImageStore) *ConsumeImage {
	return &ConsumeImage{
		client:     client,
		storage:    storage,
		subscriber: subscriber,
		db:         db,
	}
}

// Execute consumes "article found" events. Articles without an image are acknowledged and
// skipped; download, storage or persistence failures are rejected with a request to redeliver,
// and messages that cannot be decoded are rejected with a request not to redeliver. Whether the
// broker honours that request is up to the broker. It runs until the subscription is
// closed and closes the subscriber on exit.
func (uc ConsumeImage) Execute(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("starting consume image")

	defer func() {
		_ = uc.subscriber.Close()
		logger.Info().Msg("finished consume image")
	}()

	deliveries, err := uc.subscriber.Subscribe(ctx, event.TopicFeedArticleFound)
	if err != nil {
		return err
	}

	for delivery := range deliveries {
		var e event.Article
		if err := json.Unmarshal(delivery.Payload, &e); err != nil {
			logger.Error().Err(err).Msg("failed to unmarshal article")
			delivery.Nack(false)
			continue
		}

		img := e.Image
		if img.ImageID == "" {
			logger.Info().Str("article_id", e.ArticleID).Msg("image id is empty")
			delivery.Ack()
			continue
		}

		logger.Info().Str("url", img.URL).Str("article_id", e.ArticleID).Msg("downloading image")
		if err := uc.saveImageToStorage(ctx, e.ArticleID, img.URL); err != nil {
			delivery.Nack(true)
			continue
		}

		delivery.Ack()
	}

	return nil
}

// saveImageToStorage downloads the image at imgURL, stores it under
// images/original/{articleID}{extension} (the extension comes from the URL) and then
// records it as the article's original image. A non-200 response is an error.
func (uc ConsumeImage) saveImageToStorage(ctx context.Context, articleID, imgURL string) error {
	resp, err := uc.client.Get(imgURL)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d fetching image", resp.StatusCode)
	}

	u, err := url.Parse(imgURL)
	if err != nil {
		return err
	}
	ext := filepath.Ext(filepath.Base(u.Path))
	path := fmt.Sprintf("images/original/%s%s", articleID, ext)

	logger := zerolog.Ctx(ctx)
	logger.Info().Str("path", path).Msg("uploading image to s3")
	if err := uc.storage.Create(ctx, path, resp.Body); err != nil {
		return err
	}

	img := rss.NewImage(path, rss.OriginalImageType)
	logger.Info().Str("article_id", articleID).Str("path", path).Msg("saving image to database")
	if err := uc.db.SaveImage(ctx, articleID, img); err != nil {
		return err
	}

	return nil
}
