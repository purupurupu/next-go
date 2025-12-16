package model

import (
	"encoding/json"
	"time"
)

// HistoryAction represents the type of action performed on a todo
type HistoryAction string

const (
	ActionCreated         HistoryAction = "created"
	ActionUpdated         HistoryAction = "updated"
	ActionDeleted         HistoryAction = "deleted"
	ActionStatusChanged   HistoryAction = "status_changed"
	ActionPriorityChanged HistoryAction = "priority_changed"
)

// IsValidHistoryAction checks if the action is valid
func IsValidHistoryAction(action HistoryAction) bool {
	switch action {
	case ActionCreated, ActionUpdated, ActionDeleted, ActionStatusChanged, ActionPriorityChanged:
		return true
	default:
		return false
	}
}

// TodoHistory represents a history record for a todo
type TodoHistory struct {
	ID        int64           `gorm:"primaryKey" json:"id"`
	TodoID    int64           `gorm:"not null;index" json:"todo_id"`
	UserID    int64           `gorm:"not null;index" json:"user_id"`
	Action    HistoryAction   `gorm:"type:varchar(50);not null;index" json:"action"`
	Changes   json.RawMessage `gorm:"type:jsonb" json:"changes"`
	CreatedAt time.Time       `gorm:"not null" json:"created_at"`

	// Relations (will be preloaded when needed)
	Todo *Todo `gorm:"foreignKey:TodoID;constraint:OnDelete:CASCADE" json:"-"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName returns the table name for the TodoHistory model
func (TodoHistory) TableName() string {
	return "todo_histories"
}
