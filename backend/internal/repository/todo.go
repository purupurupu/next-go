package repository

import (
	"todo-api/internal/model"

	"gorm.io/gorm"
)

// OrderUpdate represents a single position update for a todo
type OrderUpdate struct {
	ID       int64 `json:"id"`
	Position int   `json:"position"`
}

// TodoRepository handles database operations for todos
type TodoRepository struct {
	db *gorm.DB
}

// NewTodoRepository creates a new TodoRepository
func NewTodoRepository(db *gorm.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// FindAllByUserID retrieves all todos for a user
func (r *TodoRepository) FindAllByUserID(userID int64) ([]model.Todo, error) {
	var todos []model.Todo
	result := r.db.
		Where("user_id = ?", userID).
		Order("COALESCE(position, 0) ASC, created_at DESC").
		Find(&todos)
	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

// FindAllByUserIDWithRelations retrieves all todos for a user with preloaded relations
func (r *TodoRepository) FindAllByUserIDWithRelations(userID int64) ([]model.Todo, error) {
	var todos []model.Todo
	result := r.db.
		Preload("Category").
		Preload("Tags").
		Where("user_id = ?", userID).
		Order("COALESCE(position, 0) ASC, created_at DESC").
		Find(&todos)
	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

// FindByID retrieves a todo by ID for a specific user
func (r *TodoRepository) FindByID(id, userID int64) (*model.Todo, error) {
	var todo model.Todo
	result := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		First(&todo)
	if result.Error != nil {
		return nil, result.Error
	}
	return &todo, nil
}

// FindByIDWithRelations retrieves a todo by ID with preloaded relations
func (r *TodoRepository) FindByIDWithRelations(id, userID int64) (*model.Todo, error) {
	var todo model.Todo
	result := r.db.
		Preload("Category").
		Preload("Tags").
		Where("id = ? AND user_id = ?", id, userID).
		First(&todo)
	if result.Error != nil {
		return nil, result.Error
	}
	return &todo, nil
}

// Create creates a new todo
func (r *TodoRepository) Create(todo *model.Todo) error {
	return r.db.Create(todo).Error
}

// Update updates an existing todo
func (r *TodoRepository) Update(todo *model.Todo) error {
	return r.db.Save(todo).Error
}

// Delete deletes a todo by ID for a specific user
func (r *TodoRepository) Delete(id, userID int64) error {
	result := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Todo{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateOrder updates the positions of multiple todos
func (r *TodoRepository) UpdateOrder(userID int64, updates []OrderUpdate) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, update := range updates {
			result := tx.Model(&model.Todo{}).
				Where("id = ? AND user_id = ?", update.ID, userID).
				Update("position", update.Position)
			if result.Error != nil {
				return result.Error
			}
			// Skip if todo not found (could be deleted) instead of failing
		}
		return nil
	})
}

// Count returns the total number of todos for a user
func (r *TodoRepository) Count(userID int64) (int64, error) {
	var count int64
	result := r.db.Model(&model.Todo{}).
		Where("user_id = ?", userID).
		Count(&count)
	return count, result.Error
}

// ExistsByID checks if a todo exists for a specific user
func (r *TodoRepository) ExistsByID(id, userID int64) (bool, error) {
	var count int64
	result := r.db.Model(&model.Todo{}).
		Where("id = ? AND user_id = ?", id, userID).
		Count(&count)
	return count > 0, result.Error
}

// ValidateCategoryOwnership checks if a category belongs to a user
func (r *TodoRepository) ValidateCategoryOwnership(categoryID, userID int64) (bool, error) {
	var count int64
	result := r.db.Model(&model.Category{}).
		Where("id = ? AND user_id = ?", categoryID, userID).
		Count(&count)
	return count > 0, result.Error
}
