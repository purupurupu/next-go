package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"todo-api/internal/errors"
	"todo-api/internal/middleware"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/internal/validator"
	"todo-api/pkg/response"
)

// TodoHandler handles todo-related endpoints
type TodoHandler struct {
	todoRepo *repository.TodoRepository
}

// NewTodoHandler creates a new TodoHandler
func NewTodoHandler(todoRepo *repository.TodoRepository) *TodoHandler {
	return &TodoHandler{
		todoRepo: todoRepo,
	}
}

// CreateTodoRequest represents the request body for creating a todo
type CreateTodoRequest struct {
	Todo struct {
		Title       string  `json:"title" validate:"required,min=1,max=255"`
		Description *string `json:"description" validate:"omitempty,max=10000"`
		CategoryID  *int64  `json:"category_id"`
		Priority    *int    `json:"priority" validate:"omitempty,min=0,max=2"`
		Status      *int    `json:"status" validate:"omitempty,min=0,max=2"`
		DueDate     *string `json:"due_date" validate:"omitempty"`
		Position    *int    `json:"position"`
	} `json:"todo" validate:"required"`
}

// UpdateTodoRequest represents the request body for updating a todo
type UpdateTodoRequest struct {
	Todo struct {
		Title       *string `json:"title" validate:"omitempty,min=1,max=255"`
		Description *string `json:"description" validate:"omitempty,max=10000"`
		CategoryID  *int64  `json:"category_id"`
		Completed   *bool   `json:"completed"`
		Priority    *int    `json:"priority" validate:"omitempty,min=0,max=2"`
		Status      *int    `json:"status" validate:"omitempty,min=0,max=2"`
		DueDate     *string `json:"due_date"`
		Position    *int    `json:"position"`
	} `json:"todo" validate:"required"`
}

// UpdateOrderRequest represents the request body for updating todo positions
type UpdateOrderRequest struct {
	Todos []struct {
		ID       int64 `json:"id" validate:"required"`
		Position int   `json:"position" validate:"required,min=0"`
	} `json:"todos" validate:"required,dive"`
}

