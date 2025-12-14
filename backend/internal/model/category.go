package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Category represents a category for organizing todos
type Category struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	UserID     int64     `gorm:"not null;index:idx_category_user_name,unique" json:"user_id"`
	Name       string    `gorm:"not null;size:50;index:idx_category_user_name,unique" json:"name"`
	Color      string    `gorm:"not null;size:7;default:'#6B7280'" json:"color"`
	TodosCount int       `gorm:"column:todos_count;not null;default:0" json:"todo_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	Todos []Todo `gorm:"foreignKey:CategoryID" json:"todos,omitempty"`
}

// TableName returns the table name for the Category model
func (Category) TableName() string {
	return "categories"
}

// BeforeSave normalizes category name to lowercase before saving
func (c *Category) BeforeSave(tx *gorm.DB) error {
	c.Name = strings.ToLower(strings.TrimSpace(c.Name))
	return nil
}
