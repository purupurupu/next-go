package repository

import (
	"time"

	"todo-api/internal/model"
)

// UserRepositoryInterface defines the contract for user repository operations
type UserRepositoryInterface interface {
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
	FindByID(id int64) (*model.User, error)
	ExistsByEmail(email string) (bool, error)
}

// TodoRepositoryInterface defines the contract for todo repository operations
type TodoRepositoryInterface interface {
	FindAllByUserID(userID int64) ([]model.Todo, error)
	FindAllByUserIDWithRelations(userID int64) ([]model.Todo, error)
	FindByID(id, userID int64) (*model.Todo, error)
	FindByIDWithRelations(id, userID int64) (*model.Todo, error)
	Create(todo *model.Todo) error
	Update(todo *model.Todo) error
	UpdateFields(id, userID int64, updates map[string]any) error
	Delete(id, userID int64) error
	UpdateOrder(userID int64, updates []OrderUpdate) error
	Count(userID int64) (int64, error)
	ExistsByID(id, userID int64) (bool, error)
	ValidateCategoryOwnership(categoryID, userID int64) (bool, error)
}

// JwtDenylistRepositoryInterface defines the contract for JWT denylist operations
type JwtDenylistRepositoryInterface interface {
	Add(jti string, exp time.Time) error
	Exists(jti string) (bool, error)
	CleanupExpired() error
}

// Ensure concrete types implement interfaces
var (
	_ UserRepositoryInterface       = (*UserRepository)(nil)
	_ TodoRepositoryInterface       = (*TodoRepository)(nil)
	_ JwtDenylistRepositoryInterface = (*JwtDenylistRepository)(nil)
)
