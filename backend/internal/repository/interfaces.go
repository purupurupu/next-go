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

// CategoryRepositoryInterface defines the contract for category repository operations
type CategoryRepositoryInterface interface {
	FindAllByUserID(userID int64) ([]model.Category, error)
	FindByID(id, userID int64) (*model.Category, error)
	ExistsByName(name string, userID int64, excludeID *int64) (bool, error)
	Create(category *model.Category) error
	Update(category *model.Category) error
	Delete(id, userID int64) error
	IncrementTodosCount(categoryID int64) error
	DecrementTodosCount(categoryID int64) error
	RecalculateTodosCount(categoryID int64) error
}

// TagRepositoryInterface defines the contract for tag repository operations
type TagRepositoryInterface interface {
	FindAllByUserID(userID int64) ([]model.Tag, error)
	FindByID(id, userID int64) (*model.Tag, error)
	ExistsByName(name string, userID int64, excludeID *int64) (bool, error)
	Create(tag *model.Tag) error
	Update(tag *model.Tag) error
	Delete(id, userID int64) error
	FindByIDs(ids []int64, userID int64) ([]model.Tag, error)
	ValidateTagOwnership(tagIDs []int64, userID int64) (bool, error)
}

// CommentRepositoryInterface defines the contract for comment repository operations
type CommentRepositoryInterface interface {
	FindAllByCommentable(commentableType string, commentableID int64) ([]model.Comment, error)
	FindByID(id int64) (*model.Comment, error)
	FindByIDWithoutDeleted(id int64) (*model.Comment, error)
	Create(comment *model.Comment) error
	Update(comment *model.Comment) error
	SoftDelete(id int64) error
	ExistsByID(id int64) (bool, error)
}

// TodoHistoryRepositoryInterface defines the contract for todo history repository operations
type TodoHistoryRepositoryInterface interface {
	Create(history *model.TodoHistory) error
	FindByTodoID(todoID int64, page, perPage int) ([]model.TodoHistory, int64, error)
	FindByTodoIDWithUser(todoID int64, page, perPage int) ([]model.TodoHistory, int64, error)
}

// Ensure concrete types implement interfaces
var (
	_ UserRepositoryInterface        = (*UserRepository)(nil)
	_ TodoRepositoryInterface        = (*TodoRepository)(nil)
	_ JwtDenylistRepositoryInterface = (*JwtDenylistRepository)(nil)
	_ CategoryRepositoryInterface    = (*CategoryRepository)(nil)
	_ TagRepositoryInterface         = (*TagRepository)(nil)
	_ CommentRepositoryInterface     = (*CommentRepository)(nil)
	_ TodoHistoryRepositoryInterface = (*TodoHistoryRepository)(nil)
)
