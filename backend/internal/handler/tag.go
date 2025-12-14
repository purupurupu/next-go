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

// TagHandler handles tag-related endpoints
type TagHandler struct {
	tagRepo *repository.TagRepository
}

// NewTagHandler creates a new TagHandler
func NewTagHandler(tagRepo *repository.TagRepository) *TagHandler {
	return &TagHandler{
		tagRepo: tagRepo,
	}
}

// CreateTagRequest represents the request body for creating a tag
type CreateTagRequest struct {
	Tag struct {
		Name  string  `json:"name" validate:"required,notblank,max=30"`
		Color *string `json:"color" validate:"omitempty,hexcolor"`
	} `json:"tag" validate:"required"`
}

// UpdateTagRequest represents the request body for updating a tag
type UpdateTagRequest struct {
	Tag struct {
		Name  *string `json:"name" validate:"omitempty,notblank,max=30"`
		Color *string `json:"color" validate:"omitempty,hexcolor"`
	} `json:"tag" validate:"required"`
}

// TagResponse represents a tag in API responses
type TagResponse struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"user_id"`
	Name      string  `json:"name"`
	Color     *string `json:"color"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// toTagResponse converts a model.Tag to TagResponse
func toTagResponse(tag *model.Tag) TagResponse {
	return TagResponse{
		ID:        tag.ID,
		UserID:    tag.UserID,
		Name:      tag.Name,
		Color:     tag.Color,
		CreatedAt: util.FormatRFC3339(tag.CreatedAt),
		UpdatedAt: util.FormatRFC3339(tag.UpdatedAt),
	}
}

// List retrieves all tags for the authenticated user
// GET /api/v1/tags
func (h *TagHandler) List(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	tags, err := h.tagRepo.FindAllByUserID(currentUser.ID)
	if err != nil {
		return errors.InternalError()
	}

	tagResponses := make([]TagResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = toTagResponse(&tag)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"tags": tagResponses,
	})
}

// Show retrieves a specific tag by ID
// GET /api/v1/tags/:id
func (h *TagHandler) Show(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	id, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	tag, err := h.tagRepo.FindByID(id, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Tag", id)
		}
		return errors.InternalError()
	}

	return c.JSON(http.StatusOK, map[string]any{
		"tag": toTagResponse(tag),
	})
}

// Create creates a new tag
// POST /api/v1/tags
func (h *TagHandler) Create(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	var req CreateTagRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	// Check for duplicate name (names are normalized to lowercase in BeforeSave)
	exists, err := h.tagRepo.ExistsByName(req.Tag.Name, currentUser.ID, nil)
	if err != nil {
		return errors.InternalError()
	}
	if exists {
		return errors.DuplicateResource("Tag", "name")
	}

	tag := &model.Tag{
		UserID: currentUser.ID,
		Name:   req.Tag.Name, // BeforeSave will normalize to lowercase
		Color:  req.Tag.Color,
	}

	if err := h.tagRepo.Create(tag); err != nil {
		return errors.InternalError()
	}

	return response.Created(c, map[string]any{
		"tag": toTagResponse(tag),
	}, "Tag created successfully")
}

// Update updates an existing tag
// PATCH /api/v1/tags/:id
func (h *TagHandler) Update(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	id, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	tag, err := h.tagRepo.FindByID(id, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Tag", id)
		}
		return errors.InternalError()
	}

	var req UpdateTagRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	// Check for duplicate name if name is being changed
	if req.Tag.Name != nil {
		exists, err := h.tagRepo.ExistsByName(*req.Tag.Name, currentUser.ID, &id)
		if err != nil {
			return errors.InternalError()
		}
		if exists {
			return errors.DuplicateResource("Tag", "name")
		}
		tag.Name = *req.Tag.Name // BeforeSave will normalize
	}

	if req.Tag.Color != nil {
		tag.Color = req.Tag.Color
	}

	if err := h.tagRepo.Update(tag); err != nil {
		return errors.InternalError()
	}

	return response.Success(c, map[string]any{
		"tag": toTagResponse(tag),
	})
}

// Delete removes a tag
// DELETE /api/v1/tags/:id
func (h *TagHandler) Delete(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	id, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	if err := h.tagRepo.Delete(id, currentUser.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Tag", id)
		}
		return errors.InternalError()
	}

	return response.NoContent(c)
}
