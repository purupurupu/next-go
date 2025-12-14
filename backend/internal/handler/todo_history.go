package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"todo-api/internal/errors"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/pkg/util"
)

// TodoHistoryHandler handles todo history endpoints
type TodoHistoryHandler struct {
	historyRepo *repository.TodoHistoryRepository
	todoRepo    *repository.TodoRepository
}

// NewTodoHistoryHandler creates a new TodoHistoryHandler
func NewTodoHistoryHandler(historyRepo *repository.TodoHistoryRepository, todoRepo *repository.TodoRepository) *TodoHistoryHandler {
	return &TodoHistoryHandler{
		historyRepo: historyRepo,
		todoRepo:    todoRepo,
	}
}

// HistoryUserSummary represents user info in history response
type HistoryUserSummary struct {
	ID    int64   `json:"id"`
	Name  *string `json:"name"`
	Email string  `json:"email"`
}

// HistoryResponse represents a single history entry in API response
type HistoryResponse struct {
	ID                  int64               `json:"id"`
	TodoID              int64               `json:"todo_id"`
	Action              string              `json:"action"`
	Changes             json.RawMessage     `json:"changes"`
	User                *HistoryUserSummary `json:"user,omitempty"`
	CreatedAt           string              `json:"created_at"`
	HumanReadableChange string              `json:"human_readable_change"`
}

// HistoryListResponse represents the response for history list endpoint
type HistoryListResponse struct {
	Histories []HistoryResponse `json:"histories"`
	Meta      HistoryMeta       `json:"meta"`
}

// HistoryMeta represents pagination metadata
type HistoryMeta struct {
	Total       int64 `json:"total"`
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	PerPage     int   `json:"per_page"`
}

// List retrieves histories for a specific todo
// GET /api/v1/todos/:todo_id/histories
func (h *TodoHistoryHandler) List(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	todoID, err := ParseIDParam(c, "todo_id")
	if err != nil {
		return err
	}

	// Verify todo exists and belongs to user
	_, err = h.todoRepo.FindByID(todoID, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Todo", todoID)
		}
		return errors.InternalErrorWithLog(err, "TodoHistoryHandler.List: failed to verify todo")
	}

	// Parse pagination params
	page := 1
	perPage := 20
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if pp := c.QueryParam("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	// Fetch histories
	histories, total, err := h.historyRepo.FindByTodoIDWithUser(todoID, page, perPage)
	if err != nil {
		return errors.InternalErrorWithLog(err, "TodoHistoryHandler.List: failed to fetch histories")
	}

	// Convert to response format
	historyResponses := make([]HistoryResponse, len(histories))
	for i, history := range histories {
		historyResponses[i] = toHistoryResponse(&history)
	}

	// Calculate total pages
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}

	return c.JSON(http.StatusOK, HistoryListResponse{
		Histories: historyResponses,
		Meta: HistoryMeta{
			Total:       total,
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     perPage,
		},
	})
}

// toHistoryResponse converts a model.TodoHistory to HistoryResponse
func toHistoryResponse(history *model.TodoHistory) HistoryResponse {
	resp := HistoryResponse{
		ID:                  history.ID,
		TodoID:              history.TodoID,
		Action:              string(history.Action),
		Changes:             history.Changes,
		CreatedAt:           util.FormatRFC3339(history.CreatedAt),
		HumanReadableChange: generateHumanReadableChange(history),
	}

	if history.User != nil {
		resp.User = &HistoryUserSummary{
			ID:    history.User.ID,
			Name:  history.User.Name,
			Email: history.User.Email,
		}
	}

	return resp
}

// =============================================================================
// Human Readable Change Generation (Japanese)
// =============================================================================

// generateHumanReadableChange generates a human-readable description of the change in Japanese
func generateHumanReadableChange(history *model.TodoHistory) string {
	switch history.Action {
	case model.ActionCreated:
		return "Todoが作成されました"
	case model.ActionDeleted:
		return "Todoが削除されました"
	case model.ActionStatusChanged:
		return generateStatusChangeMessage(history.Changes)
	case model.ActionPriorityChanged:
		return generatePriorityChangeMessage(history.Changes)
	case model.ActionUpdated:
		return generateUpdateMessage(history.Changes)
	default:
		return "変更されました"
	}
}

// generateStatusChangeMessage generates message for status change
func generateStatusChangeMessage(changes json.RawMessage) string {
	var data map[string]interface{}
	if err := json.Unmarshal(changes, &data); err != nil {
		return "ステータスが変更されました"
	}

	if statusArr, ok := data["status"].([]interface{}); ok && len(statusArr) == 2 {
		oldStatus := translateStatus(fmt.Sprint(statusArr[0]))
		newStatus := translateStatus(fmt.Sprint(statusArr[1]))
		return fmt.Sprintf("ステータスが「%s」から「%s」に変更されました", oldStatus, newStatus)
	}

	return "ステータスが変更されました"
}

