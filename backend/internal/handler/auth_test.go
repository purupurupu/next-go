package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"todo-api/internal/handler"
	"todo-api/internal/middleware"
	"todo-api/internal/repository"
	"todo-api/internal/testutil"
)

// TestSignUp_Success tests successful user registration
func TestSignUp_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	e := testutil.SetupEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, testutil.TestConfig)

	body := `{"user":{"email":"test@example.com","password":"password123","password_confirmation":"password123","name":"Test User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := authHandler.SignUp(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Authorization"))
	assert.True(t, strings.HasPrefix(rec.Header().Get("Authorization"), "Bearer "))

	var response map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	status := response["status"].(map[string]any)
	assert.Equal(t, float64(http.StatusCreated), status["code"])
	assert.Equal(t, "Signed up successfully.", status["message"])

	data := response["data"].(map[string]any)
	assert.Equal(t, "test@example.com", data["email"])
	assert.Equal(t, "Test User", data["name"])
}

// TestSignUp_DuplicateEmail tests registration with duplicate email
func TestSignUp_DuplicateEmail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	e := testutil.SetupEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, testutil.TestConfig)

	// First registration
	body := `{"user":{"email":"duplicate@example.com","password":"password123","password_confirmation":"password123","name":"First User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := authHandler.SignUp(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Second registration with same email
	req2 := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	err = authHandler.SignUp(c2)

	// Should return ApiError
	require.Error(t, err)
}

// TestSignUp_ValidationError tests registration with validation errors
func TestSignUp_ValidationError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	e := testutil.SetupEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, testutil.TestConfig)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing email",
			body: `{"user":{"password":"password123","password_confirmation":"password123","name":"Test User"}}`,
		},
		{
			name: "invalid email",
			body: `{"user":{"email":"invalid-email","password":"password123","password_confirmation":"password123","name":"Test User"}}`,
		},
		{
			name: "password too short",
			body: `{"user":{"email":"test@example.com","password":"12345","password_confirmation":"12345","name":"Test User"}}`,
		},
		{
			name: "missing name",
			body: `{"user":{"email":"test@example.com","password":"password123","password_confirmation":"password123"}}`,
		},
		{
			name: "name too short",
			body: `{"user":{"email":"test@example.com","password":"password123","password_confirmation":"password123","name":"A"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := authHandler.SignUp(c)
			require.Error(t, err)
		})
	}
}

// TestSignUp_PasswordMismatch tests registration with password mismatch
func TestSignUp_PasswordMismatch(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	e := testutil.SetupEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, testutil.TestConfig)

	body := `{"user":{"email":"test@example.com","password":"password123","password_confirmation":"differentpassword","name":"Test User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := authHandler.SignUp(c)
	require.Error(t, err)
}

// TestSignIn_Success tests successful login
func TestSignIn_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	e := testutil.SetupEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, testutil.TestConfig)

	// First register a user
	signUpBody := `{"user":{"email":"login@example.com","password":"password123","password_confirmation":"password123","name":"Login User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(signUpBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := authHandler.SignUp(c)
	require.NoError(t, err)

	// Now login
	signInBody := `{"user":{"email":"login@example.com","password":"password123"}}`
	req2 := httptest.NewRequest(http.MethodPost, "/auth/sign_in", strings.NewReader(signInBody))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = authHandler.SignIn(c2)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.NotEmpty(t, rec2.Header().Get("Authorization"))
	assert.True(t, strings.HasPrefix(rec2.Header().Get("Authorization"), "Bearer "))

	var response map[string]any
	err = json.Unmarshal(rec2.Body.Bytes(), &response)
	require.NoError(t, err)

	status := response["status"].(map[string]any)
	assert.Equal(t, float64(http.StatusOK), status["code"])
	assert.Equal(t, "Logged in successfully.", status["message"])

	data := response["data"].(map[string]any)
	assert.Equal(t, "login@example.com", data["email"])
	assert.Equal(t, "Login User", data["name"])
}

// TestSignIn_InvalidCredentials tests login with wrong password
func TestSignIn_InvalidCredentials(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	e := testutil.SetupEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, testutil.TestConfig)

	// First register a user
	signUpBody := `{"user":{"email":"wrong@example.com","password":"password123","password_confirmation":"password123","name":"Wrong Password User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(signUpBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := authHandler.SignUp(c)
	require.NoError(t, err)

	// Try to login with wrong password
	signInBody := `{"user":{"email":"wrong@example.com","password":"wrongpassword"}}`
	req2 := httptest.NewRequest(http.MethodPost, "/auth/sign_in", strings.NewReader(signInBody))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = authHandler.SignIn(c2)
	require.Error(t, err)
}

// TestSignIn_NonExistentUser tests login with non-existent user
func TestSignIn_NonExistentUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	e := testutil.SetupEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, testutil.TestConfig)

	signInBody := `{"user":{"email":"nonexistent@example.com","password":"password123"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_in", strings.NewReader(signInBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := authHandler.SignIn(c)
	require.Error(t, err)
}

// TestSignOut_Success tests successful logout
func TestSignOut_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	e := testutil.SetupEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, testutil.TestConfig)

	// First register and get token
	signUpBody := `{"user":{"email":"logout@example.com","password":"password123","password_confirmation":"password123","name":"Logout User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(signUpBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := authHandler.SignUp(c)
	require.NoError(t, err)

	token := rec.Header().Get("Authorization")
	require.NotEmpty(t, token)

	// Now logout - need to set up middleware context
	req2 := httptest.NewRequest(http.MethodDelete, "/auth/sign_out", nil)
	req2.Header.Set("Authorization", token)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	// Set up JWT claims in context (simulating middleware)
	authMiddleware := middleware.JWTAuth(testutil.TestConfig, userRepo, denylistRepo)
	wrappedHandler := authMiddleware(func(c echo.Context) error {
		return authHandler.SignOut(c)
	})

	err = wrappedHandler(c2)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec2.Code)

	var response map[string]any
	err = json.Unmarshal(rec2.Body.Bytes(), &response)
	require.NoError(t, err)

	status := response["status"].(map[string]any)
	assert.Equal(t, float64(http.StatusOK), status["code"])
	assert.Equal(t, "Logged out successfully.", status["message"])
}

// TestSignOut_RevokedToken tests that revoked token cannot be used
func TestSignOut_RevokedToken(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(db)

	e := testutil.SetupEcho()
	userRepo := repository.NewUserRepository(db)
	denylistRepo := repository.NewJwtDenylistRepository(db)
	authHandler := handler.NewAuthHandler(userRepo, denylistRepo, testutil.TestConfig)

	// Register and get token
	signUpBody := `{"user":{"email":"revoked@example.com","password":"password123","password_confirmation":"password123","name":"Revoked User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(signUpBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := authHandler.SignUp(c)
	require.NoError(t, err)

	token := rec.Header().Get("Authorization")
	require.NotEmpty(t, token)

	// First logout (revoke token)
	req2 := httptest.NewRequest(http.MethodDelete, "/auth/sign_out", nil)
	req2.Header.Set("Authorization", token)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	authMiddlewareFunc := middleware.JWTAuth(testutil.TestConfig, userRepo, denylistRepo)
	wrappedHandler := authMiddlewareFunc(func(c echo.Context) error {
		return authHandler.SignOut(c)
	})
	err = wrappedHandler(c2)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Try to use the revoked token again
	req3 := httptest.NewRequest(http.MethodDelete, "/auth/sign_out", nil)
	req3.Header.Set("Authorization", token)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err = wrappedHandler(c3)
	require.Error(t, err) // Should fail because token is revoked
}
