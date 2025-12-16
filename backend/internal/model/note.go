package model

import (
	"time"

	"gorm.io/gorm"
)

// Note represents a markdown note
type Note struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	UserID       int64      `gorm:"not null;index" json:"user_id"`
	Title        *string    `gorm:"type:varchar(150)" json:"title"`
	BodyMD       *string    `gorm:"column:body_md;type:text" json:"body_md"`
	BodyPlain    *string    `gorm:"column:body_plain;type:text" json:"-"`
	Pinned       bool       `gorm:"default:false" json:"pinned"`
	ArchivedAt   *time.Time `gorm:"index" json:"archived_at"`
	TrashedAt    *time.Time `gorm:"index" json:"trashed_at"`
	LastEditedAt time.Time  `gorm:"not null" json:"last_edited_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relations
	User      *User          `gorm:"foreignKey:UserID" json:"-"`
	Revisions []NoteRevision `gorm:"foreignKey:NoteID" json:"-"`
}

// TableName returns the table name for the Note model
func (Note) TableName() string {
	return "notes"
}

// IsArchived returns true if the note is archived
func (n *Note) IsArchived() bool {
	return n.ArchivedAt != nil
}

// IsTrashed returns true if the note is trashed
func (n *Note) IsTrashed() bool {
	return n.TrashedAt != nil
}

// IsOwnedBy checks if the note is owned by the given user
func (n *Note) IsOwnedBy(userID int64) bool {
	return n.UserID == userID
}

// BeforeCreate sets default values before creating
func (n *Note) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if n.LastEditedAt.IsZero() {
		n.LastEditedAt = now
	}
	return nil
}
