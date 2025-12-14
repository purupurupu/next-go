package model

import (
	"time"

	"gorm.io/gorm"
)

// CommentableType constants
const (
	CommentableTypeTodo = "Todo"
)

// EditWindowMinutes defines how long a comment can be edited after creation
const EditWindowMinutes = 15

// Comment represents a comment on a resource (polymorphic)
type Comment struct {
	ID              int64          `gorm:"primaryKey" json:"id"`
	Content         string         `gorm:"type:text;not null" json:"content"`
	UserID          int64          `gorm:"not null;index" json:"user_id"`
	CommentableType string         `gorm:"not null;size:50;index:idx_commentable" json:"commentable_type"`
	CommentableID   int64          `gorm:"not null;index:idx_commentable" json:"commentable_id"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName returns the table name for the Comment model
func (Comment) TableName() string {
	return "comments"
}

// IsEditable checks if the comment can be edited (within 15 minutes and not deleted)
func (c *Comment) IsEditable() bool {
	if c.DeletedAt.Valid {
		return false
	}
	return time.Since(c.CreatedAt) < EditWindowMinutes*time.Minute
}

// IsOwnedBy checks if the comment is owned by the given user
func (c *Comment) IsOwnedBy(userID int64) bool {
	return c.UserID == userID
}

// IsValidCommentableType validates the commentable type
func IsValidCommentableType(t string) bool {
	return t == CommentableTypeTodo
}
