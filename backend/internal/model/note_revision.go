package model

import "time"

// MaxRevisionsPerNote defines the maximum number of revisions to keep per note
const MaxRevisionsPerNote = 50

// NoteRevision represents a historical version of a note's content
type NoteRevision struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	NoteID    int64     `gorm:"not null;index" json:"note_id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	Title     *string   `gorm:"type:varchar(150)" json:"title"`
	BodyMD    *string   `gorm:"column:body_md;type:text" json:"body_md"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Note *Note `gorm:"foreignKey:NoteID;constraint:OnDelete:CASCADE" json:"-"`
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName returns the table name for the NoteRevision model
func (NoteRevision) TableName() string {
	return "note_revisions"
}
