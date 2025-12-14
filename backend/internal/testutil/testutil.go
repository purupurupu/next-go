package testutil

import (
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"todo-api/internal/config"
	"todo-api/internal/model"
	"todo-api/internal/validator"
)

// TestConfig provides default test configuration
var TestConfig = &config.Config{
	JWTSecret:          "test-secret-key-for-testing-purposes",
	JWTExpirationHours: 24,
}

// GetTestDSN returns the database DSN for testing
// Uses TEST_DATABASE_URL environment variable if set, otherwise falls back to default
func GetTestDSN() string {
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	// Default: connect to db_test service in docker compose
	return "host=db_test user=postgres password=password dbname=todo_next_test port=5432 sslmode=disable"
}

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *gorm.DB {
	dsn := GetTestDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Database not available, skipping test: %v", err)
	}

	// Auto migrate models
	err = db.AutoMigrate(
		&model.User{},
		&model.JwtDenylist{},
		&model.Category{},
		&model.Tag{},
		&model.Todo{},
		&model.TodoTag{},
		&model.Comment{},
		&model.TodoHistory{},
	)
	require.NoError(t, err)

	return db
}

// CleanupTestDB cleans up test data
func CleanupTestDB(db *gorm.DB) {
	// Delete in order respecting foreign key constraints
	db.Exec("DELETE FROM comments")
	db.Exec("DELETE FROM todo_histories")
	db.Exec("DELETE FROM todo_tags")
	db.Exec("DELETE FROM todos")
	db.Exec("DELETE FROM tags")
	db.Exec("DELETE FROM categories")
	db.Exec("DELETE FROM jwt_denylists")
	db.Exec("DELETE FROM users")
}

// SetupEcho creates an Echo instance for testing
func SetupEcho() *echo.Echo {
	e := echo.New()
	validator.SetupValidator(e)
	return e
}
