package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"todo-api/internal/model"
	"todo-api/internal/testutil"
)

// TestTodoList_Success tests successful todo list retrieval
func TestTodoList_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("todolist@example.com")
	f.CreateTodo(user.ID, "Todo 1")
	f.CreateTodo(user.ID, "Todo 2")

	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos", "", f.TodoHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	todos := testutil.JSONArrayResponse(t, rec)
	assert.Len(t, todos, 2)
}

// TestTodoList_Empty tests empty todo list
func TestTodoList_Empty(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("emptylist@example.com")

	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos", "", f.TodoHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	todos := testutil.JSONArrayResponse(t, rec)
	assert.Len(t, todos, 0)
}

// TestTodoList_UserScope tests that users only see their own todos
func TestTodoList_UserScope(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, token1 := f.CreateUser("user1@example.com")
	user2, _ := f.CreateUser("user2@example.com")

	f.CreateTodo(user1.ID, "User1 Todo")
	f.CreateTodo(user2.ID, "User2 Todo")

	rec, err := f.CallAuth(token1, http.MethodGet, "/api/v1/todos", "", f.TodoHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	todos := testutil.JSONArrayResponse(t, rec)
	assert.Len(t, todos, 1)

	firstTodo := testutil.TodoAt(todos, 0)
	assert.Equal(t, "User1 Todo", firstTodo["title"])
}

// TestTodoCreate_Success tests successful todo creation
func TestTodoCreate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("create@example.com")

	body := `{"todo":{"title":"New Todo","description":"A test todo","priority":2,"status":0}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", body, f.TodoHandler.Create)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	response := testutil.JSONResponse(t, rec)
	todo := testutil.ExtractTodoFromData(response)
	assert.Equal(t, "New Todo", todo["title"])
	assert.Equal(t, "A test todo", todo["description"])
	assert.Equal(t, float64(2), todo["priority"])
	assert.Equal(t, float64(0), todo["status"])
	assert.NotNil(t, todo["position"])
}

// TestTodoCreate_ValidationError tests todo creation with validation errors
func TestTodoCreate_ValidationError(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("validation@example.com")

	tests := []struct {
		name string
		body string
	}{
		{name: "missing title", body: `{"todo":{"description":"No title"}}`},
		{name: "empty title", body: `{"todo":{"title":""}}`},
		{name: "invalid priority", body: `{"todo":{"title":"Test","priority":5}}`},
		{name: "invalid status", body: `{"todo":{"title":"Test","status":10}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", tt.body, f.TodoHandler.Create)
			require.Error(t, err)
		})
	}
}

// TestTodoShow_Success tests successful todo retrieval by ID
func TestTodoShow_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("show@example.com")
	todo := f.CreateTodo(user.ID, "Show Me")

	rec, err := f.CallAuth(token, http.MethodGet, testutil.TodoPath(todo.ID), "", f.TodoHandler.Show)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	todoResp := testutil.ExtractTodo(response)
	assert.Equal(t, "Show Me", todoResp["title"])
}

// TestTodoShow_NotFound tests todo not found error
func TestTodoShow_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("notfound@example.com")

	_, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/99999", "", f.TodoHandler.Show)
	require.Error(t, err)
}

// TestTodoShow_OtherUserTodo tests that users cannot see other users' todos
func TestTodoShow_OtherUserTodo(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("owner@example.com")
	_, token2 := f.CreateUser("other@example.com")

	todo := f.CreateTodo(user1.ID, "User1's Todo")

	_, err := f.CallAuth(token2, http.MethodGet, testutil.TodoPath(todo.ID), "", f.TodoHandler.Show)
	require.Error(t, err)
}

// TestTodoUpdate_Success tests successful todo update
func TestTodoUpdate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("update@example.com")
	todo := f.CreateTodo(user.ID, "Original Title")

	body := `{"todo":{"title":"Updated Title","priority":2,"completed":true}}`
	rec, err := f.CallAuth(token, http.MethodPatch, testutil.TodoPath(todo.ID), body, f.TodoHandler.Update)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	todoResp := testutil.ExtractTodo(response)
	assert.Equal(t, "Updated Title", todoResp["title"])
	assert.Equal(t, float64(2), todoResp["priority"])
	assert.Equal(t, true, todoResp["completed"])
	assert.Equal(t, float64(2), todoResp["status"]) // Should be completed
}

