// Package storage provides object storage adapters.
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type (
	// S3 stores objects in a single bucket of an S3-compatible service. Uploads go through the
	// AWS SDK transfer manager and downloads are granted through presigned URLs.
	S3 struct {
		client *s3.PresignClient
		tm     *transfermanager.Client
		bucket string
	}
)

// NewS3 creates an S3 store for the bucket at the given endpoint and region. It loads the
// default AWS configuration (credentials come from the standard AWS sources) and uses
// path-style addressing, which S3-compatible services commonly require.
func NewS3(endpoint, region, bucket string) (*S3, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &S3{
		client: s3.NewPresignClient(client),
		tm:     transfermanager.New(client),
		bucket: bucket,
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

// Presign returns a pre-signed URL for the object at the given path, valid for the given TTL.
func (s *S3) Presign(ctx context.Context, path string, ttl time.Duration) (string, error) {
	req, err := s.client.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("failed to presign object: %w", err)
	}

	return req.URL, nil
}