// generatePriorityChangeMessage generates message for priority change
func generatePriorityChangeMessage(changes json.RawMessage) string {
	var data map[string]interface{}
	if err := json.Unmarshal(changes, &data); err != nil {
		return "優先度が変更されました"
	}

	if priorityArr, ok := data["priority"].([]interface{}); ok && len(priorityArr) == 2 {
		oldPriority := translatePriority(fmt.Sprint(priorityArr[0]))
		newPriority := translatePriority(fmt.Sprint(priorityArr[1]))
		return fmt.Sprintf("優先度が「%s」から「%s」に変更されました", oldPriority, newPriority)
	}

	return "優先度が変更されました"
}

// generateUpdateMessage generates message for general updates
func generateUpdateMessage(changes json.RawMessage) string {
	var data map[string]interface{}
	if err := json.Unmarshal(changes, &data); err != nil {
		return "Todoが更新されました"
	}

	messages := []string{}

	// Title change
	if titleArr, ok := data["title"].([]interface{}); ok && len(titleArr) == 2 {
		messages = append(messages, fmt.Sprintf("タイトルが「%v」から「%v」に変更されました", titleArr[0], titleArr[1]))
	}

	// Description change
	if descArr, ok := data["description"].([]interface{}); ok && len(descArr) == 2 {
		if descArr[0] == nil && descArr[1] != nil {
			messages = append(messages, "説明が追加されました")
		} else if descArr[0] != nil && descArr[1] == nil {
			messages = append(messages, "説明が削除されました")
		} else {
			messages = append(messages, "説明が更新されました")
		}
	}

	// Due date change
	if dueDateArr, ok := data["due_date"].([]interface{}); ok && len(dueDateArr) == 2 {
		if dueDateArr[0] == nil && dueDateArr[1] != nil {
			messages = append(messages, fmt.Sprintf("期限が「%v」に設定されました", dueDateArr[1]))
		} else if dueDateArr[0] != nil && dueDateArr[1] == nil {
			messages = append(messages, "期限が削除されました")
		} else {
			messages = append(messages, fmt.Sprintf("期限が「%v」から「%v」に変更されました", dueDateArr[0], dueDateArr[1]))
		}
	}

	// Category change
	if catArr, ok := data["category_id"].([]interface{}); ok && len(catArr) == 2 {
		if catArr[0] == nil && catArr[1] != nil {
			messages = append(messages, "カテゴリが設定されました")
		} else if catArr[0] != nil && catArr[1] == nil {
			messages = append(messages, "カテゴリが削除されました")
		} else {
			messages = append(messages, "カテゴリが変更されました")
		}
	}

	// Completed change
	if compArr, ok := data["completed"].([]interface{}); ok && len(compArr) == 2 {
		if compArr[1] == true {
			messages = append(messages, "完了としてマークされました")
		} else {
			messages = append(messages, "未完了に戻されました")
		}
	}

	// Status change (in general update)
	if statusArr, ok := data["status"].([]interface{}); ok && len(statusArr) == 2 {
		oldStatus := translateStatus(fmt.Sprint(statusArr[0]))
		newStatus := translateStatus(fmt.Sprint(statusArr[1]))
		messages = append(messages, fmt.Sprintf("ステータスが「%s」から「%s」に変更されました", oldStatus, newStatus))
	}

	// Priority change (in general update)
	if priorityArr, ok := data["priority"].([]interface{}); ok && len(priorityArr) == 2 {
		oldPriority := translatePriority(fmt.Sprint(priorityArr[0]))
		newPriority := translatePriority(fmt.Sprint(priorityArr[1]))
		messages = append(messages, fmt.Sprintf("優先度が「%s」から「%s」に変更されました", oldPriority, newPriority))
	}

	if len(messages) == 0 {
		return "Todoが更新されました"
	}

	if len(messages) == 1 {
		return messages[0]
	}

	return fmt.Sprintf("%d件の項目が更新されました", len(messages))
}

// translateStatus translates status to Japanese
func translateStatus(status string) string {
	switch status {
	case "pending":
		return "未着手"
	case "in_progress":
		return "進行中"
	case "completed":
		return "完了"
	default:
		return status
	}
}

// translatePriority translates priority to Japanese
func translatePriority(priority string) string {
	switch priority {
	case "low":
		return "低"
	case "medium":
		return "中"
	case "high":
		return "高"
	default:
		return priority
	}
}
