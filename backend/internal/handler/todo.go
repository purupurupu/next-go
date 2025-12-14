package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"todo-api/internal/errors"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/internal/service"
	"todo-api/pkg/response"
	"todo-api/pkg/util"
)

// TodoHandler handles todo-related endpoints
type TodoHandler struct {
	todoService *service.TodoService
	todoRepo    *repository.TodoRepository
}

// NewTodoHandler creates a new TodoHandler
func NewTodoHandler(todoService *service.TodoService, todoRepo *repository.TodoRepository) *TodoHandler {
	return &TodoHandler{
		todoService: todoService,
		todoRepo:    todoRepo,
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
		return errors.InternalErrorWithLog(err, "TodoHandler.List: failed to fetch todos")
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
		return errors.InternalErrorWithLog(err, "TodoHandler.Show: failed to fetch todo")
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

	todo, err := h.todoService.Create(service.CreateInput{
		UserID:      currentUser.ID,
		Title:       req.Todo.Title,
		Description: req.Todo.Description,
		CategoryID:  req.Todo.CategoryID,
		Priority:    req.Todo.Priority,
		Status:      req.Todo.Status,
		DueDate:     req.Todo.DueDate,
		Position:    req.Todo.Position,
	})
	if err != nil {
		return err
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

	var req UpdateTodoRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	todo, err := h.todoService.Update(id, currentUser.ID, service.UpdateInput{
		Title:       req.Todo.Title,
		Description: req.Todo.Description,
		CategoryID:  req.Todo.CategoryID,
		Completed:   req.Todo.Completed,
		Priority:    req.Todo.Priority,
		Status:      req.Todo.Status,
		DueDate:     req.Todo.DueDate,
		Position:    req.Todo.Position,
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", id)
		}
		return err
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

	if err := h.todoService.Delete(id, currentUser.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", id)
		}
		return err
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
		return errors.InternalErrorWithLog(err, "TodoHandler.UpdateOrder: failed to update order")
	}

	return response.Message(c, "Order updated successfully")
}