// TodoResponse represents a todo in API responses
type TodoResponse struct {
	ID          int64            `json:"id"`
	UserID      int64            `json:"user_id"`
	CategoryID  *int64           `json:"category_id"`
	Title       string           `json:"title"`
	Description *string          `json:"description"`
	Completed   bool             `json:"completed"`
	Position    *int             `json:"position"`
	Priority    int              `json:"priority"`
	Status      int              `json:"status"`
	DueDate     *string          `json:"due_date"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
	Category    *CategorySummary `json:"category,omitempty"`
	Tags        []TagSummary     `json:"tags,omitempty"`
}

// CategorySummary represents a category summary in todo responses
type CategorySummary struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// TagSummary represents a tag summary in todo responses
type TagSummary struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

// toTodoResponse converts a model.Todo to TodoResponse
func toTodoResponse(todo *model.Todo) TodoResponse {
	resp := TodoResponse{
		ID:          todo.ID,
		UserID:      todo.UserID,
		CategoryID:  todo.CategoryID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		Position:    todo.Position,
		Priority:    int(todo.Priority),
		Status:      int(todo.Status),
		CreatedAt:   todo.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   todo.UpdatedAt.Format(time.RFC3339),
	}

	if todo.DueDate != nil {
		dueDate := todo.DueDate.Format("2006-01-02")
		resp.DueDate = &dueDate
	}

	if todo.Category != nil {
		resp.Category = &CategorySummary{
			ID:    todo.Category.ID,
			Name:  todo.Category.Name,
			Color: todo.Category.Color,
		}
	}

	if len(todo.Tags) > 0 {
		resp.Tags = make([]TagSummary, len(todo.Tags))
		for i, tag := range todo.Tags {
			resp.Tags[i] = TagSummary{
				ID:    tag.ID,
				Name:  tag.Name,
				Color: tag.Color,
			}
		}
	}

	return resp
}

// List retrieves all todos for the authenticated user
// GET /api/v1/todos
func (h *TodoHandler) List(c echo.Context) error {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return errors.AuthenticationFailed("User not authenticated")
	}

	todos, err := h.todoRepo.FindAllByUserIDWithRelations(currentUser.ID)
	if err != nil {
		return errors.InternalError()
	}

	// Convert to response format
	todoResponses := make([]TodoResponse, len(todos))
	for i, todo := range todos {
		todoResponses[i] = toTodoResponse(&todo)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"todos": todoResponses,
	})
}

// Show retrieves a specific todo by ID
// GET /api/v1/todos/:id
func (h *TodoHandler) Show(c echo.Context) error {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return errors.AuthenticationFailed("User not authenticated")
	}

	// Parse todo ID from path
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return errors.ValidationFailed(map[string][]string{
			"id": {"Invalid todo ID"},
		})
	}

	todo, err := h.todoRepo.FindByIDWithRelations(id, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", id)
		}
		return errors.InternalError()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"todo": toTodoResponse(todo),
	})
}

// Create creates a new todo
// POST /api/v1/todos
func (h *TodoHandler) Create(c echo.Context) error {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return errors.AuthenticationFailed("User not authenticated")
	}

	var req CreateTodoRequest
	if err := c.Bind(&req); err != nil {
		return errors.ValidationFailed(map[string][]string{
			"body": {"Invalid request body"},
		})
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		return errors.ValidationFailed(validator.FormatValidationErrors(err))
	}

	// Validate category ownership if provided
	if req.Todo.CategoryID != nil {
		valid, err := h.todoRepo.ValidateCategoryOwnership(*req.Todo.CategoryID, currentUser.ID)
		if err != nil {
			return errors.InternalError()
		}
		if !valid {
			return errors.ValidationFailed(map[string][]string{
				"category_id": {"Category not found or not owned by user"},
			})
		}
	}

	// Parse due date if provided
	var dueDate *time.Time
	if req.Todo.DueDate != nil && *req.Todo.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.Todo.DueDate)
		if err != nil {
			return errors.ValidationFailed(map[string][]string{
				"due_date": {"Invalid date format. Use YYYY-MM-DD"},
			})
		}
		// Check if due date is in the past (only for creation)
		today := time.Now().Truncate(24 * time.Hour)
		if parsed.Before(today) {
			return errors.ValidationFailed(map[string][]string{
				"due_date": {"Due date cannot be in the past"},
			})
		}
		dueDate = &parsed
	}

	// Create todo
	todo := &model.Todo{
		UserID:      currentUser.ID,
		Title:       req.Todo.Title,
		Description: req.Todo.Description,
		CategoryID:  req.Todo.CategoryID,
		DueDate:     dueDate,
		Position:    req.Todo.Position,
	}

	// Set priority if provided
	if req.Todo.Priority != nil {
		todo.Priority = model.Priority(*req.Todo.Priority)
	} else {
		todo.Priority = model.PriorityMedium
	}

	// Set status if provided
	if req.Todo.Status != nil {
		todo.Status = model.Status(*req.Todo.Status)
	} else {
		todo.Status = model.StatusPending
	}

	if err := h.todoRepo.Create(todo); err != nil {
		return errors.InternalError()
	}

	// Reload to get auto-generated position
	todo, err := h.todoRepo.FindByIDWithRelations(todo.ID, currentUser.ID)
	if err != nil {
		return errors.InternalError()
	}

	return response.Created(c, map[string]interface{}{
		"todo": toTodoResponse(todo),
	}, "Todo created successfully")
}

// Update updates an existing todo
// PATCH /api/v1/todos/:id
func (h *TodoHandler) Update(c echo.Context) error {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return errors.AuthenticationFailed("User not authenticated")
	}

	// Parse todo ID from path
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return errors.ValidationFailed(map[string][]string{
			"id": {"Invalid todo ID"},
		})
	}

	// Get existing todo
	todo, err := h.todoRepo.FindByID(id, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", id)
		}
		return errors.InternalError()
	}

	var req UpdateTodoRequest
	if err := c.Bind(&req); err != nil {
		return errors.ValidationFailed(map[string][]string{
			"body": {"Invalid request body"},
		})
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		return errors.ValidationFailed(validator.FormatValidationErrors(err))
	}

	// Apply updates
	if req.Todo.Title != nil {
		todo.Title = *req.Todo.Title
	}

	if req.Todo.Description != nil {
		todo.Description = req.Todo.Description
	}

	// Handle category_id update (including setting to null)
	if req.Todo.CategoryID != nil {
		if *req.Todo.CategoryID == 0 {
			// Setting to null
			todo.CategoryID = nil
		} else {
			// Validate category ownership
			valid, err := h.todoRepo.ValidateCategoryOwnership(*req.Todo.CategoryID, currentUser.ID)
			if err != nil {
				return errors.InternalError()
			}
			if !valid {
				return errors.ValidationFailed(map[string][]string{
					"category_id": {"Category not found or not owned by user"},
				})
			}
			todo.CategoryID = req.Todo.CategoryID
		}
	}

	if req.Todo.Completed != nil {
		todo.Completed = *req.Todo.Completed
		// Update status based on completed flag
		if *req.Todo.Completed {
			todo.Status = model.StatusCompleted
		} else if todo.Status == model.StatusCompleted {
			todo.Status = model.StatusPending
		}
	}

	if req.Todo.Priority != nil {
		todo.Priority = model.Priority(*req.Todo.Priority)
	}

	if req.Todo.Status != nil {
		todo.Status = model.Status(*req.Todo.Status)
		// Update completed based on status
		todo.Completed = (todo.Status == model.StatusCompleted)
	}

	if req.Todo.DueDate != nil {
		if *req.Todo.DueDate == "" {
			todo.DueDate = nil
		} else {
			parsed, err := time.Parse("2006-01-02", *req.Todo.DueDate)
			if err != nil {
				return errors.ValidationFailed(map[string][]string{
					"due_date": {"Invalid date format. Use YYYY-MM-DD"},
				})
			}
			todo.DueDate = &parsed
		}
	}

	if req.Todo.Position != nil {
		todo.Position = req.Todo.Position
	}

	if err := h.todoRepo.Update(todo); err != nil {
		return errors.InternalError()
	}

	// Reload with relations
	todo, err = h.todoRepo.FindByIDWithRelations(id, currentUser.ID)
	if err != nil {
		return errors.InternalError()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"todo": toTodoResponse(todo),
	})
}

// Delete removes a todo
// DELETE /api/v1/todos/:id
func (h *TodoHandler) Delete(c echo.Context) error {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return errors.AuthenticationFailed("User not authenticated")
	}

	// Parse todo ID from path
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return errors.ValidationFailed(map[string][]string{
			"id": {"Invalid todo ID"},
		})
	}

	if err := h.todoRepo.Delete(id, currentUser.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", id)
		}
		return errors.InternalError()
	}

	return response.NoContent(c)
}

// UpdateOrder updates the positions of multiple todos
// PATCH /api/v1/todos/update_order
func (h *TodoHandler) UpdateOrder(c echo.Context) error {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		return errors.AuthenticationFailed("User not authenticated")
	}

	var req UpdateOrderRequest
	if err := c.Bind(&req); err != nil {
		return errors.ValidationFailed(map[string][]string{
			"body": {"Invalid request body"},
		})
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		return errors.ValidationFailed(validator.FormatValidationErrors(err))
	}

	// Convert to repository format
	updates := make([]repository.OrderUpdate, len(req.Todos))
	for i, todo := range req.Todos {
		updates[i] = repository.OrderUpdate{
			ID:       todo.ID,
			Position: todo.Position,
		}
	}

	if err := h.todoRepo.UpdateOrder(currentUser.ID, updates); err != nil {
		return errors.InternalError()
	}

	return response.Message(c, "Order updated successfully")
}
