package service

import (
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"todo-api/internal/errors"
	"todo-api/internal/model"
	"todo-api/internal/repository"
)

// NoteService handles note business logic
type NoteService struct {
	noteRepo     *repository.NoteRepository
	revisionRepo *repository.NoteRevisionRepository
}

// NewNoteService creates a new NoteService
func NewNoteService(
	noteRepo *repository.NoteRepository,
	revisionRepo *repository.NoteRevisionRepository,
) *NoteService {
	return &NoteService{
		noteRepo:     noteRepo,
		revisionRepo: revisionRepo,
	}
}

// CreateNoteInput represents input for creating a note
type CreateNoteInput struct {
	UserID int64
	Title  *string
	BodyMD *string
	Pinned *bool
}

// UpdateNoteInput represents input for updating a note
type UpdateNoteInput struct {
	Title    *string
	BodyMD   *string
	Pinned   *bool
	Archived *bool
	Trashed  *bool
}

// Create creates a new note with initial revision
func (s *NoteService) Create(input CreateNoteInput) (*model.Note, error) {
	// Generate body_plain from body_md
	var bodyPlain *string
	if input.BodyMD != nil && *input.BodyMD != "" {
		plain := s.stripMarkdown(*input.BodyMD)
		bodyPlain = &plain
	}

	pinned := false
	if input.Pinned != nil {
		pinned = *input.Pinned
	}

	note := &model.Note{
		UserID:    input.UserID,
		Title:     input.Title,
		BodyMD:    input.BodyMD,
		BodyPlain: bodyPlain,
		Pinned:    pinned,
	}

	if err := s.noteRepo.Create(note); err != nil {
		return nil, errors.InternalErrorWithLog(err, "NoteService.Create: failed to create note")
	}

	// Create initial revision
	if err := s.createRevision(note, input.UserID); err != nil {
		log.Error().Err(err).Msg("NoteService.Create: failed to create initial revision")
	}

	return note, nil
}

// Update updates a note, creating revision if body_md changed
func (s *NoteService) Update(noteID, userID int64, input UpdateNoteInput) (*model.Note, error) {
	// Get existing note (including trashed for recovery)
	note, err := s.noteRepo.FindByIDIncludingTrashed(noteID, userID)
	if err != nil {
		return nil, err
	}

	// Track if content changed for revision
	bodyChanged := false
	contentEdited := false // for last_edited_at update

	// Apply title update
	if input.Title != nil {
		if note.Title == nil || *note.Title != *input.Title {
			contentEdited = true
		}
		note.Title = input.Title
	}

	// Apply body_md update
	if input.BodyMD != nil {
		oldBody := ""
		if note.BodyMD != nil {
			oldBody = *note.BodyMD
		}
		if oldBody != *input.BodyMD {
			bodyChanged = true
			contentEdited = true
		}
		note.BodyMD = input.BodyMD

		// Update body_plain
		if *input.BodyMD != "" {
			plain := s.stripMarkdown(*input.BodyMD)
			note.BodyPlain = &plain
		} else {
			note.BodyPlain = nil
		}
	}

	// Apply pinned update
	if input.Pinned != nil {
		note.Pinned = *input.Pinned
	}

	// Apply archived update
	if input.Archived != nil {
		if *input.Archived {
			now := time.Now()
			note.ArchivedAt = &now
		} else {
			note.ArchivedAt = nil
		}
	}

	// Apply trashed update
	if input.Trashed != nil {
		if *input.Trashed {
			now := time.Now()
			note.TrashedAt = &now
		} else {
			note.TrashedAt = nil
		}
	}

	// Update last_edited_at if content was edited
	if contentEdited {
		note.LastEditedAt = time.Now()
	}

	// Save changes
	if err := s.noteRepo.Update(note); err != nil {
		return nil, errors.InternalErrorWithLog(err, "NoteService.Update: failed to update note")
	}

	// Create revision only if body_md changed
	if bodyChanged {
		if err := s.createRevision(note, userID); err != nil {
			log.Error().Err(err).Msg("NoteService.Update: failed to create revision")
		}

		// Enforce revision limit
		if err := s.enforceRevisionLimit(noteID); err != nil {
			log.Error().Err(err).Msg("NoteService.Update: failed to enforce revision limit")
		}
	}

	return note, nil
}

