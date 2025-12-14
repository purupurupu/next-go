package handler_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"todo-api/internal/testutil"
)

// =============================================================================
// TodoHistory List Tests
// =============================================================================

func TestTodoHistory_ListAfterCreate(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("historylistcreate@example.com")

	// Create todo via API (this will trigger history recording)
	body := `{"todo":{"title":"New Todo","priority":2}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", body, f.TodoHandler.Create)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	todoID := int64(todo["id"].(float64))

	// Get histories
	rec2, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(todoID), "", f.HistoryHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec2.Code)

	historyResp := testutil.JSONResponse(t, rec2)
	histories := historyResp["histories"].([]interface{})
	assert.Len(t, histories, 1)

	firstHistory := histories[0].(map[string]interface{})
	assert.Equal(t, "created", firstHistory["action"])
	assert.NotEmpty(t, firstHistory["human_readable_change"])
	assert.Equal(t, "Todoが作成されました", firstHistory["human_readable_change"])
}

func TestTodoHistory_ListAfterUpdate(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("historylistupdate@example.com")

	// Create todo via API
	createBody := `{"todo":{"title":"Original Title"}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", createBody, f.TodoHandler.Create)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	todoID := int64(todo["id"].(float64))

	// Update todo
	updateBody := `{"todo":{"title":"Updated Title"}}`
	rec2, err := f.CallAuth(token, http.MethodPatch, testutil.TodoPath(todoID), updateBody, f.TodoHandler.Update)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec2.Code)

	// Get histories
	rec3, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(todoID), "", f.HistoryHandler.List)
	require.NoError(t, err)

	historyResp := testutil.JSONResponse(t, rec3)
	histories := historyResp["histories"].([]interface{})
	assert.GreaterOrEqual(t, len(histories), 2)

	// Most recent should be the update (histories are ordered DESC)
	latestHistory := histories[0].(map[string]interface{})
	assert.Equal(t, "updated", latestHistory["action"])
}

func TestTodoHistory_StatusChanged(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("historystatuschanged@example.com")

	// Create todo via API
	createBody := `{"todo":{"title":"Status Test"}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", createBody, f.TodoHandler.Create)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	todoID := int64(todo["id"].(float64))

	// Change status only
	updateBody := `{"todo":{"status":1}}`
	_, err = f.CallAuth(token, http.MethodPatch, testutil.TodoPath(todoID), updateBody, f.TodoHandler.Update)
	require.NoError(t, err)

	// Get histories
	rec2, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(todoID), "", f.HistoryHandler.List)
	require.NoError(t, err)

	historyResp := testutil.JSONResponse(t, rec2)
	histories := historyResp["histories"].([]interface{})

	latestHistory := histories[0].(map[string]interface{})
	assert.Equal(t, "status_changed", latestHistory["action"])
	assert.Contains(t, latestHistory["human_readable_change"].(string), "ステータス")
}

func TestTodoHistory_PriorityChanged(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("historyprioritychanged@example.com")

	// Create todo via API
	createBody := `{"todo":{"title":"Priority Test"}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", createBody, f.TodoHandler.Create)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	todoID := int64(todo["id"].(float64))

	// Change priority only
	updateBody := `{"todo":{"priority":2}}`
	_, err = f.CallAuth(token, http.MethodPatch, testutil.TodoPath(todoID), updateBody, f.TodoHandler.Update)
	require.NoError(t, err)

	// Get histories
	rec2, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(todoID), "", f.HistoryHandler.List)
	require.NoError(t, err)

	historyResp := testutil.JSONResponse(t, rec2)
	histories := historyResp["histories"].([]interface{})

	latestHistory := histories[0].(map[string]interface{})
	assert.Equal(t, "priority_changed", latestHistory["action"])
	assert.Contains(t, latestHistory["human_readable_change"].(string), "優先度")
}

