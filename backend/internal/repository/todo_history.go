package repository

import (
	"todo-api/internal/model"

	"gorm.io/gorm"
)

// TodoHistoryRepository handles database operations for todo histories
type TodoHistoryRepository struct {
	db *gorm.DB
}

// NewTodoHistoryRepository creates a new TodoHistoryRepository
func NewTodoHistoryRepository(db *gorm.DB) *TodoHistoryRepository {
	return &TodoHistoryRepository{db: db}
}

// Create creates a new todo history record
func (r *TodoHistoryRepository) Create(history *model.TodoHistory) error {
	return r.db.Create(history).Error
}

// FindByTodoID retrieves histories for a specific todo with pagination
func (r *TodoHistoryRepository) FindByTodoID(todoID int64, page, perPage int) ([]model.TodoHistory, int64, error) {
	var histories []model.TodoHistory
	var total int64

	// Count total
	if err := r.db.Model(&model.TodoHistory{}).Where("todo_id = ?", todoID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * perPage
	result := r.db.
		Where("todo_id = ?", todoID).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&histories)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return histories, total, nil
}

// FindByTodoIDWithUser retrieves histories with preloaded user information
func (r *TodoHistoryRepository) FindByTodoIDWithUser(todoID int64, page, perPage int) ([]model.TodoHistory, int64, error) {
	var histories []model.TodoHistory
	var total int64

	// Count total
	if err := r.db.Model(&model.TodoHistory{}).Where("todo_id = ?", todoID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with user preload
	offset := (page - 1) * perPage
	result := r.db.
		Preload("User").
		Where("todo_id = ?", todoID).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&histories)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return histories, total, nil
}