// Delete handles soft/hard delete
func (s *NoteService) Delete(noteID, userID int64, force bool) error {
	// Check note exists
	exists, err := s.noteRepo.ExistsByID(noteID, userID)
	if err != nil {
		return errors.InternalErrorWithLog(err, "NoteService.Delete: failed to check note existence")
	}
	if !exists {
		return errors.NotFound("Note", noteID)
	}

	if force {
		// Hard delete
		return s.noteRepo.HardDelete(noteID, userID)
	}

	// Soft delete
	return s.noteRepo.SoftDelete(noteID, userID)
}

// RestoreRevision restores a note from a specific revision
func (s *NoteService) RestoreRevision(noteID, revisionID, userID int64) (*model.Note, error) {
	// Get the note
	note, err := s.noteRepo.FindByIDIncludingTrashed(noteID, userID)
	if err != nil {
		return nil, err
	}

	// Get the revision
	revision, err := s.revisionRepo.FindByID(revisionID, noteID)
	if err != nil {
		return nil, errors.NotFound("NoteRevision", revisionID)
	}

	// Create revision of current state before restoring
	if err := s.createRevision(note, userID); err != nil {
		log.Error().Err(err).Msg("NoteService.RestoreRevision: failed to create pre-restore revision")
	}

	// Restore note from revision
	note.Title = revision.Title
	note.BodyMD = revision.BodyMD
	if revision.BodyMD != nil && *revision.BodyMD != "" {
		plain := s.stripMarkdown(*revision.BodyMD)
		note.BodyPlain = &plain
	} else {
		note.BodyPlain = nil
	}
	note.LastEditedAt = time.Now()

	// Save
	if err := s.noteRepo.Update(note); err != nil {
		return nil, errors.InternalErrorWithLog(err, "NoteService.RestoreRevision: failed to update note")
	}

	// Enforce revision limit
	if err := s.enforceRevisionLimit(noteID); err != nil {
		log.Error().Err(err).Msg("NoteService.RestoreRevision: failed to enforce revision limit")
	}

	return note, nil
}

// createRevision creates a revision and enforces 50 limit
func (s *NoteService) createRevision(note *model.Note, userID int64) error {
	revision := &model.NoteRevision{
		NoteID: note.ID,
		UserID: userID,
		Title:  note.Title,
		BodyMD: note.BodyMD,
	}
	return s.revisionRepo.Create(revision)
}

// enforceRevisionLimit deletes old revisions if exceeding 50
func (s *NoteService) enforceRevisionLimit(noteID int64) error {
	count, err := s.revisionRepo.CountByNoteID(noteID)
	if err != nil {
		return err
	}

	if count > model.MaxRevisionsPerNote {
		return s.revisionRepo.DeleteOldestByNoteID(noteID, model.MaxRevisionsPerNote)
	}

	return nil
}

// stripMarkdown removes markdown syntax to create plain text for search
func (s *NoteService) stripMarkdown(md string) string {
	if md == "" {
		return ""
	}

	result := md

	// Remove code blocks first (```code```)
	result = regexp.MustCompile("```[\\s\\S]*?```").ReplaceAllString(result, "")

	// Remove inline code (`code`)
	result = regexp.MustCompile("`([^`]+)`").ReplaceAllString(result, "$1")

	// Remove images ![alt](url)
	result = regexp.MustCompile(`!\[.*?\]\([^)]+\)`).ReplaceAllString(result, "")

	// Remove links [text](url) -> text
	result = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(result, "$1")

	// Remove headers (# ## ### etc)
	result = regexp.MustCompile(`(?m)^#{1,6}\s+`).ReplaceAllString(result, "")

	// Remove bold/italic markers
	result = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(result, "$1")
	result = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(result, "$1")
	result = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(result, "$1")
	result = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(result, "$1")

	// Remove strikethrough
	result = regexp.MustCompile(`~~([^~]+)~~`).ReplaceAllString(result, "$1")

	// Remove blockquotes
	result = regexp.MustCompile(`(?m)^>\s+`).ReplaceAllString(result, "")

	// Remove unordered list markers
	result = regexp.MustCompile(`(?m)^[\s]*[-*+]\s+`).ReplaceAllString(result, "")

	// Remove ordered list markers
	result = regexp.MustCompile(`(?m)^[\s]*\d+\.\s+`).ReplaceAllString(result, "")

	// Remove horizontal rules
	result = regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`).ReplaceAllString(result, "")

	// Clean up extra whitespace
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	return strings.TrimSpace(result)
}
