package repository

import (
	"todo-api/internal/model"

	"gorm.io/gorm"
)

// NoteRevisionRepository handles database operations for note revisions
type NoteRevisionRepository struct {
	db *gorm.DB
}

// NewNoteRevisionRepository creates a new NoteRevisionRepository
func NewNoteRevisionRepository(db *gorm.DB) *NoteRevisionRepository {
	return &NoteRevisionRepository{db: db}
}

// Create creates a new revision
func (r *NoteRevisionRepository) Create(revision *model.NoteRevision) error {
	return r.db.Create(revision).Error
}

// FindByNoteID retrieves all revisions for a note with pagination
func (r *NoteRevisionRepository) FindByNoteID(noteID int64, page, perPage int) ([]model.NoteRevision, int64, error) {
	var revisions []model.NoteRevision
	var total int64

	// Count total
	if err := r.db.Model(&model.NoteRevision{}).Where("note_id = ?", noteID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results (newest first)
	offset := (page - 1) * perPage
	result := r.db.
		Where("note_id = ?", noteID).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&revisions)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return revisions, total, nil
}

// FindByID retrieves a specific revision
func (r *NoteRevisionRepository) FindByID(id, noteID int64) (*model.NoteRevision, error) {
	var revision model.NoteRevision
	result := r.db.
		Where("id = ? AND note_id = ?", id, noteID).
		First(&revision)
	if result.Error != nil {
		return nil, result.Error
	}
	return &revision, nil
}

// CountByNoteID counts revisions for a note
func (r *NoteRevisionRepository) CountByNoteID(noteID int64) (int64, error) {
	var count int64
	result := r.db.Model(&model.NoteRevision{}).
		Where("note_id = ?", noteID).
		Count(&count)
	return count, result.Error
}

// DeleteOldestByNoteID deletes the oldest revisions exceeding the limit
func (r *NoteRevisionRepository) DeleteOldestByNoteID(noteID int64, keepCount int) error {
	// Get IDs of revisions to keep (newest ones)
	var keepIDs []int64
	err := r.db.Model(&model.NoteRevision{}).
		Select("id").
		Where("note_id = ?", noteID).
		Order("created_at DESC").
		Limit(keepCount).
		Pluck("id", &keepIDs).Error
	if err != nil {
		return err
	}

	if len(keepIDs) == 0 {
		return nil
	}

	// Delete revisions not in the keep list
	return r.db.
		Where("note_id = ? AND id NOT IN ?", noteID, keepIDs).
		Delete(&model.NoteRevision{}).Error
}
