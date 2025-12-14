package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"todo-api/internal/handler"
	"todo-api/internal/middleware"
	"todo-api/internal/model"
	"todo-api/internal/repository"
	"todo-api/internal/service"
)

// TestFixture holds all dependencies needed for handler tests
type TestFixture struct {
	T               *testing.T
	DB              *gorm.DB
	Echo            *echo.Echo
	UserRepo        *repository.UserRepository
	DenylistRepo    *repository.JwtDenylistRepository
	TodoRepo        *repository.TodoRepository
	CategoryRepo    *repository.CategoryRepository
	TagRepo         *repository.TagRepository
	AuthHandler     *handler.AuthHandler
	TodoHandler     *handler.TodoHandler
	CategoryHandler *handler.CategoryHandler
	TagHandler      *handler.TagHandler
}

// SetupTestFixture creates a new TestFixture with all dependencies initialized
func SetupTestFixture(t *testing.T) *TestFixture {
	db := SetupTestDB(t)
	e := SetupEcho()

	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)

	// Initialize services
	todoService := service.NewTodoService(todoRepo, categoryRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, TestConfig)
	todoHandler := handler.NewTodoHandler(todoService, todoRepo)
	categoryHandler := handler.NewCategoryHandler(categoryRepo)
	tagHandler := handler.NewTagHandler(tagRepo)

	t.Cleanup(func() {
		CleanupTestDB(db)
	})

	return &TestFixture{
		T:               t,
		DB:              db,
		Echo:            e,
		UserRepo:        userRepo,
		DenylistRepo:    denylistRepo,
		TodoRepo:        todoRepo,
		CategoryRepo:    categoryRepo,
		TagRepo:         tagRepo,
		AuthHandler:     authHandler,
		TodoHandler:     todoHandler,
		CategoryHandler: categoryHandler,
		TagHandler:      tagHandler,
	}
}

// CreateUser creates a test user and returns the user and JWT token
func (f *TestFixture) CreateUser(email string) (*model.User, string) {
	body := fmt.Sprintf(`{"user":{"email":"%s","password":"password123","password_confirmation":"password123","name":"Test User"}}`, email)
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	err := f.AuthHandler.SignUp(c)
	require.NoError(f.T, err)
	require.Equal(f.T, http.StatusCreated, rec.Code)

	token := rec.Header().Get("Authorization")
	require.NotEmpty(f.T, token)

	user, err := f.UserRepo.FindByEmail(email)
	require.NoError(f.T, err)

	return user, token
}

// CreateTodo creates a test todo for a user
func (f *TestFixture) CreateTodo(userID int64, title string) *model.Todo {
	todo := &model.Todo{
		UserID: userID,
		Title:  title,
	}
	require.NoError(f.T, f.DB.Create(todo).Error)
	return todo
}

// CreateTodoWithPosition creates a test todo with a specific position
func (f *TestFixture) CreateTodoWithPosition(userID int64, title string, position int) *model.Todo {
	todo := &model.Todo{
		UserID:   userID,
		Title:    title,
		Position: &position,
	}
	require.NoError(f.T, f.DB.Create(todo).Error)
	return todo
}

// CallAuthGeneric calls a handler with JWT authentication middleware
// This is the unified method for all resource types (todos, categories, tags)
func (f *TestFixture) CallAuthGeneric(token, method, path, body string, handlerFunc echo.HandlerFunc) (*httptest.ResponseRecorder, error) {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	// Extract path params for any resource type
	// Pattern: /api/v1/{resource}/{id} or /{resource}/{id}
	resources := []string{"todos", "categories", "tags"}
	for _, resource := range resources {
		pattern := "/" + resource + "/"
		if strings.Contains(path, pattern) {
			// Skip special endpoints like /todos/update_order
			if resource == "todos" && strings.HasSuffix(path, "/update_order") {
				continue
			}
			parts := strings.Split(path, pattern)
			if len(parts) > 1 && parts[1] != "" {
				c.SetParamNames("id")
				c.SetParamValues(parts[1])
				break
			}
		}
	}

	authMiddleware := middleware.JWTAuth(TestConfig, f.UserRepo, f.DenylistRepo)
	wrappedHandler := authMiddleware(handlerFunc)
	err := wrappedHandler(c)

	return rec, err
}

