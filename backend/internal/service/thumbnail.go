package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"

	"todo-api/internal/storage"
)

// ThumbnailSize represents a thumbnail size configuration
type ThumbnailSize struct {
	Name   string
	Width  int
	Height int
}

// Predefined thumbnail sizes
var (
	ThumbSize  = ThumbnailSize{Name: "thumb", Width: 300, Height: 300}
	MediumSize = ThumbnailSize{Name: "medium", Width: 800, Height: 800}
)

// ThumbnailResult contains the paths to generated thumbnails
type ThumbnailResult struct {
	ThumbPath  string
	MediumPath string
}

// ThumbnailService handles thumbnail generation
type ThumbnailService struct {
	storage storage.Storage
}

// NewThumbnailService creates a new thumbnail service
func NewThumbnailService(storage storage.Storage) *ThumbnailService {
	return &ThumbnailService{storage: storage}
}

// GenerateThumbnails generates thumbnails for an image
func (s *ThumbnailService) GenerateThumbnails(ctx context.Context, reader io.Reader, basePath string, contentType string) (*ThumbnailResult, error) {
	// Decode the image
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	result := &ThumbnailResult{}

	// Generate thumb (300x300)
	thumbPath, err := s.generateAndUpload(ctx, img, basePath, ThumbSize, format)
	if err != nil {
		return nil, fmt.Errorf("failed to generate thumb: %w", err)
	}
	result.ThumbPath = thumbPath

	// Generate medium (800x800)
	mediumPath, err := s.generateAndUpload(ctx, img, basePath, MediumSize, format)
	if err != nil {
		return nil, fmt.Errorf("failed to generate medium: %w", err)
	}
	result.MediumPath = mediumPath

	return result, nil
}

func (s *ThumbnailService) generateAndUpload(ctx context.Context, img image.Image, basePath string, size ThumbnailSize, format string) (string, error) {
	// Resize the image (fit within bounds, maintaining aspect ratio)
	resized := imaging.Fit(img, size.Width, size.Height, imaging.Lanczos)

	// Encode to buffer
	var buf bytes.Buffer
	var contentType string

	switch format {
	case "jpeg":
		err := imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(85))
		if err != nil {
			return "", err
		}
		contentType = "image/jpeg"
	case "png":
		err := imaging.Encode(&buf, resized, imaging.PNG)
		if err != nil {
			return "", err
		}
		contentType = "image/png"
	case "gif":
		err := imaging.Encode(&buf, resized, imaging.GIF)
		if err != nil {
			return "", err
		}
		contentType = "image/gif"
	default:
		// Default to JPEG for other formats (like webp)
		err := imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(85))
		if err != nil {
			return "", err
		}
		contentType = "image/jpeg"
	}

	// Generate path: basePath_thumb.jpg
	ext := getExtensionForFormat(format)
	baseWithoutExt := strings.TrimSuffix(basePath, filepath.Ext(basePath))
	path := fmt.Sprintf("%s_%s%s", baseWithoutExt, size.Name, ext)

	// Upload
	_, err := s.storage.Upload(ctx, path, &buf, int64(buf.Len()), contentType)
	if err != nil {
		return "", err
	}

	return path, nil
}

func getExtensionForFormat(format string) string {
	switch format {
	case "jpeg":
		return ".jpg"
	case "png":
		return ".png"
	case "gif":
		return ".gif"
	default:
		return ".jpg"
	}
}

// IsImageContentType checks if the content type is an image
func IsImageContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}
