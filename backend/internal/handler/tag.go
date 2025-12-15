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
	Name  string  `json:"name" validate:"required,notblank,max=30"`
	Color *string `json:"color" validate:"omitempty,hexcolor"`
}

// UpdateTagRequest represents the request body for updating a tag
type UpdateTagRequest struct {
	Name  *string `json:"name" validate:"omitempty,notblank,max=30"`
	Color *string `json:"color" validate:"omitempty,hexcolor"`
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
		return errors.InternalErrorWithLog(err, "TagHandler.List: failed to fetch tags")
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
		return errors.InternalErrorWithLog(err, "TagHandler.Show: failed to fetch tag")
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
	exists, err := h.tagRepo.ExistsByName(req.Name, currentUser.ID, nil)
	if err != nil {
		return errors.InternalErrorWithLog(err, "TagHandler.Create: failed to check duplicate name")
	}
	if exists {
		return errors.DuplicateResource("Tag", "name")
	}

	tag := &model.Tag{
		UserID: currentUser.ID,
		Name:   req.Name, // BeforeSave will normalize to lowercase
		Color:  req.Color,
	}

	if err := h.tagRepo.Create(tag); err != nil {
		return errors.InternalErrorWithLog(err, "TagHandler.Create: failed to create tag")
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
		return errors.InternalErrorWithLog(err, "TagHandler.Update: failed to fetch tag")
	}

	var req UpdateTagRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	// Check for duplicate name if name is being changed
	if req.Name != nil {
		exists, err := h.tagRepo.ExistsByName(*req.Name, currentUser.ID, &id)
		if err != nil {
			return errors.InternalErrorWithLog(err, "TagHandler.Update: failed to check duplicate name")
		}
		if exists {
			return errors.DuplicateResource("Tag", "name")
		}
		tag.Name = *req.Name // BeforeSave will normalize
	}

	if req.Color != nil {
		tag.Color = req.Color
	}

	if err := h.tagRepo.Update(tag); err != nil {
		return errors.InternalErrorWithLog(err, "TagHandler.Update: failed to update tag")
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
		return errors.InternalErrorWithLog(err, "TagHandler.Delete: failed to delete tag")
	}

	return response.NoContent(c)
}
