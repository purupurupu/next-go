package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"todo-api/internal/errors"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/internal/storage"
)

// FileService handles file upload business logic
type FileService struct {
	fileRepo *repository.FileRepository
	todoRepo *repository.TodoRepository
	storage  storage.Storage
	thumbSvc *ThumbnailService
}

// NewFileService creates a new FileService
func NewFileService(
	fileRepo *repository.FileRepository,
	todoRepo *repository.TodoRepository,
	storage storage.Storage,
	thumbSvc *ThumbnailService,
) *FileService {
	return &FileService{
		fileRepo: fileRepo,
		todoRepo: todoRepo,
		storage:  storage,
		thumbSvc: thumbSvc,
	}
}

// UploadInput represents input for file upload
type UploadInput struct {
	UserID      int64
	TodoID      int64
	FileName    string
	ContentType string
	FileSize    int64
	FileReader  io.Reader
}

// Upload uploads a file and generates thumbnails if applicable
func (s *FileService) Upload(ctx context.Context, input UploadInput) (*model.File, error) {
	// Validate file size
	if input.FileSize > model.MaxFileSize {
		return nil, errors.ValidationFailed(map[string][]string{
			"file": {fmt.Sprintf("ファイルサイズは%dMB以下にしてください", model.MaxFileSize/1024/1024)},
		})
	}

	// Validate MIME type
	if !model.IsAllowedMimeType(input.ContentType) {
		return nil, errors.ValidationFailed(map[string][]string{
			"file": {"許可されていないファイル形式です"},
		})
	}

	// Validate todo ownership
	if _, err := s.todoRepo.FindByID(input.TodoID, input.UserID); err != nil {
		return nil, errors.NotFound("Todo", input.TodoID)
	}

	// Generate unique storage path
	storagePath := s.generateStoragePath(input.UserID, input.FileName)

	// Read file into buffer for potential reuse (thumbnails)
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, input.FileReader); err != nil {
		return nil, errors.InternalErrorWithLog(err, "FileService.Upload: failed to read file")
	}

	// Upload original file
	_, err := s.storage.Upload(ctx, storagePath, bytes.NewReader(buf.Bytes()), input.FileSize, input.ContentType)
	if err != nil {
		return nil, errors.InternalErrorWithLog(err, "FileService.Upload: failed to upload file")
	}

	// Create file record
	file := &model.File{
		UserID:         input.UserID,
		AttachableType: model.AttachableTypeTodo,
		AttachableID:   input.TodoID,
		OriginalName:   input.FileName,
		StoragePath:    storagePath,
		ContentType:    input.ContentType,
		FileSize:       input.FileSize,
		FileType:       model.GetFileType(input.ContentType),
	}

	// Generate thumbnails for images
	if file.IsImage() {
		thumbResult, err := s.thumbSvc.GenerateThumbnails(ctx, bytes.NewReader(buf.Bytes()), storagePath, input.ContentType)
		if err != nil {
			log.Error().Err(err).Msg("FileService.Upload: failed to generate thumbnails")
			// Continue without thumbnails
		} else {
			file.ThumbPath = &thumbResult.ThumbPath
			file.MediumPath = &thumbResult.MediumPath
		}
	}

	// Save to database
	if err := s.fileRepo.Create(file); err != nil {
		// Cleanup uploaded files on failure
		s.cleanupStoredFiles(ctx, file)
		return nil, errors.InternalErrorWithLog(err, "FileService.Upload: failed to save file record")
	}

	return file, nil
}

// Delete deletes a file
func (s *FileService) Delete(ctx context.Context, fileID, todoID, userID int64) error {
	// Validate todo ownership
	if _, err := s.todoRepo.FindByID(todoID, userID); err != nil {
		return errors.NotFound("Todo", todoID)
	}

	// Get file
	file, err := s.fileRepo.FindByID(fileID)
	if err != nil {
		return errors.NotFound("File", fileID)
	}

	// Verify file belongs to the todo
	if file.AttachableType != model.AttachableTypeTodo || file.AttachableID != todoID {
		return errors.NotFound("File", fileID)
	}

	// Verify ownership
	if !file.IsOwnedBy(userID) {
		return errors.AuthorizationFailed("File", "delete")
	}

	// Delete from storage
	s.cleanupStoredFiles(ctx, file)

	// Delete from database
	if err := s.fileRepo.Delete(fileID); err != nil {
		return errors.InternalErrorWithLog(err, "FileService.Delete: failed to delete file record")
	}

	return nil
}

