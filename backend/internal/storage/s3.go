package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"

	appconfig "todo-api/internal/config"
)

// S3Storage implements Storage interface for S3-compatible storage
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage creates a new S3 storage client
func NewS3Storage(cfg *appconfig.S3Config) (*S3Storage, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				HostnameImmutable: true,
			}, nil
		},
	)

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	storage := &S3Storage{
		client: client,
		bucket: cfg.Bucket,
	}

	// Ensure bucket exists
	if err := storage.ensureBucket(context.Background()); err != nil {
		log.Warn().Err(err).Str("bucket", cfg.Bucket).Msg("Failed to ensure bucket exists")
	}

	return storage, nil
}

// ensureBucket creates the bucket if it doesn't exist
func (s *S3Storage) ensureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil // Bucket exists
	}

	// Try to create bucket
	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	log.Info().Str("bucket", s.bucket).Msg("Created S3 bucket")
	return nil
}

// Upload uploads a file to storage
func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*UploadResult, error) {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	return &UploadResult{
		Path:        key,
		ContentType: contentType,
		Size:        size,
	}, nil
}

// Download returns a reader for the file at the given key
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	output, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return output.Body, nil
}

// Delete removes the file at the given key
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// GetURL returns a pre-signed URL for the file
func (s *S3Storage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	resp, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return resp.URL, nil
}

// Exists checks if a file exists at the given key
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}

	return true, nil
}
