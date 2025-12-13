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

	response := testutil.JSONResponse(t, rec)
	todos := testutil.ExtractTodos(response)
	assert.Len(t, todos, 2)
}

// TestTodoList_Empty tests empty todo list
func TestTodoList_Empty(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("emptylist@example.com")

	rec, err := f.CallAuth(token, http.MethodGet, "/api/v1/todos", "", f.TodoHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	todos := testutil.ExtractTodos(response)
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

	response := testutil.JSONResponse(t, rec)
	todos := testutil.ExtractTodos(response)
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
