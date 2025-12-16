package repository

import (
	"todo-api/internal/model"

	"gorm.io/gorm"
)

// FileRepository handles database operations for files
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new FileRepository
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// FindByID retrieves a file by ID
func (r *FileRepository) FindByID(id int64) (*model.File, error) {
	var file model.File
	result := r.db.
		Preload("User").
		Where("id = ?", id).
		First(&file)
	if result.Error != nil {
		return nil, result.Error
	}
	return &file, nil
}

// FindByAttachable retrieves all files for a specific resource
func (r *FileRepository) FindByAttachable(attachableType string, attachableID int64) ([]model.File, error) {
	var files []model.File
	result := r.db.
		Where("attachable_type = ? AND attachable_id = ?", attachableType, attachableID).
		Order("created_at DESC").
		Find(&files)
	return files, result.Error
}

// FindByAttachableWithUser retrieves all files for a specific resource with user info
func (r *FileRepository) FindByAttachableWithUser(attachableType string, attachableID int64) ([]model.File, error) {
	var files []model.File
	result := r.db.
		Preload("User").
		Where("attachable_type = ? AND attachable_id = ?", attachableType, attachableID).
		Order("created_at DESC").
		Find(&files)
	return files, result.Error
}

// Create creates a new file record
func (r *FileRepository) Create(file *model.File) error {
	return r.db.Create(file).Error
}

// Delete deletes a file record by ID
func (r *FileRepository) Delete(id int64) error {
	return r.db.Delete(&model.File{}, id).Error
}

// DeleteByAttachable deletes all files for a specific resource
func (r *FileRepository) DeleteByAttachable(attachableType string, attachableID int64) error {
	return r.db.
		Where("attachable_type = ? AND attachable_id = ?", attachableType, attachableID).
		Delete(&model.File{}).Error
}

// ExistsByID checks if a file exists by ID
func (r *FileRepository) ExistsByID(id int64) (bool, error) {
	var count int64
	result := r.db.Model(&model.File{}).
		Where("id = ?", id).
		Count(&count)
	return count > 0, result.Error
}
