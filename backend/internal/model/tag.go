package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Tag represents a tag for labeling todos
type Tag struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"not null;index:idx_tag_user_name,unique" json:"user_id"`
	Name      string    `gorm:"not null;size:30;index:idx_tag_user_name,unique" json:"name"`
	Color     *string   `gorm:"size:7;default:'#6B7280'" json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Todos []Todo `gorm:"many2many:todo_tags;" json:"todos,omitempty"`
}

// TableName returns the table name for the Tag model
func (Tag) TableName() string {
	return "tags"
}

// BeforeSave normalizes tag name to lowercase before saving
func (t *Tag) BeforeSave(tx *gorm.DB) error {
	t.Name = strings.ToLower(strings.TrimSpace(t.Name))
	return nil
}

// TodoTag represents the many-to-many relationship between todos and tags
type TodoTag struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	TodoID    int64     `gorm:"not null;index;uniqueIndex:idx_todo_tag" json:"todo_id"`
	TagID     int64     `gorm:"not null;index;uniqueIndex:idx_todo_tag" json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the TodoTag model
func (TodoTag) TableName() string {
	return "todo_tags"
}