// Download returns a reader for downloading a file
func (s *FileService) Download(ctx context.Context, fileID, todoID, userID int64) (io.ReadCloser, *model.File, error) {
	// Validate todo ownership
	if _, err := s.todoRepo.FindByID(todoID, userID); err != nil {
		return nil, nil, errors.NotFound("Todo", todoID)
	}

	// Get file
	file, err := s.fileRepo.FindByID(fileID)
	if err != nil {
		return nil, nil, errors.NotFound("File", fileID)
	}

	// Verify file belongs to the todo
	if file.AttachableType != model.AttachableTypeTodo || file.AttachableID != todoID {
		return nil, nil, errors.NotFound("File", fileID)
	}

	// Download from storage
	reader, err := s.storage.Download(ctx, file.StoragePath)
	if err != nil {
		return nil, nil, errors.InternalErrorWithLog(err, "FileService.Download: failed to download file")
	}

	return reader, file, nil
}

// DownloadThumbnail returns a reader for downloading a thumbnail
func (s *FileService) DownloadThumbnail(ctx context.Context, fileID, todoID, userID int64, size string) (io.ReadCloser, *model.File, error) {
	// Validate todo ownership
	if _, err := s.todoRepo.FindByID(todoID, userID); err != nil {
		return nil, nil, errors.NotFound("Todo", todoID)
	}

	// Get file
	file, err := s.fileRepo.FindByID(fileID)
	if err != nil {
		return nil, nil, errors.NotFound("File", fileID)
	}

	// Verify file belongs to the todo
	if file.AttachableType != model.AttachableTypeTodo || file.AttachableID != todoID {
		return nil, nil, errors.NotFound("File", fileID)
	}

	// Verify it's an image
	if !file.IsImage() {
		return nil, nil, errors.ValidationFailed(map[string][]string{
			"file": {"サムネイルは画像ファイルのみ利用可能です"},
		})
	}

	// Get thumbnail path
	var thumbPath *string
	switch size {
	case "thumb":
		thumbPath = file.ThumbPath
	case "medium":
		thumbPath = file.MediumPath
	default:
		return nil, nil, errors.ValidationFailed(map[string][]string{
			"size": {"無効なサムネイルサイズです"},
		})
	}

	if thumbPath == nil {
		return nil, nil, errors.NotFound("Thumbnail", fileID)
	}

	// Download from storage
	reader, err := s.storage.Download(ctx, *thumbPath)
	if err != nil {
		return nil, nil, errors.InternalErrorWithLog(err, "FileService.DownloadThumbnail: failed to download thumbnail")
	}

	return reader, file, nil
}

// ListByTodo returns all files for a todo
func (s *FileService) ListByTodo(ctx context.Context, todoID, userID int64) ([]model.File, error) {
	// Validate todo ownership
	if _, err := s.todoRepo.FindByID(todoID, userID); err != nil {
		return nil, errors.NotFound("Todo", todoID)
	}

	files, err := s.fileRepo.FindByAttachable(model.AttachableTypeTodo, todoID)
	if err != nil {
		return nil, errors.InternalErrorWithLog(err, "FileService.ListByTodo: failed to list files")
	}

	return files, nil
}

func (s *FileService) generateStoragePath(userID int64, fileName string) string {
	ext := filepath.Ext(fileName)
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("2006/01/02")
	return fmt.Sprintf("uploads/%d/%s/%s%s", userID, timestamp, uniqueID, ext)
}

func (s *FileService) cleanupStoredFiles(ctx context.Context, file *model.File) {
	// Delete original
	if err := s.storage.Delete(ctx, file.StoragePath); err != nil {
		log.Error().Err(err).Str("path", file.StoragePath).Msg("Failed to delete original file")
	}

	// Delete thumbnails
	if file.ThumbPath != nil {
		if err := s.storage.Delete(ctx, *file.ThumbPath); err != nil {
			log.Error().Err(err).Str("path", *file.ThumbPath).Msg("Failed to delete thumb")
		}
	}
	if file.MediumPath != nil {
		if err := s.storage.Delete(ctx, *file.MediumPath); err != nil {
			log.Error().Err(err).Str("path", *file.MediumPath).Msg("Failed to delete medium")
		}
	}
}
