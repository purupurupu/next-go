package repository

import (
	"fmt"
	"strings"
	"time"

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

// SearchInput represents the input for repository search operation
type SearchInput struct {
	UserID         int64
	Query          string
	Statuses       []model.Status
	Priority       *model.Priority
	CategoryID     *int64
	CategoryIDNull bool
	TagIDs         []int64
	TagMode        string
	DueDateFrom    *time.Time
	DueDateTo      *time.Time
	SortBy         string
	SortOrder      string
	Page           int
	PerPage        int
}

// Search searches todos with filters and pagination
func (r *TodoRepository) Search(input SearchInput) ([]model.Todo, int64, error) {
	// Base query with user scope (required)
	query := r.db.Model(&model.Todo{}).Where("user_id = ?", input.UserID)

	// Text search (ILIKE for case-insensitive)
	if input.Query != "" {
		searchPattern := "%" + input.Query + "%"
		query = query.Where("(title ILIKE ? OR description ILIKE ?)", searchPattern, searchPattern)
	}

	// Status filter (multiple)
	if len(input.Statuses) > 0 {
		query = query.Where("status IN ?", input.Statuses)
	}

	// Priority filter
	if input.Priority != nil {
		query = query.Where("priority = ?", *input.Priority)
	}

	// Category filter
	if input.CategoryIDNull {
		query = query.Where("category_id IS NULL")
	} else if input.CategoryID != nil {
		query = query.Where("category_id = ?", *input.CategoryID)
	}

	// Tag filter
	if len(input.TagIDs) > 0 {
		if input.TagMode == "all" {
			// AND search: must have all specified tags
			for _, tagID := range input.TagIDs {
				query = query.Where("EXISTS (SELECT 1 FROM todo_tags WHERE todo_tags.todo_id = todos.id AND todo_tags.tag_id = ?)", tagID)
			}
		} else {
			// OR search (default): must have any of the specified tags
			query = query.Where("EXISTS (SELECT 1 FROM todo_tags WHERE todo_tags.todo_id = todos.id AND todo_tags.tag_id IN ?)", input.TagIDs)
		}
	}

	// Due date range filter
	if input.DueDateFrom != nil {
		query = query.Where("due_date >= ?", input.DueDateFrom)
	}
	if input.DueDateTo != nil {
		query = query.Where("due_date <= ?", input.DueDateTo)
	}

	// Get total count before pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	query = r.applySort(query, input.SortBy, input.SortOrder)

	// Apply pagination
	offset := (input.Page - 1) * input.PerPage
	query = query.Offset(offset).Limit(input.PerPage)

	// Preload relations and fetch
	var todos []model.Todo
	if err := query.Preload("Category").Preload("Tags").Find(&todos).Error; err != nil {
		return nil, 0, err
	}

	return todos, total, nil
}

// applySort applies sorting to the query
func (r *TodoRepository) applySort(query *gorm.DB, sortBy, sortOrder string) *gorm.DB {
	// Default sort field
	if sortBy == "" {
		sortBy = "created_at"
	}

	// Normalize sort order
	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	// Special handling for due_date: NULL values should be last
	if sortBy == "due_date" {
		if sortOrder == "ASC" {
			return query.Order("CASE WHEN due_date IS NULL THEN 1 ELSE 0 END, due_date ASC")
		}
		return query.Order("CASE WHEN due_date IS NULL THEN 1 ELSE 0 END, due_date DESC")
	}

	return query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
}
