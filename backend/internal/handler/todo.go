package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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
	Title       string  `json:"title" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=10000"`
	CategoryID  *int64  `json:"category_id"`
	Priority    *string `json:"priority" validate:"omitempty,oneof=low medium high"`
	Status      *string `json:"status" validate:"omitempty,oneof=pending in_progress completed"`
	DueDate     *string `json:"due_date" validate:"omitempty"`
	Position    *int    `json:"position"`
}

// UpdateTodoRequest represents the request body for updating a todo
type UpdateTodoRequest struct {
	Title       *string  `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string  `json:"description" validate:"omitempty,max=10000"`
	CategoryID  *int64   `json:"category_id"`
	Completed   *bool    `json:"completed"`
	Priority    *string  `json:"priority" validate:"omitempty,oneof=low medium high"`
	Status      *string  `json:"status" validate:"omitempty,oneof=pending in_progress completed"`
	DueDate     *string  `json:"due_date"`
	Position    *int     `json:"position"`
	TagIDs      *[]int64 `json:"tag_ids"`
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
	CategoryID  *int64           `json:"category_id"`
	Title       string           `json:"title"`
	Description *string          `json:"description"`
	Completed   bool             `json:"completed"`
	Position    *int             `json:"position"`
	Priority    string           `json:"priority"`
	Status      string           `json:"status"`
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
		CategoryID:  todo.CategoryID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		Position:    todo.Position,
		Priority:    todo.Priority.String(),
		Status:      todo.Status.String(),
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

	return c.JSON(http.StatusOK, todoResponses)
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

	return c.JSON(http.StatusOK, toTodoResponse(todo))
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
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Priority:    req.Priority,
		Status:      req.Status,
		DueDate:     req.DueDate,
		Position:    req.Position,
	})
	if err != nil {
		return err
	}

	return response.Created(c, toTodoResponse(todo))
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
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Completed:   req.Completed,
		Priority:    req.Priority,
		Status:      req.Status,
		DueDate:     req.DueDate,
		Position:    req.Position,
		TagIDs:      req.TagIDs,
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", id)
		}
		return err
	}

	return response.OK(c, toTodoResponse(todo))
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

	return response.NoContent(c)
}

// SearchMetaResponse represents the meta information in search response
type SearchMetaResponse struct {
	Total          int64          `json:"total"`
	CurrentPage    int            `json:"current_page"`
	TotalPages     int            `json:"total_pages"`
	PerPage        int            `json:"per_page"`
	FiltersApplied map[string]any `json:"filters_applied"`
}

// SearchSuggestion represents a search suggestion
type SearchSuggestion struct {
	Type           string   `json:"type"`
	Message        string   `json:"message"`
	CurrentFilters []string `json:"current_filters,omitempty"`
}

// SearchResponse represents the response for search endpoint
type SearchResponse struct {
	Data        []TodoResponse     `json:"data"`
	Meta        SearchMetaResponse `json:"meta"`
	Suggestions []SearchSuggestion `json:"suggestions,omitempty"`
}

// Search searches todos with filters
// GET /api/v1/todos/search
func (h *TodoHandler) Search(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	// Parse search parameters
	searchInput, err := h.parseSearchParams(c)
	if err != nil {
		return err
	}
	searchInput.UserID = currentUser.ID

	// Set defaults before calling service (service also validates)
	if searchInput.Page < 1 {
		searchInput.Page = 1
	}
	if searchInput.PerPage < 1 {
		searchInput.PerPage = 20
	}
	if searchInput.PerPage > 100 {
		searchInput.PerPage = 100
	}

	// Execute search
	result, err := h.todoService.Search(*searchInput)
	if err != nil {
		return err
	}

	// Convert to response format
	todoResponses := make([]TodoResponse, len(result.Todos))
	for i, todo := range result.Todos {
		todoResponses[i] = toTodoResponse(&todo)
	}

	// Calculate total pages
	totalPages := 0
	if result.Total > 0 {
		totalPages = int((result.Total + int64(searchInput.PerPage) - 1) / int64(searchInput.PerPage))
	}

	// Build active filters map
	filtersApplied := h.buildFiltersApplied(searchInput)

	// Generate suggestions for empty results
	suggestions := h.generateSuggestions(result, searchInput)

	return c.JSON(http.StatusOK, SearchResponse{
		Data: todoResponses,
		Meta: SearchMetaResponse{
			Total:          result.Total,
			CurrentPage:    searchInput.Page,
			TotalPages:     totalPages,
			PerPage:        searchInput.PerPage,
			FiltersApplied: filtersApplied,
		},
		Suggestions: suggestions,
	})
}

// parseSearchParams parses query parameters for search
func (h *TodoHandler) parseSearchParams(c echo.Context) (*service.SearchInput, error) {
	input := &service.SearchInput{}

	// Text search query
	input.Query = c.QueryParam("q")

	// Parse status filter (supports both status[] and status with comma-separated values)
	statusParams := c.QueryParams()["status[]"]
	if len(statusParams) == 0 {
		statusCSV := c.QueryParam("status")
		if statusCSV != "" {
			statusParams = strings.Split(statusCSV, ",")
		}
	}
	for _, s := range statusParams {
		switch strings.TrimSpace(s) {
		case "pending":
			input.Statuses = append(input.Statuses, model.StatusPending)
		case "in_progress":
			input.Statuses = append(input.Statuses, model.StatusInProgress)
		case "completed":
			input.Statuses = append(input.Statuses, model.StatusCompleted)
		}
	}

	// Parse priority filter (supports both priority[] and priority)
	priorityParams := c.QueryParams()["priority[]"]
	if len(priorityParams) == 0 {
		priorityCSV := c.QueryParam("priority")
		if priorityCSV != "" {
			priorityParams = strings.Split(priorityCSV, ",")
		}
	}
	if len(priorityParams) > 0 {
		// Use the first priority value
		switch strings.TrimSpace(priorityParams[0]) {
		case "low":
			p := model.PriorityLow
			input.Priority = &p
		case "medium":
			p := model.PriorityMedium
			input.Priority = &p
		case "high":
			p := model.PriorityHigh
			input.Priority = &p
		}
	}

	// Parse category filter (-1 or "null" means uncategorized)
	categoryID := c.QueryParam("category_id")
	if categoryID != "" {
		if categoryID == "-1" || categoryID == "null" {
			input.CategoryIDNull = true
		} else {
			id, err := strconv.ParseInt(categoryID, 10, 64)
			if err == nil {
				input.CategoryID = &id
			}
		}
	}

	// Parse tag filter (supports both tag_ids[] and tag_ids)
	tagParams := c.QueryParams()["tag_ids[]"]
	if len(tagParams) == 0 {
		tagCSV := c.QueryParam("tag_ids")
		if tagCSV != "" {
			tagParams = strings.Split(tagCSV, ",")
		}
	}
	for _, idStr := range tagParams {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err == nil {
			input.TagIDs = append(input.TagIDs, id)
		}
	}

	// Tag mode (any or all)
	input.TagMode = c.QueryParam("tag_mode")

	// Parse date range filters
	if dueDateFrom := c.QueryParam("due_date_from"); dueDateFrom != "" {
		t, err := time.Parse("2006-01-02", dueDateFrom)
		if err == nil {
			input.DueDateFrom = &t
		}
	}
	if dueDateTo := c.QueryParam("due_date_to"); dueDateTo != "" {
		t, err := time.Parse("2006-01-02", dueDateTo)
		if err == nil {
			input.DueDateTo = &t
		}
	}

	// Sort parameters
	input.SortBy = c.QueryParam("sort_by")
	input.SortOrder = c.QueryParam("sort_order")

	// Pagination
	if page := c.QueryParam("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			input.Page = p
		}
	}
	if perPage := c.QueryParam("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil {
			input.PerPage = pp
		}
	}

	return input, nil
}

// buildFiltersApplied builds a map of applied filters for the response
func (h *TodoHandler) buildFiltersApplied(input *service.SearchInput) map[string]any {
	filters := make(map[string]any)

	if input.Query != "" {
		filters["search"] = input.Query
	}

	if len(input.Statuses) > 0 {
		statusStrings := make([]string, len(input.Statuses))
		for i, s := range input.Statuses {
			switch s {
			case model.StatusPending:
				statusStrings[i] = "pending"
			case model.StatusInProgress:
				statusStrings[i] = "in_progress"
			case model.StatusCompleted:
				statusStrings[i] = "completed"
			}
		}
		filters["status"] = statusStrings
	}

	if input.Priority != nil {
		switch *input.Priority {
		case model.PriorityLow:
			filters["priority"] = "low"
		case model.PriorityMedium:
			filters["priority"] = "medium"
		case model.PriorityHigh:
			filters["priority"] = "high"
		}
	}

	if input.CategoryID != nil {
		filters["category_id"] = *input.CategoryID
	} else if input.CategoryIDNull {
		filters["category_id"] = nil
	}

	if len(input.TagIDs) > 0 {
		filters["tag_ids"] = input.TagIDs
		filters["tag_mode"] = input.TagMode
	}

	if input.DueDateFrom != nil {
		filters["due_date_from"] = input.DueDateFrom.Format("2006-01-02")
	}
	if input.DueDateTo != nil {
		filters["due_date_to"] = input.DueDateTo.Format("2006-01-02")
	}

	return filters
}

// generateSuggestions generates suggestions for empty search results
func (h *TodoHandler) generateSuggestions(result *service.SearchResult, input *service.SearchInput) []SearchSuggestion {
	if result.Total > 0 || !result.HasFilters {
		return nil
	}

	var suggestions []SearchSuggestion

	if input.Query != "" {
		suggestions = append(suggestions, SearchSuggestion{
			Type:    "spelling",
			Message: "スペルを確認してください",
		})
	}

	if result.HasFilters {
		var currentFilters []string
		if input.Query != "" {
			currentFilters = append(currentFilters, "検索キーワード")
		}
		if len(input.Statuses) > 0 {
			currentFilters = append(currentFilters, "ステータス")
		}
		if input.Priority != nil {
			currentFilters = append(currentFilters, "優先度")
		}
		if input.CategoryID != nil || input.CategoryIDNull {
			currentFilters = append(currentFilters, "カテゴリ")
		}
		if len(input.TagIDs) > 0 {
			currentFilters = append(currentFilters, "タグ")
		}
		if input.DueDateFrom != nil || input.DueDateTo != nil {
			currentFilters = append(currentFilters, "期限")
		}

		suggestions = append(suggestions, SearchSuggestion{
			Type:           "reduce_filters",
			Message:        "フィルター条件を減らしてみてください",
			CurrentFilters: currentFilters,
		})
	}

	return suggestions
}
