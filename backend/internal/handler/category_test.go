package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"todo-api/internal/model"
	"todo-api/internal/testutil"
)

// TestCategoryList_Success tests successful category list retrieval
func TestCategoryList_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("catlist@example.com")
	f.CreateCategory(user.ID, "Work", "#FF0000")
	f.CreateCategory(user.ID, "Personal", "#00FF00")

	rec, err := f.CallAuthCategory(token, http.MethodGet, "/api/v1/categories", "", f.CategoryHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	categories := testutil.ExtractCategories(response)
	assert.Len(t, categories, 2)

	// Should be sorted by name (Personal < Work)
	first := testutil.CategoryAt(categories, 0)
	assert.Equal(t, "Personal", first["name"])
}

// TestCategoryList_Empty tests empty category list
func TestCategoryList_Empty(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("catempty@example.com")

	rec, err := f.CallAuthCategory(token, http.MethodGet, "/api/v1/categories", "", f.CategoryHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	categories := testutil.ExtractCategories(response)
	assert.Len(t, categories, 0)
}

// TestCategoryList_UserScope tests that users only see their own categories
func TestCategoryList_UserScope(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, token1 := f.CreateUser("catuser1@example.com")
	user2, _ := f.CreateUser("catuser2@example.com")

	f.CreateCategory(user1.ID, "User1 Category", "#FF0000")
	f.CreateCategory(user2.ID, "User2 Category", "#00FF00")

	rec, err := f.CallAuthCategory(token1, http.MethodGet, "/api/v1/categories", "", f.CategoryHandler.List)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	categories := testutil.ExtractCategories(response)
	assert.Len(t, categories, 1)
	assert.Equal(t, "User1 Category", testutil.CategoryAt(categories, 0)["name"])
}

// TestCategoryCreate_Success tests successful category creation
func TestCategoryCreate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("catcreate@example.com")

	body := `{"category":{"name":"New Category","color":"#FF5500"}}`
	rec, err := f.CallAuthCategory(token, http.MethodPost, "/api/v1/categories", body, f.CategoryHandler.Create)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	response := testutil.JSONResponse(t, rec)
	category := testutil.ExtractCategoryFromData(response)
	assert.Equal(t, "New Category", category["name"])
	assert.Equal(t, "#FF5500", category["color"])
	assert.Equal(t, float64(0), category["todo_count"])
}

// TestCategoryCreate_DuplicateName tests category creation with duplicate name
func TestCategoryCreate_DuplicateName(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("catdup@example.com")
	f.CreateCategory(user.ID, "Work", "#FF0000")

	// Try to create with same name (case-insensitive)
	body := `{"category":{"name":"WORK","color":"#00FF00"}}`
	_, err := f.CallAuthCategory(token, http.MethodPost, "/api/v1/categories", body, f.CategoryHandler.Create)
	require.Error(t, err)
}

