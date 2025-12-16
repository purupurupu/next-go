package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"todo-api/internal/errors"
	"todo-api/internal/model"
	"todo-api/internal/service"
	"todo-api/pkg/response"
	"todo-api/pkg/util"
)

// FileHandler handles file-related endpoints
type FileHandler struct {
	fileService *service.FileService
}

// NewFileHandler creates a new FileHandler
func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// FileResponse represents a file in API responses
type FileResponse struct {
	ID           int64   `json:"id"`
	OriginalName string  `json:"original_name"`
	ContentType  string  `json:"content_type"`
	FileSize     int64   `json:"file_size"`
	FileType     string  `json:"file_type"`
	ThumbURL     *string `json:"thumb_url,omitempty"`
	MediumURL    *string `json:"medium_url,omitempty"`
	DownloadURL  string  `json:"download_url"`
	CreatedAt    string  `json:"created_at"`
}

// toFileResponse converts a model.File to FileResponse
func toFileResponse(file *model.File, todoID int64) FileResponse {
	resp := FileResponse{
		ID:           file.ID,
		OriginalName: file.OriginalName,
		ContentType:  file.ContentType,
		FileSize:     file.FileSize,
		FileType:     string(file.FileType),
		DownloadURL:  fmt.Sprintf("/api/v1/todos/%d/files/%d", todoID, file.ID),
		CreatedAt:    util.FormatRFC3339(file.CreatedAt),
	}

	if file.ThumbPath != nil {
		thumbURL := fmt.Sprintf("/api/v1/todos/%d/files/%d/thumb", todoID, file.ID)
		resp.ThumbURL = &thumbURL
	}
	if file.MediumPath != nil {
		mediumURL := fmt.Sprintf("/api/v1/todos/%d/files/%d/medium", todoID, file.ID)
		resp.MediumURL = &mediumURL
	}

	return resp
}

// List retrieves all files for a todo
// GET /api/v1/todos/:todo_id/files
func (h *FileHandler) List(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	files, err := h.fileService.ListByTodo(c.Request().Context(), todoID, currentUser.ID)
	if err != nil {
		return err
	}

	fileResponses := make([]FileResponse, len(files))
	for i, file := range files {
		fileResponses[i] = toFileResponse(&file, todoID)
	}

	return c.JSON(http.StatusOK, fileResponses)
}

// Upload handles file upload
// POST /api/v1/todos/:todo_id/files
func (h *FileHandler) Upload(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return errors.ValidationFailed(map[string][]string{
			"file": {"ファイルが必要です"},
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return errors.InternalErrorWithLog(err, "FileHandler.Upload: failed to open file")
	}
	defer src.Close()

	// Detect content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	input := service.UploadInput{
		UserID:      currentUser.ID,
		TodoID:      todoID,
		FileName:    file.Filename,
		ContentType: contentType,
		FileSize:    file.Size,
		FileReader:  src,
	}

	uploadedFile, err := h.fileService.Upload(c.Request().Context(), input)
	if err != nil {
		return err
	}

	return response.Created(c, toFileResponse(uploadedFile, todoID))
}

// Download handles file download
// GET /api/v1/todos/:todo_id/files/:file_id
func (h *FileHandler) Download(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	fileID, err := ParseIDParam(c, "file_id")
	if err != nil {
		return err
	}

	reader, file, err := h.fileService.Download(c.Request().Context(), fileID, todoID, currentUser.ID)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Set headers for download
	c.Response().Header().Set("Content-Type", file.ContentType)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.OriginalName))

	return c.Stream(http.StatusOK, file.ContentType, reader)
}

// DownloadThumb handles thumbnail download
// GET /api/v1/todos/:todo_id/files/:file_id/thumb
func (h *FileHandler) DownloadThumb(c echo.Context) error {
	return h.downloadThumbnail(c, "thumb")
}

// DownloadMedium handles medium thumbnail download
// GET /api/v1/todos/:todo_id/files/:file_id/medium
func (h *FileHandler) DownloadMedium(c echo.Context) error {
	return h.downloadThumbnail(c, "medium")
}

func (h *FileHandler) downloadThumbnail(c echo.Context, size string) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	fileID, err := ParseIDParam(c, "file_id")
	if err != nil {
		return err
	}

	reader, _, err := h.fileService.DownloadThumbnail(c.Request().Context(), fileID, todoID, currentUser.ID, size)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Thumbnails are always served inline, not as attachment
	c.Response().Header().Set("Content-Type", "image/jpeg")
	c.Response().Header().Set("Cache-Control", "public, max-age=31536000")

	return c.Stream(http.StatusOK, "image/jpeg", reader)
}

// Delete handles file deletion
// DELETE /api/v1/todos/:todo_id/files/:file_id
func (h *FileHandler) Delete(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	fileID, err := ParseIDParam(c, "file_id")
	if err != nil {
		return err
	}

	if err := h.fileService.Delete(c.Request().Context(), fileID, todoID, currentUser.ID); err != nil {
		return err
	}

	return response.NoContent(c)
}
