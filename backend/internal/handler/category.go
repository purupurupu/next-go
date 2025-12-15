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
	Name  string `json:"name" validate:"required,notblank,max=50"`
	Color string `json:"color" validate:"required,hexcolor"`
}

// UpdateCategoryRequest represents the request body for updating a category
type UpdateCategoryRequest struct {
	Name  *string `json:"name" validate:"omitempty,notblank,max=50"`
	Color *string `json:"color" validate:"omitempty,hexcolor"`
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
		return errors.InternalErrorWithLog(err, "CategoryHandler.List: failed to fetch categories")
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
		return errors.InternalErrorWithLog(err, "CategoryHandler.Show: failed to fetch category")
	}

	return c.JSON(http.StatusOK, toCategoryResponse(category))
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
	exists, err := h.categoryRepo.ExistsByName(req.Name, currentUser.ID, nil)
	if err != nil {
		return errors.InternalErrorWithLog(err, "CategoryHandler.Create: failed to check duplicate name")
	}
	if exists {
		return errors.DuplicateResource("Category", "name")
	}

	category := &model.Category{
		UserID: currentUser.ID,
		Name:   req.Name,
		Color:  req.Color,
	}

	if err := h.categoryRepo.Create(category); err != nil {
		return errors.InternalErrorWithLog(err, "CategoryHandler.Create: failed to create category")
	}

	return response.Created(c, toCategoryResponse(category), "Category created successfully")
}

// Update updates an existing category
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
		return errors.InternalErrorWithLog(err, "CategoryHandler.Update: failed to fetch category")
	}

	var req UpdateCategoryRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	// Check for duplicate name if name is being changed
	if req.Name != nil && *req.Name != category.Name {
		exists, err := h.categoryRepo.ExistsByName(*req.Name, currentUser.ID, &id)
		if err != nil {
			return errors.InternalErrorWithLog(err, "CategoryHandler.Update: failed to check duplicate name")
		}
		if exists {
			return errors.DuplicateResource("Category", "name")
		}
		category.Name = *req.Name
	}

	if req.Color != nil {
		category.Color = *req.Color
	}

	if err := h.categoryRepo.Update(category); err != nil {
		return errors.InternalErrorWithLog(err, "CategoryHandler.Update: failed to update category")
	}

	return response.Success(c, toCategoryResponse(category))
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
		return errors.InternalErrorWithLog(err, "CategoryHandler.Delete: failed to delete category")
	}

	return response.NoContent(c)
}
