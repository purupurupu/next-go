package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"todo-api/internal/errors"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/pkg/response"
	"todo-api/pkg/util"
)

// TodoHandler handles todo-related endpoints
type TodoHandler struct {
	todoRepo     *repository.TodoRepository
	categoryRepo *repository.CategoryRepository
}

// NewTodoHandler creates a new TodoHandler
func NewTodoHandler(todoRepo *repository.TodoRepository, categoryRepo *repository.CategoryRepository) *TodoHandler {
	return &TodoHandler{
		todoRepo:     todoRepo,
		categoryRepo: categoryRepo,
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
		DueDate:     util.FormatDate(todo.DueDate),
		CreatedAt:   util.FormatRFC3339(todo.CreatedAt),
		UpdatedAt:   util.FormatRFC3339(todo.UpdatedAt),
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
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
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

	return c.JSON(http.StatusOK, map[string]any{
		"todos": todoResponses,
	})
}

// Show retrieves a specific todo by ID
// GET /api/v1/todos/:id
func (h *TodoHandler) Show(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	id, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	todo, err := h.todoRepo.FindByIDWithRelations(id, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", id)
		}
		return errors.InternalError()
	}

	return c.JSON(http.StatusOK, map[string]any{
		"todo": toTodoResponse(todo),
	})
}

// Create creates a new todo
// POST /api/v1/todos
func (h *TodoHandler) Create(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	var req CreateTodoRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
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
		dueDate, err = util.ParseDate(*req.Todo.DueDate)
		if err != nil {
			return errors.ValidationFailed(map[string][]string{
				"due_date": {"Invalid date format. Use YYYY-MM-DD"},
			})
		}
		// Check if due date is in the past (only for creation)
		if util.IsBeforeToday(*dueDate) {
			return errors.ValidationFailed(map[string][]string{
				"due_date": {"Due date cannot be in the past"},
			})
		}
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

	// Increment category todo count if category is set
	if todo.CategoryID != nil {
		if err := h.categoryRepo.IncrementTodosCount(*todo.CategoryID); err != nil {
			// Log but don't fail the request
		}
	}

	// Reload to get auto-generated position
	todo, err = h.todoRepo.FindByIDWithRelations(todo.ID, currentUser.ID)
	if err != nil {
		return errors.InternalError()
	}

	return response.Created(c, map[string]any{
		"todo": toTodoResponse(todo),
	}, "Todo created successfully")
}

// Update updates an existing todo
// PATCH /api/v1/todos/:id
func (h *TodoHandler) Update(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	id, err := ParseIDParam(c, "id")
	if err != nil {
		return err
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
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	// Apply updates
	if req.Todo.Title != nil {
		todo.Title = *req.Todo.Title
	}

	if req.Todo.Description != nil {
		todo.Description = req.Todo.Description
	}

	// Handle category_id update (including setting to null)
	oldCategoryID := todo.CategoryID
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

	// Check if category changed
	categoryChanged := false
	if oldCategoryID == nil && todo.CategoryID != nil {
		categoryChanged = true
	} else if oldCategoryID != nil && todo.CategoryID == nil {
		categoryChanged = true
	} else if oldCategoryID != nil && todo.CategoryID != nil && *oldCategoryID != *todo.CategoryID {
		categoryChanged = true
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
		dueDate, err := util.ParseDate(*req.Todo.DueDate)
		if err != nil {
			return errors.ValidationFailed(map[string][]string{
				"due_date": {"Invalid date format. Use YYYY-MM-DD"},
			})
		}
		todo.DueDate = dueDate
	}

	if req.Todo.Position != nil {
		todo.Position = req.Todo.Position
	}

	if err := h.todoRepo.Update(todo); err != nil {
		return errors.InternalError()
	}

	// Update category counts if category changed
	if categoryChanged {
		if oldCategoryID != nil {
			_ = h.categoryRepo.DecrementTodosCount(*oldCategoryID)
		}
		if todo.CategoryID != nil {
			_ = h.categoryRepo.IncrementTodosCount(*todo.CategoryID)
		}
	}

	// Reload with relations
	todo, err = h.todoRepo.FindByIDWithRelations(id, currentUser.ID)
	if err != nil {
		return errors.InternalError()
	}

	return response.Success(c, map[string]any{
		"todo": toTodoResponse(todo),
	})
}

// Delete removes a todo
// DELETE /api/v1/todos/:id
func (h *TodoHandler) Delete(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	id, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	// Get todo first to update category count
	todo, err := h.todoRepo.FindByID(id, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", id)
		}
		return errors.InternalError()
	}

	categoryID := todo.CategoryID

	if err := h.todoRepo.Delete(id, currentUser.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", id)
		}
		return errors.InternalError()
	}

	// Decrement category count if category was set
	if categoryID != nil {
		_ = h.categoryRepo.DecrementTodosCount(*categoryID)
	}

	return response.NoContent(c)
}

// UpdateOrder updates the positions of multiple todos
// PATCH /api/v1/todos/update_order
func (h *TodoHandler) UpdateOrder(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	var req UpdateOrderRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
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
