package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"todo-api/internal/config"
	"todo-api/internal/handler"
	"todo-api/internal/middleware"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/internal/validator"
)

// Test configuration for todo tests
var todoTestConfig = &config.Config{
	JWTSecret:          "test-secret-key-for-testing-purposes",
	JWTExpirationHours: 24,
}

// setupTodoTestDB creates a test database connection for todo tests
func setupTodoTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=postgres password=password dbname=todo_next_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skip("Database not available, skipping test")
	}

	// Auto migrate models
	err = db.AutoMigrate(
		&model.User{},
		&model.JwtDenylist{},
		&model.Category{},
		&model.Tag{},
		&model.Todo{},
		&model.TodoTag{},
	)
	require.NoError(t, err)

	return db
}

// cleanupTodoTestDB cleans up test data for todo tests
func cleanupTodoTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM todo_tags")
	db.Exec("DELETE FROM todos")
	db.Exec("DELETE FROM tags")
	db.Exec("DELETE FROM categories")
	db.Exec("DELETE FROM jwt_denylists")
	db.Exec("DELETE FROM users")
}

// createTestUser creates a test user and returns the user and JWT token
func createTestUser(t *testing.T, db *gorm.DB, e *echo.Echo, email string) (*model.User, string) {
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, todoTestConfig)

	body := fmt.Sprintf(`{"user":{"email":"%s","password":"password123","password_confirmation":"password123","name":"Test User"}}`, email)
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := authHandler.SignUp(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, rec.Code)

	token := rec.Header().Get("Authorization")
	require.NotEmpty(t, token)

	user, err := userRepo.FindByEmail(email)
	require.NoError(t, err)

	return user, token
}

// setupTodoEcho creates an Echo instance for todo testing
func setupTodoEcho() *echo.Echo {
	e := echo.New()
	validator.SetupValidator(e)
	return e
}

// callWithAuth calls a handler with authentication middleware
func callWithAuth(e *echo.Echo, token string, method, path, body string, handlerFunc echo.HandlerFunc, cfg *config.Config, userRepo *repository.UserRepository, denylistRepo *repository.JwtDenylistRepository) (*httptest.ResponseRecorder, error) {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Extract path params
	if strings.Contains(path, "/todos/") && !strings.HasSuffix(path, "/todos/update_order") {
		parts := strings.Split(path, "/todos/")
		if len(parts) > 1 {
			c.SetParamNames("id")
			c.SetParamValues(parts[1])
		}
	}

	authMiddleware := middleware.JWTAuth(cfg, userRepo, denylistRepo)
	wrappedHandler := authMiddleware(handlerFunc)
	err := wrappedHandler(c)

	return rec, err
}