// TestTodoUpdate_PartialUpdate tests partial todo update
func TestTodoUpdate_PartialUpdate(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("partial@example.com")
	desc := "Original description"
	todo := &model.Todo{
		UserID:      user.ID,
		Title:       "Original Title",
		Description: &desc,
		Priority:    model.PriorityLow,
	}
	require.NoError(t, f.DB.Create(todo).Error)

	body := `{"todo":{"title":"New Title"}}`
	rec, err := f.CallAuth(token, http.MethodPatch, testutil.TodoPath(todo.ID), body, f.TodoHandler.Update)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	todoResp := testutil.ExtractTodo(response)
	assert.Equal(t, "New Title", todoResp["title"])
	assert.Equal(t, "Original description", todoResp["description"])
	// Note: GORM applies default:1 when Priority=0 (zero value), so we expect 1 (medium)
	assert.Equal(t, float64(1), todoResp["priority"])
}

// TestTodoUpdate_OtherUserTodo tests that users cannot update other users' todos
func TestTodoUpdate_OtherUserTodo(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("updateowner@example.com")
	_, token2 := f.CreateUser("updateother@example.com")

	todo := f.CreateTodo(user1.ID, "User1's Todo")

	body := `{"todo":{"title":"Hacked!"}}`
	_, err := f.CallAuth(token2, http.MethodPatch, testutil.TodoPath(todo.ID), body, f.TodoHandler.Update)
	require.Error(t, err)
}

// TestTodoDelete_Success tests successful todo deletion
func TestTodoDelete_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("delete@example.com")
	todo := f.CreateTodo(user.ID, "Delete Me")

	rec, err := f.CallAuth(token, http.MethodDelete, testutil.TodoPath(todo.ID), "", f.TodoHandler.Delete)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify todo is deleted
	var count int64
	f.DB.Model(&model.Todo{}).Where("id = ?", todo.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestTodoDelete_NotFound tests deleting non-existent todo
func TestTodoDelete_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("deletenotfound@example.com")

	_, err := f.CallAuth(token, http.MethodDelete, "/api/v1/todos/99999", "", f.TodoHandler.Delete)
	require.Error(t, err)
}

