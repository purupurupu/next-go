package repository

import (
	"strings"

	"gorm.io/gorm"

	"todo-api/internal/model"
)

// CategoryRepository handles database operations for categories
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// FindAllByUserID retrieves all categories for a user ordered by name
func (r *CategoryRepository) FindAllByUserID(userID int64) ([]model.Category, error) {
	var categories []model.Category
	result := r.db.
		Where("user_id = ?", userID).
		Order("name ASC").
		Find(&categories)
	return categories, result.Error
}

// FindByID retrieves a category by ID for a specific user
func (r *CategoryRepository) FindByID(id, userID int64) (*model.Category, error) {
	var category model.Category
	result := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		First(&category)
	if result.Error != nil {
		return nil, result.Error
	}
	return &category, nil
}

// ExistsByName checks if a category with the given name exists for a user (case-insensitive)
func (r *CategoryRepository) ExistsByName(name string, userID int64, excludeID *int64) (bool, error) {
	var count int64
	query := r.db.Model(&model.Category{}).
		Where("LOWER(name) = LOWER(?) AND user_id = ?", strings.TrimSpace(name), userID)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	result := query.Count(&count)
	return count > 0, result.Error
}

// Create creates a new category
func (r *CategoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

// Update updates an existing category
func (r *CategoryRepository) Update(category *model.Category) error {
	return r.db.Save(category).Error
}

// Delete deletes a category and nullifies related todos' category_id
func (r *CategoryRepository) Delete(id, userID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First, verify category exists and belongs to user
		var category model.Category
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&category).Error; err != nil {
			return err
		}

		// Nullify category_id in related todos
		if err := tx.Model(&model.Todo{}).
			Where("category_id = ? AND user_id = ?", id, userID).
			Update("category_id", nil).Error; err != nil {
			return err
		}

		// Delete the category
		return tx.Delete(&category).Error
	})
}

// IncrementTodosCount increments the todos_count for a category
func (r *CategoryRepository) IncrementTodosCount(categoryID int64) error {
	return r.db.Model(&model.Category{}).
		Where("id = ?", categoryID).
		UpdateColumn("todos_count", gorm.Expr("todos_count + ?", 1)).Error
}

// DecrementTodosCount decrements the todos_count for a category (minimum 0)
func (r *CategoryRepository) DecrementTodosCount(categoryID int64) error {
	return r.db.Model(&model.Category{}).
		Where("id = ?", categoryID).
		UpdateColumn("todos_count", gorm.Expr("GREATEST(todos_count - ?, 0)", 1)).Error
}

// RecalculateTodosCount recalculates the todos_count from actual todos
func (r *CategoryRepository) RecalculateTodosCount(categoryID int64) error {
	return r.db.Exec(`
		UPDATE categories
		SET todos_count = (
			SELECT COUNT(*) FROM todos WHERE category_id = ?
		)
		WHERE id = ?
	`, categoryID, categoryID).Error
}
