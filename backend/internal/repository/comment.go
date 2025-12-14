package repository

import (
	"todo-api/internal/model"

	"gorm.io/gorm"
)

// CommentRepository handles database operations for comments
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository creates a new CommentRepository
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// FindAllByCommentable retrieves all comments for a specific resource
// Excludes soft-deleted comments, ordered by created_at ASC
func (r *CommentRepository) FindAllByCommentable(commentableType string, commentableID int64) ([]model.Comment, error) {
	var comments []model.Comment
	result := r.db.
		Preload("User").
		Where("commentable_type = ? AND commentable_id = ?", commentableType, commentableID).
		Order("created_at ASC").
		Find(&comments)
	return comments, result.Error
}

// FindByID retrieves a comment by ID (includes soft-deleted for ownership check)
func (r *CommentRepository) FindByID(id int64) (*model.Comment, error) {
	var comment model.Comment
	result := r.db.
		Unscoped(). // Include soft-deleted
		Preload("User").
		Where("id = ?", id).
		First(&comment)
	if result.Error != nil {
		return nil, result.Error
	}
	return &comment, nil
}

// FindByIDWithoutDeleted retrieves a comment by ID (excludes soft-deleted)
func (r *CommentRepository) FindByIDWithoutDeleted(id int64) (*model.Comment, error) {
	var comment model.Comment
	result := r.db.
		Preload("User").
		Where("id = ?", id).
		First(&comment)
	if result.Error != nil {
		return nil, result.Error
	}
	return &comment, nil
}

// Create creates a new comment
func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

// Update updates an existing comment
func (r *CommentRepository) Update(comment *model.Comment) error {
	return r.db.Save(comment).Error
}

// SoftDelete soft deletes a comment by setting deleted_at
func (r *CommentRepository) SoftDelete(id int64) error {
	return r.db.Delete(&model.Comment{}, id).Error
}

// ExistsByID checks if a comment exists (excludes soft-deleted)
func (r *CommentRepository) ExistsByID(id int64) (bool, error) {
	var count int64
	result := r.db.Model(&model.Comment{}).
		Where("id = ?", id).
		Count(&count)
	return count > 0, result.Error
}
