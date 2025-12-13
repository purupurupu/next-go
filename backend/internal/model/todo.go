package model

import (
	"time"

	"gorm.io/gorm"
)

// Priority represents the priority level of a todo
type Priority int

const (
	PriorityLow    Priority = 0
	PriorityMedium Priority = 1
	PriorityHigh   Priority = 2
)

// Status represents the status of a todo
type Status int

const (
	StatusPending    Status = 0
	StatusInProgress Status = 1
	StatusCompleted  Status = 2
)

// Todo represents a task in the system
type Todo struct {
	ID          int64      `gorm:"primaryKey" json:"id"`
	UserID      int64      `gorm:"not null;index" json:"user_id"`
	CategoryID  *int64     `gorm:"index" json:"category_id"`
	Title       string     `gorm:"not null;size:255" json:"title"`
	Description *string    `gorm:"type:text" json:"description"`
	Completed   bool       `gorm:"default:false" json:"completed"`
	Position    *int       `gorm:"index" json:"position"`
	Priority    Priority   `gorm:"not null;default:1;index" json:"priority"`
	Status      Status     `gorm:"not null;default:0;index" json:"status"`
	DueDate     *time.Time `gorm:"type:date;index" json:"due_date"`
	CreatedAt   time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"index" json:"updated_at"`

	// Relations (will be preloaded when needed)
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags     []Tag     `gorm:"many2many:todo_tags;" json:"tags,omitempty"`
}

// TableName returns the table name for the Todo model
func (Todo) TableName() string {
	return "todos"
}

// BeforeCreate sets the position for new todos
func (t *Todo) BeforeCreate(tx *gorm.DB) error {
	if t.Position == nil {
		// Get the max position for the user's todos
		var maxPosition int
		tx.Model(&Todo{}).
			Where("user_id = ?", t.UserID).
			Select("COALESCE(MAX(position), 0)").
			Scan(&maxPosition)

		newPosition := maxPosition + 1
		t.Position = &newPosition
	}
	return nil
}

// IsValidPriority checks if the priority value is valid
func IsValidPriority(p Priority) bool {
	return p >= PriorityLow && p <= PriorityHigh
}

// IsValidStatus checks if the status value is valid
func IsValidStatus(s Status) bool {
	return s >= StatusPending && s <= StatusCompleted
}

// PriorityString returns the string representation of priority
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	default:
		return "unknown"
	}
}

// StatusString returns the string representation of status
func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusInProgress:
		return "in_progress"
	case StatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

// Category model placeholder (will be implemented in Phase 3)
type Category struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	UserID     int64     `gorm:"not null;index" json:"user_id"`
	Name       string    `gorm:"not null;size:50" json:"name"`
	Color      string    `gorm:"not null;size:7;default:'#6B7280'" json:"color"`
	TodosCount int       `gorm:"not null;default:0" json:"todos_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName returns the table name for the Category model
func (Category) TableName() string {
	return "categories"
}

// Tag model placeholder (will be implemented in Phase 3)
type Tag struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	Name      string    `gorm:"not null;size:30" json:"name"`
	Color     *string   `gorm:"size:7;default:'#6B7280'" json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the Tag model
func (Tag) TableName() string {
	return "tags"
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
