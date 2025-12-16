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

// CommentHandler handles comment-related endpoints
type CommentHandler struct {
	commentRepo *repository.CommentRepository
	todoRepo    *repository.TodoRepository
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(commentRepo *repository.CommentRepository, todoRepo *repository.TodoRepository) *CommentHandler {
	return &CommentHandler{
		commentRepo: commentRepo,
		todoRepo:    todoRepo,
	}
}

// CreateCommentRequest represents the request body for creating a comment
type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// UpdateCommentRequest represents the request body for updating a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// CommentResponse represents a comment in API responses
type CommentResponse struct {
	ID        int64               `json:"id"`
	Content   string              `json:"content"`
	Editable  bool                `json:"editable"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
	User      *CommentUserSummary `json:"user,omitempty"`
}

// CommentUserSummary represents a user summary in comment responses
type CommentUserSummary struct {
	ID    int64   `json:"id"`
	Name  *string `json:"name"`
	Email string  `json:"email"`
}

// toCommentResponse converts a model.Comment to CommentResponse
func toCommentResponse(comment *model.Comment, currentUserID int64) CommentResponse {
	resp := CommentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		Editable:  comment.IsEditable() && comment.IsOwnedBy(currentUserID),
		CreatedAt: util.FormatRFC3339(comment.CreatedAt),
		UpdatedAt: util.FormatRFC3339(comment.UpdatedAt),
	}

	if comment.User != nil {
		resp.User = &CommentUserSummary{
			ID:    comment.User.ID,
			Name:  comment.User.Name,
			Email: comment.User.Email,
		}
	}

	return resp
}

// List retrieves all comments for a todo
// GET /api/v1/todos/:todo_id/comments
func (h *CommentHandler) List(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	// Verify todo exists and belongs to user
	if _, err := h.todoRepo.FindByID(todoID, currentUser.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", todoID)
		}
		return errors.InternalErrorWithLog(err, "CommentHandler.List: failed to fetch todo")
	}

	comments, err := h.commentRepo.FindAllByCommentable(model.CommentableTypeTodo, todoID)
	if err != nil {
		return errors.InternalErrorWithLog(err, "CommentHandler.List: failed to fetch comments")
	}

	commentResponses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		commentResponses[i] = toCommentResponse(&comment, currentUser.ID)
	}

	return c.JSON(http.StatusOK, commentResponses)
}

// Create creates a new comment for a todo
// POST /api/v1/todos/:todo_id/comments
func (h *CommentHandler) Create(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	// Verify todo exists and belongs to user
	if _, err := h.todoRepo.FindByID(todoID, currentUser.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", todoID)
		}
		return errors.InternalErrorWithLog(err, "CommentHandler.Create: failed to fetch todo")
	}

	var req CreateCommentRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	comment := &model.Comment{
		Content:         req.Content,
		UserID:          currentUser.ID,
		CommentableType: model.CommentableTypeTodo,
		CommentableID:   todoID,
	}

	if err := h.commentRepo.Create(comment); err != nil {
		return errors.InternalErrorWithLog(err, "CommentHandler.Create: failed to create comment")
	}

	// Reload with user relation
	comment, err = h.commentRepo.FindByIDWithoutDeleted(comment.ID)
	if err != nil {
		return errors.InternalErrorWithLog(err, "CommentHandler.Create: failed to reload comment")
	}

	return response.Created(c, toCommentResponse(comment, currentUser.ID))
}

// Update updates an existing comment
// PATCH /api/v1/todos/:todo_id/comments/:id
func (h *CommentHandler) Update(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	commentID, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	// Verify todo exists and belongs to user
	if _, err := h.todoRepo.FindByID(todoID, currentUser.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", todoID)
		}
		return errors.InternalErrorWithLog(err, "CommentHandler.Update: failed to fetch todo")
	}

	// Get comment
	comment, err := h.commentRepo.FindByIDWithoutDeleted(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Comment", commentID)
		}
		return errors.InternalErrorWithLog(err, "CommentHandler.Update: failed to fetch comment")
	}

	// Verify comment belongs to the todo
	if comment.CommentableType != model.CommentableTypeTodo || comment.CommentableID != todoID {
		return errors.NotFound("Comment", commentID)
	}

	// Verify ownership
	if !comment.IsOwnedBy(currentUser.ID) {
		return errors.AuthorizationFailed("Comment", "update")
	}

	// Verify edit time window
	if !comment.IsEditable() {
		return errors.EditTimeExpired()
	}

	var req UpdateCommentRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	comment.Content = req.Content

	if err := h.commentRepo.Update(comment); err != nil {
		return errors.InternalErrorWithLog(err, "CommentHandler.Update: failed to update comment")
	}

	return response.OK(c, toCommentResponse(comment, currentUser.ID))
}

// Delete soft-deletes a comment
// DELETE /api/v1/todos/:todo_id/comments/:id
func (h *CommentHandler) Delete(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	commentID, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	// Verify todo exists and belongs to user
	if _, err := h.todoRepo.FindByID(todoID, currentUser.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", todoID)
		}
		return errors.InternalErrorWithLog(err, "CommentHandler.Delete: failed to fetch todo")
	}

	// Get comment (include soft-deleted to check ownership)
	comment, err := h.commentRepo.FindByID(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Comment", commentID)
		}
		return errors.InternalErrorWithLog(err, "CommentHandler.Delete: failed to fetch comment")
	}

	// Verify comment belongs to the todo
	if comment.CommentableType != model.CommentableTypeTodo || comment.CommentableID != todoID {
		return errors.NotFound("Comment", commentID)
	}

	// Verify ownership
	if !comment.IsOwnedBy(currentUser.ID) {
		return errors.AuthorizationFailed("Comment", "delete")
	}

	// Check if already deleted
	if comment.DeletedAt.Valid {
		return errors.NotFound("Comment", commentID)
	}

	if err := h.commentRepo.SoftDelete(commentID); err != nil {
		return errors.InternalErrorWithLog(err, "CommentHandler.Delete: failed to delete comment")
	}

	return response.NoContent(c)
}
