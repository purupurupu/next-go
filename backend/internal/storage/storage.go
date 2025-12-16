package storage

import (
	"context"
	"io"
	"time"
)

// UploadResult contains information about an uploaded file
type UploadResult struct {
	Path        string
	ContentType string
	Size        int64
}

// Storage defines the interface for file storage operations
type Storage interface {
	// Upload uploads a file to storage
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*UploadResult, error)

	// Download returns a reader for the file at the given key
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the file at the given key
	Delete(ctx context.Context, key string) error

	// GetURL returns a pre-signed URL for the file
	GetURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// Exists checks if a file exists at the given key
	Exists(ctx context.Context, key string) (bool, error)
}
