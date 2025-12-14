package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"todo-api/internal/errors"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/pkg/response"
	"todo-api/pkg/util"
)

// CategoryHandler handles category-related endpoints
type CategoryHandler struct {
	categoryRepo *repository.CategoryRepository
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(categoryRepo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		categoryRepo: categoryRepo,
	}
}

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	Category struct {
		Name  string `json:"name" validate:"required,notblank,max=50"`
		Color string `json:"color" validate:"required,hexcolor"`
	} `json:"category" validate:"required"`
}

// UpdateCategoryRequest represents the request body for updating a category
type UpdateCategoryRequest struct {
	Category struct {
		Name  *string `json:"name" validate:"omitempty,notblank,max=50"`
		Color *string `json:"color" validate:"omitempty,hexcolor"`
	} `json:"category" validate:"required"`
}

// CategoryResponse represents a category in API responses
type CategoryResponse struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	TodoCount int    `json:"todo_count"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// toCategoryResponse converts a model.Category to CategoryResponse
func toCategoryResponse(category *model.Category) CategoryResponse {
	return CategoryResponse{
		ID:        category.ID,
		UserID:    category.UserID,
		Name:      category.Name,
		Color:     category.Color,
		TodoCount: category.TodosCount,
		CreatedAt: util.FormatRFC3339(category.CreatedAt),
		UpdatedAt: util.FormatRFC3339(category.UpdatedAt),
	}
}

// List retrieves all categories for the authenticated user
// GET /api/v1/categories
func (h *CategoryHandler) List(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	categories, err := h.categoryRepo.FindAllByUserID(currentUser.ID)
	if err != nil {
		return errors.InternalError()
	}

	categoryResponses := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		categoryResponses[i] = toCategoryResponse(&category)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"categories": categoryResponses,
	})
}

// Show retrieves a specific category by ID
// GET /api/v1/categories/:id
func (h *CategoryHandler) Show(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	id, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	category, err := h.categoryRepo.FindByID(id, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Category", id)
		}
		return errors.InternalError()
	}

	return c.JSON(http.StatusOK, map[string]any{
		"category": toCategoryResponse(category),
	})
}

// Create creates a new category
// POST /api/v1/categories
func (h *CategoryHandler) Create(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	var req CreateCategoryRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	// Check for duplicate name (case-insensitive)
	exists, err := h.categoryRepo.ExistsByName(req.Category.Name, currentUser.ID, nil)
	if err != nil {
		return errors.InternalError()
	}
	if exists {
		return errors.DuplicateResource("Category", "name")
	}

	category := &model.Category{
		UserID: currentUser.ID,
		Name:   req.Category.Name,
		Color:  req.Category.Color,
	}

	if err := h.categoryRepo.Create(category); err != nil {
		return errors.InternalError()
	}

	return response.Created(c, map[string]any{
		"category": toCategoryResponse(category),
	}, "Category created successfully")
}

// Update updates an existing category
// PUT /api/v1/categories/:id
// PATCH /api/v1/categories/:id
func (h *CategoryHandler) Update(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	id, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	category, err := h.categoryRepo.FindByID(id, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Category", id)
		}
		return errors.InternalError()
	}

	var req UpdateCategoryRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	// Check for duplicate name if name is being changed
	if req.Category.Name != nil && *req.Category.Name != category.Name {
		exists, err := h.categoryRepo.ExistsByName(*req.Category.Name, currentUser.ID, &id)
		if err != nil {
			return errors.InternalError()
		}
		if exists {
			return errors.DuplicateResource("Category", "name")
		}
		category.Name = *req.Category.Name
	}

	if req.Category.Color != nil {
		category.Color = *req.Category.Color
	}

	if err := h.categoryRepo.Update(category); err != nil {
		return errors.InternalError()
	}

	return c.JSON(http.StatusOK, map[string]any{
		"category": toCategoryResponse(category),
	})
}

// Delete removes a category
// DELETE /api/v1/categories/:id
func (h *CategoryHandler) Delete(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	id, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	if err := h.categoryRepo.Delete(id, currentUser.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Category", id)
		}
		return errors.InternalError()
	}

	return response.NoContent(c)
}
