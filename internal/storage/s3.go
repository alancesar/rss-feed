package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var ErrNotFound = errors.New("object not found")

type S3 struct {
	client *s3.Client
	tm     *transfermanager.Client
	bucket string
}

func NewS3(ctx context.Context) (*S3, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	endpoint := os.Getenv("AWS_ENDPOINT")
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	return &S3{
		client: client,
		tm:     transfermanager.New(client),
		bucket: os.Getenv("AWS_BUCKET"),
	}, nil
}

// Create uploads the content from body to S3 at the given path.
func (s *S3) Create(ctx context.Context, path string, body io.Reader) error {
	_, err := s.tm.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   body,
	})
	return err
}

// Get returns an io.Reader for the object at the given path.
// Returns ErrNotFound if no object exists at that path.
// The caller is responsible for closing the reader.
func (s *S3) Get(ctx context.Context, path string) (io.Reader, error) {
	output, err := s.tm.GetObject(ctx, &transfermanager.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return output.Body, nil
}

// Presign returns a pre-signed URL for the object at the given path, valid for the given TTL.
func (s *S3) Presign(ctx context.Context, path string, ttl time.Duration) (string, error) {
	req, err := s3.NewPresignClient(s.client).PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("failed to presign object: %w", err)
	}
	return req.URL, nil
}