// CallAuth calls a handler with JWT authentication middleware (alias for CallAuthGeneric)
func (f *TestFixture) CallAuth(token, method, path, body string, handlerFunc echo.HandlerFunc) (*httptest.ResponseRecorder, error) {
	return f.CallAuthGeneric(token, method, path, body, handlerFunc)
}

// ResourcePath returns the path for a specific resource by ID
func ResourcePath(resource string, id int64) string {
	return fmt.Sprintf("/api/v1/%s/%d", resource, id)
}

// TodoPath returns the path for a specific todo ID
func TodoPath(id int64) string {
	return ResourcePath("todos", id)
}

// CategoryPath returns the path for a specific category ID
func CategoryPath(id int64) string {
	return ResourcePath("categories", id)
}

// TagPath returns the path for a specific tag ID
func TagPath(id int64) string {
	return ResourcePath("tags", id)
}

// CreateCategory creates a test category for a user
func (f *TestFixture) CreateCategory(userID int64, name, color string) *model.Category {
	category := &model.Category{
		UserID: userID,
		Name:   name,
		Color:  color,
	}
	require.NoError(f.T, f.DB.Create(category).Error)
	return category
}

// CreateTag creates a test tag for a user
func (f *TestFixture) CreateTag(userID int64, name string, color *string) *model.Tag {
	tag := &model.Tag{
		UserID: userID,
		Name:   strings.ToLower(name),
		Color:  color,
	}
	require.NoError(f.T, f.DB.Create(tag).Error)
	return tag
}

// CreateTodoWithCategory creates a test todo with a category
func (f *TestFixture) CreateTodoWithCategory(userID int64, title string, categoryID int64) *model.Todo {
	todo := &model.Todo{
		UserID:     userID,
		Title:      title,
		CategoryID: &categoryID,
	}
	require.NoError(f.T, f.DB.Create(todo).Error)
	return todo
}

// CallAuthCategory calls a category handler with authentication (alias for CallAuthGeneric)
func (f *TestFixture) CallAuthCategory(token, method, path, body string, handlerFunc echo.HandlerFunc) (*httptest.ResponseRecorder, error) {
	return f.CallAuthGeneric(token, method, path, body, handlerFunc)
}

// CallAuthTag calls a tag handler with authentication (alias for CallAuthGeneric)
func (f *TestFixture) CallAuthTag(token, method, path, body string, handlerFunc echo.HandlerFunc) (*httptest.ResponseRecorder, error) {
	return f.CallAuthGeneric(token, method, path, body, handlerFunc)
}

// TodoOptions contains options for creating a todo with details
type TodoOptions struct {
	Description *string
	CategoryID  *int64
	Priority    model.Priority
	Status      model.Status
	DueDate     *time.Time
	TagIDs      []int64
}

// CreateTodoWithDetails creates a test todo with full details
func (f *TestFixture) CreateTodoWithDetails(userID int64, title string, opts TodoOptions) *model.Todo {
	todo := &model.Todo{
		UserID:      userID,
		Title:       title,
		Description: opts.Description,
		CategoryID:  opts.CategoryID,
		Priority:    opts.Priority,
		Status:      opts.Status,
		DueDate:     opts.DueDate,
	}
	require.NoError(f.T, f.DB.Create(todo).Error)

	// Associate tags if provided
	if len(opts.TagIDs) > 0 {
		for _, tagID := range opts.TagIDs {
			todoTag := &model.TodoTag{TodoID: todo.ID, TagID: tagID}
			require.NoError(f.T, f.DB.Create(todoTag).Error)
		}
	}

	return todo
}

// AssociateTagWithTodo associates a tag with a todo
func (f *TestFixture) AssociateTagWithTodo(todoID, tagID int64) {
	todoTag := &model.TodoTag{TodoID: todoID, TagID: tagID}
	require.NoError(f.T, f.DB.Create(todoTag).Error)
}

// ParseDate parses a date string in YYYY-MM-DD format
func ParseDate(dateStr string) *time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}
	return &t
}
