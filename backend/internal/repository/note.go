package repository

import (
	"todo-api/internal/model"

	"gorm.io/gorm"
)

// NoteRepository handles database operations for notes
type NoteRepository struct {
	db *gorm.DB
}

// NewNoteRepository creates a new NoteRepository
func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// NoteSearchInput represents search parameters for notes
type NoteSearchInput struct {
	UserID   int64
	Query    string
	Archived *bool
	Trashed  *bool
	Pinned   *bool
	Page     int
	PerPage  int
}

// FindByID retrieves a note by ID for a specific user (excludes trashed)
func (r *NoteRepository) FindByID(id, userID int64) (*model.Note, error) {
	var note model.Note
	result := r.db.
		Where("id = ? AND user_id = ? AND trashed_at IS NULL", id, userID).
		First(&note)
	if result.Error != nil {
		return nil, result.Error
	}
	return &note, nil
}

// FindByIDIncludingTrashed retrieves a note by ID including trashed ones
func (r *NoteRepository) FindByIDIncludingTrashed(id, userID int64) (*model.Note, error) {
	var note model.Note
	result := r.db.
		Where("id = ? AND user_id = ?", id, userID).
		First(&note)
	if result.Error != nil {
		return nil, result.Error
	}
	return &note, nil
}

// Search searches notes with filters and pagination
func (r *NoteRepository) Search(input NoteSearchInput) ([]model.Note, int64, error) {
	var notes []model.Note
	var total int64

	query := r.db.Model(&model.Note{}).Where("user_id = ?", input.UserID)

	// Text search (title and body_plain)
	if input.Query != "" {
		searchPattern := "%" + input.Query + "%"
		query = query.Where("(title ILIKE ? OR body_plain ILIKE ?)", searchPattern, searchPattern)
	}

	// Archived filter
	if input.Archived != nil {
		if *input.Archived {
			query = query.Where("archived_at IS NOT NULL")
		} else {
			query = query.Where("archived_at IS NULL")
		}
	}

	// Trashed filter
	if input.Trashed != nil {
		if *input.Trashed {
			query = query.Where("trashed_at IS NOT NULL")
		} else {
			query = query.Where("trashed_at IS NULL")
		}
	} else {
		// Default: exclude trashed
		query = query.Where("trashed_at IS NULL")
	}

	// Pinned filter
	if input.Pinned != nil {
		query = query.Where("pinned = ?", *input.Pinned)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (input.Page - 1) * input.PerPage
	result := query.
		Order("pinned DESC, last_edited_at DESC").
		Offset(offset).
		Limit(input.PerPage).
		Find(&notes)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return notes, total, nil
}

// Create creates a new note
func (r *NoteRepository) Create(note *model.Note) error {
	return r.db.Create(note).Error
}

// Update updates an existing note
func (r *NoteRepository) Update(note *model.Note) error {
	return r.db.Save(note).Error
}

// SoftDelete sets trashed_at timestamp
func (r *NoteRepository) SoftDelete(id, userID int64) error {
	return r.db.Model(&model.Note{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("trashed_at", gorm.Expr("NOW()")).Error
}

// HardDelete permanently deletes a note and its revisions
func (r *NoteRepository) HardDelete(id, userID int64) error {
	// Delete revisions first (foreign key constraint)
	if err := r.db.Where("note_id = ?", id).Delete(&model.NoteRevision{}).Error; err != nil {
		return err
	}
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Note{}).Error
}

// ExistsByID checks if a note exists for a user
func (r *NoteRepository) ExistsByID(id, userID int64) (bool, error) {
	var count int64
	result := r.db.Model(&model.Note{}).
		Where("id = ? AND user_id = ?", id, userID).
		Count(&count)
	return count > 0, result.Error
}
