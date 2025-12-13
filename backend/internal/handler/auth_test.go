package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"todo-api/internal/middleware"
	"todo-api/internal/testutil"
)

// TestSignUp_Success tests successful user registration
func TestSignUp_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	body := `{"user":{"email":"test@example.com","password":"password123","password_confirmation":"password123","name":"Test User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	err := f.AuthHandler.SignUp(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Authorization"))
	assert.True(t, strings.HasPrefix(rec.Header().Get("Authorization"), "Bearer "))

	response := testutil.JSONResponse(t, rec)
	assert.Equal(t, http.StatusCreated, testutil.ExtractStatusCode(response))
	assert.Equal(t, "Signed up successfully.", testutil.ExtractMessage(response))

	data := testutil.ExtractData(response)
	assert.Equal(t, "test@example.com", data["email"])
	assert.Equal(t, "Test User", data["name"])
}

// TestSignUp_DuplicateEmail tests registration with duplicate email
func TestSignUp_DuplicateEmail(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	// First registration
	body := `{"user":{"email":"duplicate@example.com","password":"password123","password_confirmation":"password123","name":"First User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)
	err := f.AuthHandler.SignUp(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Second registration with same email
	req2 := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := f.Echo.NewContext(req2, rec2)
	err = f.AuthHandler.SignUp(c2)

	// Should return ApiError
	require.Error(t, err)
}

// TestSignUp_ValidationError tests registration with validation errors
func TestSignUp_ValidationError(t *testing.T) {
	f := testutil.SetupTestFixture(t)

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
			c := f.Echo.NewContext(req, rec)

			err := f.AuthHandler.SignUp(c)
			require.Error(t, err)
		})
	}
}

// TestSignUp_PasswordMismatch tests registration with password mismatch
func TestSignUp_PasswordMismatch(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	body := `{"user":{"email":"test@example.com","password":"password123","password_confirmation":"differentpassword","name":"Test User"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_up", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	err := f.AuthHandler.SignUp(c)
	require.Error(t, err)
}

// TestSignIn_Success tests successful login
func TestSignIn_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	// First register a user
	f.CreateUser("login@example.com")

	// Now login
	signInBody := `{"user":{"email":"login@example.com","password":"password123"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_in", strings.NewReader(signInBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	err := f.AuthHandler.SignIn(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Authorization"))
	assert.True(t, strings.HasPrefix(rec.Header().Get("Authorization"), "Bearer "))

	response := testutil.JSONResponse(t, rec)
	assert.Equal(t, http.StatusOK, testutil.ExtractStatusCode(response))
	assert.Equal(t, "Logged in successfully.", testutil.ExtractMessage(response))

	data := testutil.ExtractData(response)
	assert.Equal(t, "login@example.com", data["email"])
	assert.Equal(t, "Test User", data["name"])
}

// TestSignIn_InvalidCredentials tests login with wrong password
func TestSignIn_InvalidCredentials(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	// First register a user
	f.CreateUser("wrong@example.com")

	// Try to login with wrong password
	signInBody := `{"user":{"email":"wrong@example.com","password":"wrongpassword"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_in", strings.NewReader(signInBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	err := f.AuthHandler.SignIn(c)
	require.Error(t, err)
}

// TestSignIn_NonExistentUser tests login with non-existent user
func TestSignIn_NonExistentUser(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	signInBody := `{"user":{"email":"nonexistent@example.com","password":"password123"}}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign_in", strings.NewReader(signInBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	err := f.AuthHandler.SignIn(c)
	require.Error(t, err)
}

// TestSignOut_Success tests successful logout
func TestSignOut_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	// First register and get token
	_, token := f.CreateUser("logout@example.com")

	// Now logout - need to set up middleware context
	req := httptest.NewRequest(http.MethodDelete, "/auth/sign_out", nil)
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	// Set up JWT claims in context (simulating middleware)
	authMiddleware := middleware.JWTAuth(testutil.TestConfig, f.UserRepo, f.DenylistRepo)
	wrappedHandler := authMiddleware(func(c echo.Context) error {
		return f.AuthHandler.SignOut(c)
	})

	err := wrappedHandler(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	assert.Equal(t, http.StatusOK, testutil.ExtractStatusCode(response))
	assert.Equal(t, "Logged out successfully.", testutil.ExtractMessage(response))
}

// TestSignOut_RevokedToken tests that revoked token cannot be used
func TestSignOut_RevokedToken(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	// Register and get token
	_, token := f.CreateUser("revoked@example.com")

	// First logout (revoke token)
	req := httptest.NewRequest(http.MethodDelete, "/auth/sign_out", nil)
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()
	c := f.Echo.NewContext(req, rec)

	authMiddlewareFunc := middleware.JWTAuth(testutil.TestConfig, f.UserRepo, f.DenylistRepo)
	wrappedHandler := authMiddlewareFunc(func(c echo.Context) error {
		return f.AuthHandler.SignOut(c)
	})
	err := wrappedHandler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Try to use the revoked token again
	req2 := httptest.NewRequest(http.MethodDelete, "/auth/sign_out", nil)
	req2.Header.Set("Authorization", token)
	rec2 := httptest.NewRecorder()
	c2 := f.Echo.NewContext(req2, rec2)

	err = wrappedHandler(c2)
	require.Error(t, err) // Should fail because token is revoked
}