// TestTodoList_Success tests successful todo list retrieval
func TestTodoList_Success(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	user, token := createTestUser(t, db, e, "todolist@example.com")

	// Create some todos directly in DB
	todo1 := &model.Todo{UserID: user.ID, Title: "Todo 1"}
	todo2 := &model.Todo{UserID: user.ID, Title: "Todo 2"}
	require.NoError(t, db.Create(todo1).Error)
	require.NoError(t, db.Create(todo2).Error)

	rec, err := callWithAuth(e, token, http.MethodGet, "/api/v1/todos", "", todoHandler.List, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	todos := response["todos"].([]interface{})
	assert.Len(t, todos, 2)
}

// TestTodoList_Empty tests empty todo list
func TestTodoList_Empty(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	_, token := createTestUser(t, db, e, "emptylist@example.com")

	rec, err := callWithAuth(e, token, http.MethodGet, "/api/v1/todos", "", todoHandler.List, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	todos := response["todos"].([]interface{})
	assert.Len(t, todos, 0)
}

// TestTodoList_UserScope tests that users only see their own todos
func TestTodoList_UserScope(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create two test users
	user1, token1 := createTestUser(t, db, e, "user1@example.com")
	user2, _ := createTestUser(t, db, e, "user2@example.com")

	// Create todos for each user
	todo1 := &model.Todo{UserID: user1.ID, Title: "User1 Todo"}
	todo2 := &model.Todo{UserID: user2.ID, Title: "User2 Todo"}
	require.NoError(t, db.Create(todo1).Error)
	require.NoError(t, db.Create(todo2).Error)

	// User1 should only see their own todo
	rec, err := callWithAuth(e, token1, http.MethodGet, "/api/v1/todos", "", todoHandler.List, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	todos := response["todos"].([]interface{})
	assert.Len(t, todos, 1)

	firstTodo := todos[0].(map[string]interface{})
	assert.Equal(t, "User1 Todo", firstTodo["title"])
}

// TestTodoCreate_Success tests successful todo creation
func TestTodoCreate_Success(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	_, token := createTestUser(t, db, e, "create@example.com")

	body := `{"todo":{"title":"New Todo","description":"A test todo","priority":2,"status":0}}`
	rec, err := callWithAuth(e, token, http.MethodPost, "/api/v1/todos", body, todoHandler.Create, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	assert.Equal(t, "New Todo", todo["title"])
	assert.Equal(t, "A test todo", todo["description"])
	assert.Equal(t, float64(2), todo["priority"])
	assert.Equal(t, float64(0), todo["status"])
	assert.NotNil(t, todo["position"])
}

// TestTodoCreate_ValidationError tests todo creation with validation errors
func TestTodoCreate_ValidationError(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	_, token := createTestUser(t, db, e, "validation@example.com")

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing title",
			body: `{"todo":{"description":"No title"}}`,
		},
		{
			name: "empty title",
			body: `{"todo":{"title":""}}`,
		},
		{
			name: "invalid priority",
			body: `{"todo":{"title":"Test","priority":5}}`,
		},
		{
			name: "invalid status",
			body: `{"todo":{"title":"Test","status":10}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := callWithAuth(e, token, http.MethodPost, "/api/v1/todos", tt.body, todoHandler.Create, todoTestConfig, userRepo, denylistRepo)
			require.Error(t, err)
		})
	}
}

// TestTodoShow_Success tests successful todo retrieval by ID
func TestTodoShow_Success(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	user, token := createTestUser(t, db, e, "show@example.com")

	// Create a todo
	todo := &model.Todo{UserID: user.ID, Title: "Show Me"}
	require.NoError(t, db.Create(todo).Error)

	path := fmt.Sprintf("/api/v1/todos/%d", todo.ID)
	rec, err := callWithAuth(e, token, http.MethodGet, path, "", todoHandler.Show, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	todoResp := response["todo"].(map[string]interface{})
	assert.Equal(t, "Show Me", todoResp["title"])
}

// TestTodoShow_NotFound tests todo not found error
func TestTodoShow_NotFound(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	_, token := createTestUser(t, db, e, "notfound@example.com")

	_, err := callWithAuth(e, token, http.MethodGet, "/api/v1/todos/99999", "", todoHandler.Show, todoTestConfig, userRepo, denylistRepo)
	require.Error(t, err)
}

// TestTodoShow_OtherUserTodo tests that users cannot see other users' todos
func TestTodoShow_OtherUserTodo(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create two test users
	user1, _ := createTestUser(t, db, e, "owner@example.com")
	_, token2 := createTestUser(t, db, e, "other@example.com")

	// Create a todo for user1
	todo := &model.Todo{UserID: user1.ID, Title: "User1's Todo"}
	require.NoError(t, db.Create(todo).Error)

	// User2 tries to access user1's todo
	path := fmt.Sprintf("/api/v1/todos/%d", todo.ID)
	_, err := callWithAuth(e, token2, http.MethodGet, path, "", todoHandler.Show, todoTestConfig, userRepo, denylistRepo)
	require.Error(t, err)
}

// TestTodoUpdate_Success tests successful todo update
func TestTodoUpdate_Success(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	user, token := createTestUser(t, db, e, "update@example.com")

	// Create a todo
	todo := &model.Todo{UserID: user.ID, Title: "Original Title"}
	require.NoError(t, db.Create(todo).Error)

	body := `{"todo":{"title":"Updated Title","priority":2,"completed":true}}`
	path := fmt.Sprintf("/api/v1/todos/%d", todo.ID)
	rec, err := callWithAuth(e, token, http.MethodPatch, path, body, todoHandler.Update, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	todoResp := response["todo"].(map[string]interface{})
	assert.Equal(t, "Updated Title", todoResp["title"])
	assert.Equal(t, float64(2), todoResp["priority"])
	assert.Equal(t, true, todoResp["completed"])
	assert.Equal(t, float64(2), todoResp["status"]) // Should be completed
}

// TestTodoUpdate_PartialUpdate tests partial todo update
func TestTodoUpdate_PartialUpdate(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	user, token := createTestUser(t, db, e, "partial@example.com")

	// Create a todo with specific values
	desc := "Original description"
	todo := &model.Todo{
		UserID:      user.ID,
		Title:       "Original Title",
		Description: &desc,
		Priority:    model.PriorityLow,
	}
	require.NoError(t, db.Create(todo).Error)

	// Update only the title
	body := `{"todo":{"title":"New Title"}}`
	path := fmt.Sprintf("/api/v1/todos/%d", todo.ID)
	rec, err := callWithAuth(e, token, http.MethodPatch, path, body, todoHandler.Update, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	todoResp := response["todo"].(map[string]interface{})
	assert.Equal(t, "New Title", todoResp["title"])
	assert.Equal(t, "Original description", todoResp["description"]) // Should remain unchanged
	assert.Equal(t, float64(0), todoResp["priority"])                // Should remain unchanged
}

// TestTodoUpdate_OtherUserTodo tests that users cannot update other users' todos
func TestTodoUpdate_OtherUserTodo(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create two test users
	user1, _ := createTestUser(t, db, e, "updateowner@example.com")
	_, token2 := createTestUser(t, db, e, "updateother@example.com")

	// Create a todo for user1
	todo := &model.Todo{UserID: user1.ID, Title: "User1's Todo"}
	require.NoError(t, db.Create(todo).Error)

	// User2 tries to update user1's todo
	body := `{"todo":{"title":"Hacked!"}}`
	path := fmt.Sprintf("/api/v1/todos/%d", todo.ID)
	_, err := callWithAuth(e, token2, http.MethodPatch, path, body, todoHandler.Update, todoTestConfig, userRepo, denylistRepo)
	require.Error(t, err)
}

// TestTodoDelete_Success tests successful todo deletion
func TestTodoDelete_Success(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	user, token := createTestUser(t, db, e, "delete@example.com")

	// Create a todo
	todo := &model.Todo{UserID: user.ID, Title: "Delete Me"}
	require.NoError(t, db.Create(todo).Error)

	path := fmt.Sprintf("/api/v1/todos/%d", todo.ID)
	rec, err := callWithAuth(e, token, http.MethodDelete, path, "", todoHandler.Delete, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify todo is deleted
	var count int64
	db.Model(&model.Todo{}).Where("id = ?", todo.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestTodoDelete_NotFound tests deleting non-existent todo
func TestTodoDelete_NotFound(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	_, token := createTestUser(t, db, e, "deletenotfound@example.com")

	_, err := callWithAuth(e, token, http.MethodDelete, "/api/v1/todos/99999", "", todoHandler.Delete, todoTestConfig, userRepo, denylistRepo)
	require.Error(t, err)
}

// TestTodoDelete_OtherUserTodo tests that users cannot delete other users' todos
func TestTodoDelete_OtherUserTodo(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create two test users
	user1, _ := createTestUser(t, db, e, "deleteowner@example.com")
	_, token2 := createTestUser(t, db, e, "deleteother@example.com")

	// Create a todo for user1
	todo := &model.Todo{UserID: user1.ID, Title: "User1's Todo"}
	require.NoError(t, db.Create(todo).Error)

	// User2 tries to delete user1's todo
	path := fmt.Sprintf("/api/v1/todos/%d", todo.ID)
	_, err := callWithAuth(e, token2, http.MethodDelete, path, "", todoHandler.Delete, todoTestConfig, userRepo, denylistRepo)
	require.Error(t, err)

	// Verify todo still exists
	var count int64
	db.Model(&model.Todo{}).Where("id = ?", todo.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

// TestTodoUpdateOrder_Success tests successful order update
func TestTodoUpdateOrder_Success(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	user, token := createTestUser(t, db, e, "order@example.com")

	// Create todos
	pos1, pos2, pos3 := 1, 2, 3
	todo1 := &model.Todo{UserID: user.ID, Title: "Todo 1", Position: &pos1}
	todo2 := &model.Todo{UserID: user.ID, Title: "Todo 2", Position: &pos2}
	todo3 := &model.Todo{UserID: user.ID, Title: "Todo 3", Position: &pos3}
	require.NoError(t, db.Create(todo1).Error)
	require.NoError(t, db.Create(todo2).Error)
	require.NoError(t, db.Create(todo3).Error)

	// Update order: swap positions
	body := fmt.Sprintf(`{"todos":[{"id":%d,"position":3},{"id":%d,"position":1},{"id":%d,"position":2}]}`,
		todo1.ID, todo2.ID, todo3.ID)
	rec, err := callWithAuth(e, token, http.MethodPatch, "/api/v1/todos/update_order", body, todoHandler.UpdateOrder, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify positions
	var updated1, updated2, updated3 model.Todo
	db.First(&updated1, todo1.ID)
	db.First(&updated2, todo2.ID)
	db.First(&updated3, todo3.ID)

	assert.Equal(t, 3, *updated1.Position)
	assert.Equal(t, 1, *updated2.Position)
	assert.Equal(t, 2, *updated3.Position)
}

// TestTodoCreate_WithDueDate tests todo creation with due date
func TestTodoCreate_WithDueDate(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	_, token := createTestUser(t, db, e, "duedate@example.com")

	// Create with future due date
	body := `{"todo":{"title":"Due Date Todo","due_date":"2030-12-31"}}`
	rec, err := callWithAuth(e, token, http.MethodPost, "/api/v1/todos", body, todoHandler.Create, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	todo := data["todo"].(map[string]interface{})
	assert.Equal(t, "2030-12-31", todo["due_date"])
}

// TestTodoCreate_PastDueDate tests that past due date is rejected
func TestTodoCreate_PastDueDate(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	_, token := createTestUser(t, db, e, "pastdue@example.com")

	// Try to create with past due date
	body := `{"todo":{"title":"Past Due Todo","due_date":"2020-01-01"}}`
	_, err := callWithAuth(e, token, http.MethodPost, "/api/v1/todos", body, todoHandler.Create, todoTestConfig, userRepo, denylistRepo)
	require.Error(t, err)
}

// TestTodo_AutoPosition tests auto position assignment
func TestTodo_AutoPosition(t *testing.T) {
	db := setupTodoTestDB(t)
	defer cleanupTodoTestDB(db)

	e := setupTodoEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handler.NewTodoHandler(todoRepo)

	// Create test user
	_, token := createTestUser(t, db, e, "autopos@example.com")

	// Create first todo
	body1 := `{"todo":{"title":"First Todo"}}`
	rec1, err := callWithAuth(e, token, http.MethodPost, "/api/v1/todos", body1, todoHandler.Create, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	var response1 map[string]interface{}
	json.Unmarshal(rec1.Body.Bytes(), &response1)
	data1 := response1["data"].(map[string]interface{})
	todo1 := data1["todo"].(map[string]interface{})
	pos1 := todo1["position"].(float64)

	// Create second todo
	body2 := `{"todo":{"title":"Second Todo"}}`
	rec2, err := callWithAuth(e, token, http.MethodPost, "/api/v1/todos", body2, todoHandler.Create, todoTestConfig, userRepo, denylistRepo)
	require.NoError(t, err)

	var response2 map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &response2)
	data2 := response2["data"].(map[string]interface{})
	todo2 := data2["todo"].(map[string]interface{})
	pos2 := todo2["position"].(float64)

	// Second todo should have higher position
	assert.Greater(t, pos2, pos1)
}