// TestTodoDelete_OtherUserTodo tests that users cannot delete other users' todos
func TestTodoDelete_OtherUserTodo(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("deleteowner@example.com")
	_, token2 := f.CreateUser("deleteother@example.com")

	todo := f.CreateTodo(user1.ID, "User1's Todo")

	_, err := f.CallAuth(token2, http.MethodDelete, testutil.TodoPath(todo.ID), "", f.TodoHandler.Delete)
	require.Error(t, err)

	// Verify todo still exists
	var count int64
	f.DB.Model(&model.Todo{}).Where("id = ?", todo.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

// TestTodoUpdateOrder_Success tests successful order update
func TestTodoUpdateOrder_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("order@example.com")
	todo1 := f.CreateTodoWithPosition(user.ID, "Todo 1", 1)
	todo2 := f.CreateTodoWithPosition(user.ID, "Todo 2", 2)
	todo3 := f.CreateTodoWithPosition(user.ID, "Todo 3", 3)

	body := fmt.Sprintf(`{"todos":[{"id":%d,"position":3},{"id":%d,"position":1},{"id":%d,"position":2}]}`,
		todo1.ID, todo2.ID, todo3.ID)
	rec, err := f.CallAuth(token, http.MethodPatch, "/api/v1/todos/update_order", body, f.TodoHandler.UpdateOrder)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify positions
	var updated1, updated2, updated3 model.Todo
	f.DB.First(&updated1, todo1.ID)
	f.DB.First(&updated2, todo2.ID)
	f.DB.First(&updated3, todo3.ID)

	assert.Equal(t, 3, *updated1.Position)
	assert.Equal(t, 1, *updated2.Position)
	assert.Equal(t, 2, *updated3.Position)
}

// TestTodoCreate_WithDueDate tests todo creation with due date
func TestTodoCreate_WithDueDate(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("duedate@example.com")

	body := `{"todo":{"title":"Due Date Todo","due_date":"2030-12-31"}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", body, f.TodoHandler.Create)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	response := testutil.JSONResponse(t, rec)
	todo := testutil.ExtractTodoFromData(response)
	assert.Equal(t, "2030-12-31", todo["due_date"])
}

// TestTodoCreate_PastDueDate tests that past due date is rejected
func TestTodoCreate_PastDueDate(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("pastdue@example.com")

	body := `{"todo":{"title":"Past Due Todo","due_date":"2020-01-01"}}`
	_, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", body, f.TodoHandler.Create)
	require.Error(t, err)
}

// TestTodo_AutoPosition tests auto position assignment
func TestTodo_AutoPosition(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("autopos@example.com")

	// Create first todo
	body1 := `{"todo":{"title":"First Todo"}}`
	rec1, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", body1, f.TodoHandler.Create)
	require.NoError(t, err)

	response1 := testutil.JSONResponse(t, rec1)
	todo1 := testutil.ExtractTodoFromData(response1)
	pos1 := todo1["position"].(float64)

	// Create second todo
	body2 := `{"todo":{"title":"Second Todo"}}`
	rec2, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", body2, f.TodoHandler.Create)
	require.NoError(t, err)

	response2 := testutil.JSONResponse(t, rec2)
	todo2 := testutil.ExtractTodoFromData(response2)
	pos2 := todo2["position"].(float64)

	// Second todo should have higher position
	assert.Greater(t, pos2, pos1)
}

// ==================== Search Tests ====================

// TestTodoSearch_Basic tests basic search without filters
func TestTodoSearch_Basic(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("search@example.com")
	f.CreateTodo(user.ID, "Meeting notes")
	f.CreateTodo(user.ID, "Shopping list")
	f.CreateTodo(user.ID, "Project plan")

	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search", "", f.TodoHandler.Search)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 3)

	meta := response["meta"].(map[string]any)
	assert.Equal(t, float64(3), meta["total"])
	assert.Equal(t, float64(1), meta["current_page"])
}

// TestTodoSearch_TextQuery tests text search in title and description
func TestTodoSearch_TextQuery(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("textsearch@example.com")
	desc := "Important discussion about project"
	f.CreateTodoWithDetails(user.ID, "Meeting notes", testutil.TodoOptions{Description: &desc})
	f.CreateTodo(user.ID, "Shopping list")
	f.CreateTodo(user.ID, "Project plan")

	// Search by title
	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?q=meeting", "", f.TodoHandler.Search)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 1)

	// Search by description
	rec2, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?q=discussion", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response2 := testutil.JSONResponse(t, rec2)
	data2 := response2["data"].([]any)
	assert.Len(t, data2, 1)
}

// TestTodoSearch_StatusFilter tests status filter with multiple values
func TestTodoSearch_StatusFilter(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("statusfilter@example.com")
	f.CreateTodoWithDetails(user.ID, "Pending task", testutil.TodoOptions{Status: model.StatusPending})
	f.CreateTodoWithDetails(user.ID, "In progress task", testutil.TodoOptions{Status: model.StatusInProgress})
	f.CreateTodoWithDetails(user.ID, "Completed task", testutil.TodoOptions{Status: model.StatusCompleted})

	// Filter by single status
	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?status=pending", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 1)

	// Filter by multiple statuses (comma-separated)
	rec2, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?status=pending,in_progress", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response2 := testutil.JSONResponse(t, rec2)
	data2 := response2["data"].([]any)
	assert.Len(t, data2, 2)
}

// TestTodoSearch_PriorityFilter tests priority filter
func TestTodoSearch_PriorityFilter(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("priorityfilter@example.com")
	f.CreateTodoWithDetails(user.ID, "Low priority", testutil.TodoOptions{Priority: model.PriorityLow})
	f.CreateTodoWithDetails(user.ID, "Medium priority", testutil.TodoOptions{Priority: model.PriorityMedium})
	f.CreateTodoWithDetails(user.ID, "High priority", testutil.TodoOptions{Priority: model.PriorityHigh})

	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?priority=high", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 1)

	firstTodo := data[0].(map[string]any)
	assert.Equal(t, "High priority", firstTodo["title"])
}

// TestTodoSearch_CategoryFilter tests category filter
func TestTodoSearch_CategoryFilter(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("categoryfilter@example.com")
	category := f.CreateCategory(user.ID, "Work", "#FF0000")

	f.CreateTodoWithDetails(user.ID, "Work task", testutil.TodoOptions{CategoryID: &category.ID})
	f.CreateTodo(user.ID, "Personal task")

	// Filter by category
	rec, err := f.CallAuth(token, http.MethodGet, fmt.Sprintf("/api/v1/todos/search?category_id=%d", category.ID), "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 1)
	assert.Equal(t, "Work task", data[0].(map[string]any)["title"])
}

// TestTodoSearch_CategoryNullFilter tests filtering todos with no category
func TestTodoSearch_CategoryNullFilter(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("categorynull@example.com")
	category := f.CreateCategory(user.ID, "Work", "#FF0000")

	f.CreateTodoWithDetails(user.ID, "Work task", testutil.TodoOptions{CategoryID: &category.ID})
	f.CreateTodo(user.ID, "Uncategorized task 1")
	f.CreateTodo(user.ID, "Uncategorized task 2")

	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?category_id=-1", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 2)
}

// TestTodoSearch_TagFilterAny tests tag filter with "any" mode (OR)
func TestTodoSearch_TagFilterAny(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("tagany@example.com")
	tag1 := f.CreateTag(user.ID, "urgent", nil)
	tag2 := f.CreateTag(user.ID, "important", nil)

	todo1 := f.CreateTodo(user.ID, "Task with urgent tag")
	f.AssociateTagWithTodo(todo1.ID, tag1.ID)

	todo2 := f.CreateTodo(user.ID, "Task with important tag")
	f.AssociateTagWithTodo(todo2.ID, tag2.ID)

	f.CreateTodo(user.ID, "Task with no tags")

	// Search with tag_mode=any (default)
	rec, err := f.CallAuth(token, http.MethodGet, fmt.Sprintf("/api/v1/todos/search?tag_ids=%d,%d&tag_mode=any", tag1.ID, tag2.ID), "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 2)
}

// TestTodoSearch_TagFilterAll tests tag filter with "all" mode (AND)
func TestTodoSearch_TagFilterAll(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("tagall@example.com")
	tag1 := f.CreateTag(user.ID, "urgent", nil)
	tag2 := f.CreateTag(user.ID, "important", nil)

	// Todo with both tags
	todo1 := f.CreateTodo(user.ID, "Task with both tags")
	f.AssociateTagWithTodo(todo1.ID, tag1.ID)
	f.AssociateTagWithTodo(todo1.ID, tag2.ID)

	// Todo with only one tag
	todo2 := f.CreateTodo(user.ID, "Task with urgent tag only")
	f.AssociateTagWithTodo(todo2.ID, tag1.ID)

	// Search with tag_mode=all
	rec, err := f.CallAuth(token, http.MethodGet, fmt.Sprintf("/api/v1/todos/search?tag_ids=%d,%d&tag_mode=all", tag1.ID, tag2.ID), "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 1)
	assert.Equal(t, "Task with both tags", data[0].(map[string]any)["title"])
}

// TestTodoSearch_DateRangeFilter tests due date range filter
func TestTodoSearch_DateRangeFilter(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("daterange@example.com")

	date1 := testutil.ParseDate("2030-01-15")
	date2 := testutil.ParseDate("2030-02-15")
	date3 := testutil.ParseDate("2030-03-15")

	f.CreateTodoWithDetails(user.ID, "January task", testutil.TodoOptions{DueDate: date1})
	f.CreateTodoWithDetails(user.ID, "February task", testutil.TodoOptions{DueDate: date2})
	f.CreateTodoWithDetails(user.ID, "March task", testutil.TodoOptions{DueDate: date3})

	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?due_date_from=2030-01-01&due_date_to=2030-02-28", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 2)
}

// TestTodoSearch_Pagination tests pagination
func TestTodoSearch_Pagination(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("pagination@example.com")

	// Create 15 todos
	for i := 1; i <= 15; i++ {
		f.CreateTodo(user.ID, fmt.Sprintf("Todo %d", i))
	}

	// Get first page (5 per page)
	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?page=1&per_page=5", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	meta := response["meta"].(map[string]any)

	assert.Len(t, data, 5)
	assert.Equal(t, float64(15), meta["total"])
	assert.Equal(t, float64(1), meta["current_page"])
	assert.Equal(t, float64(3), meta["total_pages"])
	assert.Equal(t, float64(5), meta["per_page"])

	// Get second page
	rec2, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?page=2&per_page=5", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response2 := testutil.JSONResponse(t, rec2)
	data2 := response2["data"].([]any)
	assert.Len(t, data2, 5)
}

// TestTodoSearch_Sorting tests sort functionality
func TestTodoSearch_Sorting(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("sorting@example.com")
	f.CreateTodoWithDetails(user.ID, "A - Low", testutil.TodoOptions{Priority: model.PriorityLow})
	f.CreateTodoWithDetails(user.ID, "B - High", testutil.TodoOptions{Priority: model.PriorityHigh})
	f.CreateTodoWithDetails(user.ID, "C - Medium", testutil.TodoOptions{Priority: model.PriorityMedium})

	// Sort by priority descending
	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?sort_by=priority&sort_order=desc", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 3)

	// High priority should come first
	assert.Equal(t, "B - High", data[0].(map[string]any)["title"])
}

// TestTodoSearch_DueDateSortNullLast tests that NULL due dates come last
func TestTodoSearch_DueDateSortNullLast(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("nulllast@example.com")

	date1 := testutil.ParseDate("2030-01-15")
	f.CreateTodoWithDetails(user.ID, "With due date", testutil.TodoOptions{DueDate: date1})
	f.CreateTodo(user.ID, "No due date 1")
	f.CreateTodo(user.ID, "No due date 2")

	// Sort by due_date ascending
	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?sort_by=due_date&sort_order=asc", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 3)

	// Todo with due date should come first
	assert.Equal(t, "With due date", data[0].(map[string]any)["title"])
}

// TestTodoSearch_UserScope tests that users only see their own todos in search
func TestTodoSearch_UserScope(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, token1 := f.CreateUser("searchuser1@example.com")
	user2, _ := f.CreateUser("searchuser2@example.com")

	f.CreateTodo(user1.ID, "User1's Todo")
	f.CreateTodo(user2.ID, "User2's Todo")

	rec, err := f.CallAuth(token1, http.MethodGet, "/api/v1/todos/search", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 1)
	assert.Equal(t, "User1's Todo", data[0].(map[string]any)["title"])
}

// TestTodoSearch_EmptyResultSuggestions tests suggestions for empty results
func TestTodoSearch_EmptyResultSuggestions(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("suggestions@example.com")
	f.CreateTodo(user.ID, "Some task")

	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos/search?q=nonexistent", "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 0)

	suggestions := response["suggestions"]
	assert.NotNil(t, suggestions)
}

// TestTodoSearch_CombinedFilters tests multiple filters combined
func TestTodoSearch_CombinedFilters(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("combined@example.com")
	category := f.CreateCategory(user.ID, "Work", "#FF0000")
	tag := f.CreateTag(user.ID, "urgent", nil)

	// Create various todos
	todo1 := f.CreateTodoWithDetails(user.ID, "Work urgent pending", testutil.TodoOptions{
		CategoryID: &category.ID,
		Priority:   model.PriorityHigh,
		Status:     model.StatusPending,
	})
	f.AssociateTagWithTodo(todo1.ID, tag.ID)

	f.CreateTodoWithDetails(user.ID, "Work low priority", testutil.TodoOptions{
		CategoryID: &category.ID,
		Priority:   model.PriorityLow,
		Status:     model.StatusPending,
	})

	f.CreateTodoWithDetails(user.ID, "Personal high priority", testutil.TodoOptions{
		Priority: model.PriorityHigh,
		Status:   model.StatusPending,
	})

	// Combined filter: category + priority + tag
	rec, err := f.CallAuth(token, http.MethodGet, fmt.Sprintf("/api/v1/todos/search?category_id=%d&priority=high&tag_ids=%d", category.ID, tag.ID), "", f.TodoHandler.Search)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].([]any)
	assert.Len(t, data, 1)
	assert.Equal(t, "Work urgent pending", data[0].(map[string]any)["title"])
}
