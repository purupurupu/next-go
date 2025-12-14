package service

import (
	"time"

	"todo-api/internal/errors"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/pkg/util"
)

// TodoService handles todo business logic
type TodoService struct {
	todoRepo     *repository.TodoRepository
	categoryRepo *repository.CategoryRepository
}

// NewTodoService creates a new TodoService
func NewTodoService(
	todoRepo *repository.TodoRepository,
	categoryRepo *repository.CategoryRepository,
) *TodoService {
	return &TodoService{
		todoRepo:     todoRepo,
		categoryRepo: categoryRepo,
	}
}

// CreateInput represents input for creating a todo
type CreateInput struct {
	UserID      int64
	Title       string
	Description *string
	CategoryID  *int64
	Priority    *int
	Status      *int
	DueDate     *string
	Position    *int
}

// UpdateInput represents input for updating a todo
type UpdateInput struct {
	Title       *string
	Description *string
	CategoryID  *int64
	Completed   *bool
	Priority    *int
	Status      *int
	DueDate     *string
	Position    *int
}

// Create creates a new todo
func (s *TodoService) Create(input CreateInput) (*model.Todo, error) {
	// Validate category ownership if provided
	if input.CategoryID != nil {
		if err := s.validateCategoryOwnership(*input.CategoryID, input.UserID); err != nil {
			return nil, err
		}
	}

	// Parse and validate due date
	dueDate, err := s.parseDueDate(input.DueDate, true)
	if err != nil {
		return nil, err
	}

	// Create todo model
	todo := &model.Todo{
		UserID:      input.UserID,
		Title:       input.Title,
		Description: input.Description,
		CategoryID:  input.CategoryID,
		DueDate:     dueDate,
		Position:    input.Position,
		Priority:    s.resolvePriority(input.Priority),
		Status:      s.resolveStatus(input.Status),
	}

	if err := s.todoRepo.Create(todo); err != nil {
		return nil, errors.InternalErrorWithLog(err, "TodoService.Create: failed to create todo")
	}

	// Increment category todo count if category is set
	if todo.CategoryID != nil {
		_ = s.categoryRepo.IncrementTodosCount(*todo.CategoryID)
	}

	// Reload to get auto-generated position and relations
	return s.todoRepo.FindByIDWithRelations(todo.ID, input.UserID)
}

// Update updates an existing todo
func (s *TodoService) Update(todoID, userID int64, input UpdateInput) (*model.Todo, error) {
	// Get existing todo
	todo, err := s.todoRepo.FindByID(todoID, userID)
	if err != nil {
		return nil, err // Let handler handle gorm.ErrRecordNotFound
	}

	oldCategoryID := todo.CategoryID

	// Apply text field updates
	s.applyTextFields(todo, input)

	// Handle category update
	if err := s.applyCategory(todo, input.CategoryID, userID); err != nil {
		return nil, err
	}

	// Sync status and completed
	s.syncStatusAndCompleted(todo, input)

	// Apply other fields
	if input.Priority != nil {
		todo.Priority = model.Priority(*input.Priority)
	}

	// Parse and validate due date
	if input.DueDate != nil {
		dueDate, err := s.parseDueDate(input.DueDate, false)
		if err != nil {
			return nil, err
		}
		todo.DueDate = dueDate
	}

	if input.Position != nil {
		todo.Position = input.Position
	}

	// Save changes
	if err := s.todoRepo.Update(todo); err != nil {
		return nil, errors.InternalErrorWithLog(err, "TodoService.Update: failed to update todo")
	}

	// Update category counts if changed
	s.updateCategoryCounts(oldCategoryID, todo.CategoryID)

	// Reload with relations
	return s.todoRepo.FindByIDWithRelations(todoID, userID)
}

// Delete deletes a todo
func (s *TodoService) Delete(todoID, userID int64) error {
	// Get todo first to update category count
	todo, err := s.todoRepo.FindByID(todoID, userID)
	if err != nil {
		return err // Let handler handle gorm.ErrRecordNotFound
	}

	categoryID := todo.CategoryID

	if err := s.todoRepo.Delete(todoID, userID); err != nil {
		return err
	}

	// Decrement category count if category was set
	if categoryID != nil {
		_ = s.categoryRepo.DecrementTodosCount(*categoryID)
	}

	return nil
}

