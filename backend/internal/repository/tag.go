package repository

import (
	"strings"

	"gorm.io/gorm"

	"todo-api/internal/model"
)

// TagRepository handles database operations for tags
type TagRepository struct {
	db *gorm.DB
}

// NewTagRepository creates a new TagRepository
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// FindAllByUserID retrieves all tags for a user ordered by name
func (r *TagRepository) FindAllByUserID(userID int64) ([]model.Tag, error) {
	var tags []model.Tag
	result := r.db.
		Where("user_id = ?", userID).
		Order("name ASC").
		Find(&tags)
	return tags, result.Error
}

// FindByID retrieves a tag by ID for a specific user
func (r *TagRepository) FindByID(id, userID int64) (*model.Tag, error) {
	var tag model.Tag
	result := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		First(&tag)
	if result.Error != nil {
		return nil, result.Error
	}
	return &tag, nil
}

// ExistsByName checks if a tag with the given name exists for a user
// Note: Tag names are normalized to lowercase in BeforeSave hook
func (r *TagRepository) ExistsByName(name string, userID int64, excludeID *int64) (bool, error) {
	var count int64
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	query := r.db.Model(&model.Tag{}).
		Where("name = ? AND user_id = ?", normalizedName, userID)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	result := query.Count(&count)
	return count > 0, result.Error
}

// Create creates a new tag
func (r *TagRepository) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

// Update updates an existing tag
func (r *TagRepository) Update(tag *model.Tag) error {
	return r.db.Save(tag).Error
}

// Delete deletes a tag (CASCADE will remove from todo_tags)
func (r *TagRepository) Delete(id, userID int64) error {
	result := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Tag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindByIDs retrieves tags by IDs for a specific user
func (r *TagRepository) FindByIDs(ids []int64, userID int64) ([]model.Tag, error) {
	var tags []model.Tag
	if len(ids) == 0 {
		return tags, nil
	}
	result := r.db.
		Where("id IN ? AND user_id = ?", ids, userID).
		Find(&tags)
	return tags, result.Error
}

// ValidateTagOwnership checks if all tag IDs belong to the user
func (r *TagRepository) ValidateTagOwnership(tagIDs []int64, userID int64) (bool, error) {
	if len(tagIDs) == 0 {
		return true, nil
	}
	var count int64
	result := r.db.Model(&model.Tag{}).
		Where("id IN ? AND user_id = ?", tagIDs, userID).
		Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count == int64(len(tagIDs)), nil
}