// TestCategoryCreate_ValidationError tests category creation with validation errors
func TestCategoryCreate_ValidationError(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("catvalid@example.com")

	tests := []struct {
		name string
		body string
	}{
		{name: "missing name", body: `{"category":{"color":"#FF0000"}}`},
		{name: "empty name", body: `{"category":{"name":"","color":"#FF0000"}}`},
		{name: "blank name", body: `{"category":{"name":"   ","color":"#FF0000"}}`},
		{name: "invalid color", body: `{"category":{"name":"Test","color":"invalid"}}`},
		{name: "missing color", body: `{"category":{"name":"Test"}}`},
		{name: "name too long", body: `{"category":{"name":"` + strings.Repeat("a", 51) + `","color":"#FF0000"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.CallAuthCategory(token, http.MethodPost, "/api/v1/categories", tt.body, f.CategoryHandler.Create)
			require.Error(t, err)
		})
	}
}

// TestCategoryShow_Success tests successful category retrieval by ID
func TestCategoryShow_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("catshow@example.com")
	category := f.CreateCategory(user.ID, "Show Me", "#FF0000")

	rec, err := f.CallAuthCategory(token, http.MethodGet, testutil.CategoryPath(category.ID), "", f.CategoryHandler.Show)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	catResp := testutil.ExtractCategory(response)
	assert.Equal(t, "Show Me", catResp["name"])
}

// TestCategoryShow_NotFound tests category not found error
func TestCategoryShow_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("catnotfound@example.com")

	_, err := f.CallAuthCategory(token, http.MethodGet, "/api/v1/categories/99999", "", f.CategoryHandler.Show)
	require.Error(t, err)
}

// TestCategoryShow_OtherUser tests accessing other user's category
func TestCategoryShow_OtherUser(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("catowner@example.com")
	_, token2 := f.CreateUser("catother@example.com")

	category := f.CreateCategory(user1.ID, "Private Category", "#FF0000")

	_, err := f.CallAuthCategory(token2, http.MethodGet, testutil.CategoryPath(category.ID), "", f.CategoryHandler.Show)
	require.Error(t, err)
}

// TestCategoryUpdate_Success tests successful category update
func TestCategoryUpdate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("catupdate@example.com")
	category := f.CreateCategory(user.ID, "Original", "#FF0000")

	body := `{"category":{"name":"Updated","color":"#00FF00"}}`
	rec, err := f.CallAuthCategory(token, http.MethodPut, testutil.CategoryPath(category.ID), body, f.CategoryHandler.Update)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	catResp := testutil.ExtractCategory(response)
	assert.Equal(t, "Updated", catResp["name"])
	assert.Equal(t, "#00FF00", catResp["color"])
}

// TestCategoryUpdate_PartialUpdate tests partial category update
func TestCategoryUpdate_PartialUpdate(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("catpartial@example.com")
	category := f.CreateCategory(user.ID, "Original", "#FF0000")

	body := `{"category":{"color":"#00FF00"}}`
	rec, err := f.CallAuthCategory(token, http.MethodPut, testutil.CategoryPath(category.ID), body, f.CategoryHandler.Update)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	catResp := testutil.ExtractCategory(response)
	assert.Equal(t, "Original", catResp["name"]) // Name unchanged
	assert.Equal(t, "#00FF00", catResp["color"])
}

// TestCategoryUpdate_DuplicateName tests category update with duplicate name
func TestCategoryUpdate_DuplicateName(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("catdupupdate@example.com")
	f.CreateCategory(user.ID, "Existing", "#FF0000")
	category := f.CreateCategory(user.ID, "ToUpdate", "#00FF00")

	body := `{"category":{"name":"Existing"}}`
	_, err := f.CallAuthCategory(token, http.MethodPut, testutil.CategoryPath(category.ID), body, f.CategoryHandler.Update)
	require.Error(t, err)
}

// TestCategoryDelete_Success tests successful category deletion
func TestCategoryDelete_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("catdelete@example.com")
	category := f.CreateCategory(user.ID, "Delete Me", "#FF0000")

	rec, err := f.CallAuthCategory(token, http.MethodDelete, testutil.CategoryPath(category.ID), "", f.CategoryHandler.Delete)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify category is deleted
	var count int64
	f.DB.Model(&model.Category{}).Where("id = ?", category.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestCategoryDelete_NullifiesRelatedTodos tests that deleting category nullifies related todos
func TestCategoryDelete_NullifiesRelatedTodos(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("catnullify@example.com")
	category := f.CreateCategory(user.ID, "To Delete", "#FF0000")
	todo := f.CreateTodoWithCategory(user.ID, "Related Todo", category.ID)

	rec, err := f.CallAuthCategory(token, http.MethodDelete, testutil.CategoryPath(category.ID), "", f.CategoryHandler.Delete)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify todo's category_id is null
	var updatedTodo model.Todo
	f.DB.First(&updatedTodo, todo.ID)
	assert.Nil(t, updatedTodo.CategoryID)
}

// TestCategoryDelete_OtherUser tests deleting other user's category
func TestCategoryDelete_OtherUser(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("catdelowner@example.com")
	_, token2 := f.CreateUser("catdelother@example.com")

	category := f.CreateCategory(user1.ID, "Not Yours", "#FF0000")

	_, err := f.CallAuthCategory(token2, http.MethodDelete, testutil.CategoryPath(category.ID), "", f.CategoryHandler.Delete)
	require.Error(t, err)

	// Verify category still exists
	var count int64
	f.DB.Model(&model.Category{}).Where("id = ?", category.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}