// validateCategoryOwnership checks if a category belongs to the user
func (s *TodoService) validateCategoryOwnership(categoryID, userID int64) error {
	valid, err := s.todoRepo.ValidateCategoryOwnership(categoryID, userID)
	if err != nil {
		return errors.InternalErrorWithLog(err, "TodoService: failed to validate category ownership")
	}
	if !valid {
		return errors.ValidationFailed(map[string][]string{
			"category_id": {"Category not found or not owned by user"},
		})
	}
	return nil
}

// parseDueDate parses a date string and optionally checks if it's in the past
func (s *TodoService) parseDueDate(dateStr *string, checkPast bool) (*time.Time, error) {
	if dateStr == nil || *dateStr == "" {
		return nil, nil
	}

	dueDate, err := util.ParseDate(*dateStr)
	if err != nil {
		return nil, errors.ValidationFailed(map[string][]string{
			"due_date": {"Invalid date format. Use YYYY-MM-DD"},
		})
	}

	if checkPast && util.IsBeforeToday(*dueDate) {
		return nil, errors.ValidationFailed(map[string][]string{
			"due_date": {"Due date cannot be in the past"},
		})
	}

	return dueDate, nil
}

// applyTextFields applies title and description updates
func (s *TodoService) applyTextFields(todo *model.Todo, input UpdateInput) {
	if input.Title != nil {
		todo.Title = *input.Title
	}
	if input.Description != nil {
		todo.Description = input.Description
	}
}

// applyCategory handles category updates including setting to null
func (s *TodoService) applyCategory(todo *model.Todo, categoryID *int64, userID int64) error {
	if categoryID == nil {
		return nil
	}

	if *categoryID == 0 {
		// Setting to null
		todo.CategoryID = nil
		return nil
	}

	// Validate category ownership
	if err := s.validateCategoryOwnership(*categoryID, userID); err != nil {
		return err
	}
	todo.CategoryID = categoryID
	return nil
}

// syncStatusAndCompleted syncs status and completed fields
func (s *TodoService) syncStatusAndCompleted(todo *model.Todo, input UpdateInput) {
	if input.Completed != nil {
		todo.Completed = *input.Completed
		// Update status based on completed flag
		if *input.Completed {
			todo.Status = model.StatusCompleted
		} else if todo.Status == model.StatusCompleted {
			todo.Status = model.StatusPending
		}
	}

	if input.Status != nil {
		todo.Status = model.Status(*input.Status)
		// Update completed based on status
		todo.Completed = (todo.Status == model.StatusCompleted)
	}
}

// updateCategoryCounts updates category counts when category changes
func (s *TodoService) updateCategoryCounts(oldCategoryID, newCategoryID *int64) {
	if !s.categoryChanged(oldCategoryID, newCategoryID) {
		return
	}
	if oldCategoryID != nil {
		_ = s.categoryRepo.DecrementTodosCount(*oldCategoryID)
	}
	if newCategoryID != nil {
		_ = s.categoryRepo.IncrementTodosCount(*newCategoryID)
	}
}

// categoryChanged checks if category has changed
func (s *TodoService) categoryChanged(oldID, newID *int64) bool {
	if oldID == nil && newID != nil {
		return true
	}
	if oldID != nil && newID == nil {
		return true
	}
	if oldID != nil && newID != nil && *oldID != *newID {
		return true
	}
	return false
}

// resolvePriority returns the priority value or default
func (s *TodoService) resolvePriority(p *int) model.Priority {
	if p != nil {
		return model.Priority(*p)
	}
	return model.PriorityMedium
}

// resolveStatus returns the status value or default
func (s *TodoService) resolveStatus(st *int) model.Status {
	if st != nil {
		return model.Status(*st)
	}
	return model.StatusPending
}
