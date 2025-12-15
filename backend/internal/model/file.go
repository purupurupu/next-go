package model

import (
	"time"
)

// AttachableType constants for files
const (
	AttachableTypeTodo = "Todo"
)

// FileType represents the type of file
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeDocument FileType = "document"
	FileTypeOther    FileType = "other"
)

// MaxFileSize is the maximum allowed file size in bytes (10MB)
const MaxFileSize = 10 * 1024 * 1024

// AllowedMimeTypes defines the allowed MIME types for file uploads
var AllowedMimeTypes = map[string]FileType{
	// Images
	"image/jpeg": FileTypeImage,
	"image/png":  FileTypeImage,
	"image/gif":  FileTypeImage,
	"image/webp": FileTypeImage,

	// Documents
	"application/pdf": FileTypeDocument,
	"text/plain":      FileTypeDocument,

	// MS Office - Word
	"application/msword": FileTypeDocument,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": FileTypeDocument,

	// MS Office - Excel
	"application/vnd.ms-excel":                                          FileTypeDocument,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": FileTypeDocument,

	// MS Office - PowerPoint
	"application/vnd.ms-powerpoint":                                              FileTypeDocument,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": FileTypeDocument,
}

// File represents an uploaded file attached to a resource
type File struct {
	ID             int64    `gorm:"primaryKey" json:"id"`
	UserID         int64    `gorm:"not null;index" json:"user_id"`
	AttachableType string   `gorm:"not null;size:50;index:idx_attachable" json:"attachable_type"`
	AttachableID   int64    `gorm:"not null;index:idx_attachable" json:"attachable_id"`
	OriginalName   string   `gorm:"not null;size:255" json:"original_name"`
	StoragePath    string   `gorm:"not null;size:500" json:"storage_path"`
	ContentType    string   `gorm:"not null;size:100" json:"content_type"`
	FileSize       int64    `gorm:"not null" json:"file_size"`
	FileType       FileType `gorm:"not null;size:20" json:"file_type"`
	ThumbPath      *string  `gorm:"size:500" json:"thumb_path,omitempty"`
	MediumPath     *string  `gorm:"size:500" json:"medium_path,omitempty"`

	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName returns the table name for the File model
func (File) TableName() string {
	return "files"
}

// IsImage returns true if the file is an image
func (f *File) IsImage() bool {
	return f.FileType == FileTypeImage
}

// IsOwnedBy checks if the file is owned by the given user
func (f *File) IsOwnedBy(userID int64) bool {
	return f.UserID == userID
}

// IsAllowedMimeType checks if the MIME type is allowed
func IsAllowedMimeType(mimeType string) bool {
	_, ok := AllowedMimeTypes[mimeType]
	return ok
}

// GetFileType returns the FileType for a given MIME type
func GetFileType(mimeType string) FileType {
	if ft, ok := AllowedMimeTypes[mimeType]; ok {
		return ft
	}
	return FileTypeOther
}

// IsValidAttachableType validates the attachable type for files
func IsValidFileAttachableType(t string) bool {
	return t == AttachableTypeTodo
}
