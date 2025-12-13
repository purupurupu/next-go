package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"todo-api/internal/handler"
	"todo-api/internal/middleware"
	"todo-api/internal/model"
	"todo-api/internal/repository"
)

// TestFixture holds all dependencies needed for handler tests
type TestFixture struct {
	T            *testing.T
	DB           *gorm.DB
	Echo         *echo.Echo
	UserRepo     *repository.UserRepository
	DenylistRepo *repository.JwtDenylistRepository
	TodoRepo     *repository.TodoRepository
	AuthHandler  *handler.AuthHandler
	TodoHandler  *handler.TodoHandler
}

// SetupTestFixture creates a new TestFixture with all dependencies initialized
func SetupTestFixture(t *testing.T) *TestFixture {
	db := SetupTestDB(t)
	e := SetupEcho()

	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	todoRepo := repository.NewTodoRepository(db)

	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, TestConfig)
	todoHandler := handler.NewTodoHandler(todoRepo)

	t.Cleanup(func() {
		CleanupTestDB(db)
	})

	return &TestFixture{
		T:            t,
		DB:           db,
		Echo:         e,
		UserRepo:     userRepo,
		DenylistRepo: denylistRepo,
		TodoRepo:     todoRepo,
		AuthHandler:  authHandler,
		TodoHandler:  todoHandler,
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

// CallAuth calls a handler with JWT authentication middleware
func (f *TestFixture) CallAuth(token, method, path, body string, handlerFunc echo.HandlerFunc) (*httptest.ResponseRecorder, error) {
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

	// Extract path params
	if strings.Contains(path, "/todos/") && !strings.HasSuffix(path, "/todos/update_order") {
		parts := strings.Split(path, "/todos/")
		if len(parts) > 1 {
			c.SetParamNames("id")
			c.SetParamValues(parts[1])
		}
	}

	authMiddleware := middleware.JWTAuth(TestConfig, f.UserRepo, f.DenylistRepo)
	wrappedHandler := authMiddleware(handlerFunc)
	err := wrappedHandler(c)

	return rec, err
}

// TodoPath returns the path for a specific todo ID
func TodoPath(id int64) string {
	return fmt.Sprintf("/api/v1/todos/%d", id)
}