func TestTodoHistory_Pagination(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("historypagination@example.com")

	// Create todo via API
	createBody := `{"todo":{"title":"Pagination Test"}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", createBody, f.TodoHandler.Create)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	todoID := int64(todo["id"].(float64))

	// Create multiple updates to generate history
	for i := 0; i < 5; i++ {
		updateBody := `{"todo":{"title":"Title ` + string(rune('A'+i)) + `"}}`
		_, err = f.CallAuth(token, http.MethodPatch, testutil.TodoPath(todoID), updateBody, f.TodoHandler.Update)
		require.NoError(t, err)
	}

	// Get first page with small per_page
	rec2, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(todoID)+"?page=1&per_page=2", "", f.HistoryHandler.List)
	require.NoError(t, err)

	historyResp := testutil.JSONResponse(t, rec2)
	histories := historyResp["histories"].([]interface{})
	meta := historyResp["meta"].(map[string]interface{})

	assert.Len(t, histories, 2)
	assert.GreaterOrEqual(t, int(meta["total"].(float64)), 5)
	assert.Equal(t, float64(1), meta["current_page"])
}

func TestTodoHistory_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("historynotfound@example.com")

	_, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(99999), "", f.HistoryHandler.List)
	require.Error(t, err)
}

func TestTodoHistory_UserScope(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("historyuserscope1@example.com")
	_, token2 := f.CreateUser("historyuserscope2@example.com")

	todo := f.CreateTodo(user1.ID, "User1's Todo")

	_, err := f.CallAuth(token2, http.MethodGet, testutil.TodoHistoriesPath(todo.ID), "", f.HistoryHandler.List)
	require.Error(t, err)
}

func TestTodoHistory_HumanReadableChange(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("historyhr@example.com")

	// Create todo via API
	createBody := `{"todo":{"title":"HR Test"}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", createBody, f.TodoHandler.Create)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	todoID := int64(todo["id"].(float64))

	// Update title
	updateBody := `{"todo":{"title":"New Title"}}`
	_, err = f.CallAuth(token, http.MethodPatch, testutil.TodoPath(todoID), updateBody, f.TodoHandler.Update)
	require.NoError(t, err)

	rec2, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(todoID), "", f.HistoryHandler.List)
	require.NoError(t, err)

	historyResp := testutil.JSONResponse(t, rec2)
	histories := historyResp["histories"].([]interface{})
	latestHistory := histories[0].(map[string]interface{})

	hrChange := latestHistory["human_readable_change"].(string)
	assert.Contains(t, hrChange, "タイトル")
}

func TestTodoHistory_UserInfoIncluded(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("historyuserinfo@example.com")

	// Create todo via API
	body := `{"todo":{"title":"User Info Test"}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", body, f.TodoHandler.Create)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	todoID := int64(todo["id"].(float64))

	rec2, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(todoID), "", f.HistoryHandler.List)
	require.NoError(t, err)

	historyResp := testutil.JSONResponse(t, rec2)
	histories := historyResp["histories"].([]interface{})
	firstHistory := histories[0].(map[string]interface{})

	user := firstHistory["user"].(map[string]interface{})
	assert.NotNil(t, user["id"])
	assert.NotNil(t, user["email"])
}

func TestTodoHistory_NoChangeNoHistory(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("historynochange@example.com")

	// Create todo via API
	createBody := `{"todo":{"title":"No Change Test"}}`
	rec, err := f.CallAuth(token, http.MethodPost, "/api/v1/todos", createBody, f.TodoHandler.Create)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	todoID := int64(todo["id"].(float64))

	// Get initial history count
	rec2, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(todoID), "", f.HistoryHandler.List)
	require.NoError(t, err)
	historyResp1 := testutil.JSONResponse(t, rec2)
	initialCount := len(historyResp1["histories"].([]interface{}))

	// Update with same title (no actual change)
	updateBody := `{"todo":{"title":"No Change Test"}}`
	_, err = f.CallAuth(token, http.MethodPatch, testutil.TodoPath(todoID), updateBody, f.TodoHandler.Update)
	require.NoError(t, err)

	// Get history count after update
	rec3, err := f.CallAuth(token, http.MethodGet, testutil.TodoHistoriesPath(todoID), "", f.HistoryHandler.List)
	require.NoError(t, err)
	historyResp2 := testutil.JSONResponse(t, rec3)
	finalCount := len(historyResp2["histories"].([]interface{}))

	// History count should not increase when there's no actual change
	assert.Equal(t, initialCount, finalCount)
}
