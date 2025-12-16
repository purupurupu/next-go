package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"todo-api/internal/errors"
	"todo-api/internal/repository"
	"todo-api/internal/service"
	"todo-api/pkg/response"
	"todo-api/pkg/util"
)

// NoteHandler handles note-related endpoints
type NoteHandler struct {
	noteService  *service.NoteService
	noteRepo     *repository.NoteRepository
	revisionRepo *repository.NoteRevisionRepository
}

// NewNoteHandler creates a new NoteHandler
func NewNoteHandler(
	noteService *service.NoteService,
	noteRepo *repository.NoteRepository,
	revisionRepo *repository.NoteRevisionRepository,
) *NoteHandler {
	return &NoteHandler{
		noteService:  noteService,
		noteRepo:     noteRepo,
		revisionRepo: revisionRepo,
	}
}

// CreateNoteRequest represents the request body for creating a note
type CreateNoteRequest struct {
	Title  *string `json:"title" validate:"omitempty,max=150"`
	BodyMD *string `json:"body_md" validate:"omitempty,max=100000"`
	Pinned *bool   `json:"pinned"`
}

// UpdateNoteRequest represents the request body for updating a note
type UpdateNoteRequest struct {
	Title    *string `json:"title" validate:"omitempty,max=150"`
	BodyMD   *string `json:"body_md" validate:"omitempty,max=100000"`
	Pinned   *bool   `json:"pinned"`
	Archived *bool   `json:"archived"`
	Trashed  *bool   `json:"trashed"`
}

// NoteResponse represents a note in API responses
type NoteResponse struct {
	ID           int64   `json:"id"`
	Title        *string `json:"title"`
	BodyMD       *string `json:"body_md"`
	Pinned       bool    `json:"pinned"`
	Archived     bool    `json:"archived"`
	Trashed      bool    `json:"trashed"`
	ArchivedAt   *string `json:"archived_at"`
	TrashedAt    *string `json:"trashed_at"`
	LastEditedAt string  `json:"last_edited_at"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// NoteListResponse represents a list of notes with pagination
type NoteListResponse struct {
	Data []NoteResponse `json:"data"`
	Meta NoteMeta       `json:"meta"`
}

// NoteMeta represents pagination metadata
type NoteMeta struct {
	Total       int64 `json:"total"`
	CurrentPage int   `json:"current_page"`
	TotalPages  int   `json:"total_pages"`
	PerPage     int   `json:"per_page"`
}

// RevisionResponse represents a note revision in API responses
type RevisionResponse struct {
	ID        int64   `json:"id"`
	NoteID    int64   `json:"note_id"`
	Title     *string `json:"title"`
	BodyMD    *string `json:"body_md"`
	CreatedAt string  `json:"created_at"`
}

// RevisionListResponse represents a list of revisions with pagination
type RevisionListResponse struct {
	Data []RevisionResponse `json:"data"`
	Meta NoteMeta           `json:"meta"`
}

// List retrieves notes with optional filters
// GET /api/v1/notes?q=&archived=&trashed=&pinned=&page=&per_page=
func (h *NoteHandler) List(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	// Parse query parameters
	query := c.QueryParam("q")

	var archived *bool
	if archivedStr := c.QueryParam("archived"); archivedStr != "" {
		b := archivedStr == "true"
		archived = &b
	}

	var trashed *bool
	if trashedStr := c.QueryParam("trashed"); trashedStr != "" {
		b := trashedStr == "true"
		trashed = &b
	}

	var pinned *bool
	if pinnedStr := c.QueryParam("pinned"); pinnedStr != "" {
		b := pinnedStr == "true"
		pinned = &b
	}

	page := 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 20
	if perPageStr := c.QueryParam("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	// Search notes
	notes, total, err := h.noteRepo.Search(repository.NoteSearchInput{
		UserID:   currentUser.ID,
		Query:    query,
		Archived: archived,
		Trashed:  trashed,
		Pinned:   pinned,
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		return errors.InternalErrorWithLog(err, "NoteHandler.List: failed to search notes")
	}

	// Convert to response format
	noteResponses := make([]NoteResponse, len(notes))
	for i, note := range notes {
		var archivedAt, trashedAt *string
		if note.ArchivedAt != nil {
			s := util.FormatRFC3339(*note.ArchivedAt)
			archivedAt = &s
		}
		if note.TrashedAt != nil {
			s := util.FormatRFC3339(*note.TrashedAt)
			trashedAt = &s
		}
		noteResponses[i] = NoteResponse{
			ID:           note.ID,
			Title:        note.Title,
			BodyMD:       note.BodyMD,
			Pinned:       note.Pinned,
			Archived:     note.IsArchived(),
			Trashed:      note.IsTrashed(),
			ArchivedAt:   archivedAt,
			TrashedAt:    trashedAt,
			LastEditedAt: util.FormatRFC3339(note.LastEditedAt),
			CreatedAt:    util.FormatRFC3339(note.CreatedAt),
			UpdatedAt:    util.FormatRFC3339(note.UpdatedAt),
		}
	}

	// Calculate total pages
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, NoteListResponse{
		Data: noteResponses,
		Meta: NoteMeta{
			Total:       total,
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     perPage,
		},
	})
}

// Create creates a new note
// POST /api/v1/notes
func (h *NoteHandler) Create(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	var req CreateNoteRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	note, err := h.noteService.Create(service.CreateNoteInput{
		UserID: currentUser.ID,
		Title:  req.Title,
		BodyMD: req.BodyMD,
		Pinned: req.Pinned,
	})
	if err != nil {
		return err
	}

	var archivedAt, trashedAt *string
	if note.ArchivedAt != nil {
		s := util.FormatRFC3339(*note.ArchivedAt)
		archivedAt = &s
	}
	if note.TrashedAt != nil {
		s := util.FormatRFC3339(*note.TrashedAt)
		trashedAt = &s
	}

	resp := NoteResponse{
		ID:           note.ID,
		Title:        note.Title,
		BodyMD:       note.BodyMD,
		Pinned:       note.Pinned,
		Archived:     note.IsArchived(),
		Trashed:      note.IsTrashed(),
		ArchivedAt:   archivedAt,
		TrashedAt:    trashedAt,
		LastEditedAt: util.FormatRFC3339(note.LastEditedAt),
		CreatedAt:    util.FormatRFC3339(note.CreatedAt),
		UpdatedAt:    util.FormatRFC3339(note.UpdatedAt),
	}

	return response.Created(c, resp)
}

// Show retrieves a single note
// GET /api/v1/notes/:id
func (h *NoteHandler) Show(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	noteID, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	note, err := h.noteRepo.FindByIDIncludingTrashed(noteID, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Note", noteID)
		}
		return errors.InternalErrorWithLog(err, "NoteHandler.Show: failed to fetch note")
	}

	var archivedAt, trashedAt *string
	if note.ArchivedAt != nil {
		s := util.FormatRFC3339(*note.ArchivedAt)
		archivedAt = &s
	}
	if note.TrashedAt != nil {
		s := util.FormatRFC3339(*note.TrashedAt)
		trashedAt = &s
	}

	resp := NoteResponse{
		ID:           note.ID,
		Title:        note.Title,
		BodyMD:       note.BodyMD,
		Pinned:       note.Pinned,
		Archived:     note.IsArchived(),
		Trashed:      note.IsTrashed(),
		ArchivedAt:   archivedAt,
		TrashedAt:    trashedAt,
		LastEditedAt: util.FormatRFC3339(note.LastEditedAt),
		CreatedAt:    util.FormatRFC3339(note.CreatedAt),
		UpdatedAt:    util.FormatRFC3339(note.UpdatedAt),
	}

	return c.JSON(http.StatusOK, resp)
}

// Update updates a note
// PATCH /api/v1/notes/:id
func (h *NoteHandler) Update(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	noteID, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	// Check note exists
	exists, err := h.noteRepo.ExistsByID(noteID, currentUser.ID)
	if err != nil {
		return errors.InternalErrorWithLog(err, "NoteHandler.Update: failed to check note existence")
	}
	if !exists {
		return errors.NotFound("Note", noteID)
	}

	var req UpdateNoteRequest
	if err := BindAndValidate(c, &req); err != nil {
		return err
	}

	note, err := h.noteService.Update(noteID, currentUser.ID, service.UpdateNoteInput{
		Title:    req.Title,
		BodyMD:   req.BodyMD,
		Pinned:   req.Pinned,
		Archived: req.Archived,
		Trashed:  req.Trashed,
	})
	if err != nil {
		return err
	}

	var archivedAt, trashedAt *string
	if note.ArchivedAt != nil {
		s := util.FormatRFC3339(*note.ArchivedAt)
		archivedAt = &s
	}
	if note.TrashedAt != nil {
		s := util.FormatRFC3339(*note.TrashedAt)
		trashedAt = &s
	}

	resp := NoteResponse{
		ID:           note.ID,
		Title:        note.Title,
		BodyMD:       note.BodyMD,
		Pinned:       note.Pinned,
		Archived:     note.IsArchived(),
		Trashed:      note.IsTrashed(),
		ArchivedAt:   archivedAt,
		TrashedAt:    trashedAt,
		LastEditedAt: util.FormatRFC3339(note.LastEditedAt),
		CreatedAt:    util.FormatRFC3339(note.CreatedAt),
		UpdatedAt:    util.FormatRFC3339(note.UpdatedAt),
	}

	return response.OK(c, resp)
}

// Delete soft/hard deletes a note
// DELETE /api/v1/notes/:id?force=true
func (h *NoteHandler) Delete(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	noteID, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	force := c.QueryParam("force") == "true"

	if err := h.noteService.Delete(noteID, currentUser.ID, force); err != nil {
		return err
	}

	return response.NoContent(c)
}

// ListRevisions retrieves revisions for a note
// GET /api/v1/notes/:id/revisions
func (h *NoteHandler) ListRevisions(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	noteID, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	// Check note exists and belongs to user
	exists, err := h.noteRepo.ExistsByID(noteID, currentUser.ID)
	if err != nil {
		return errors.InternalErrorWithLog(err, "NoteHandler.ListRevisions: failed to check note existence")
	}
	if !exists {
		return errors.NotFound("Note", noteID)
	}

	// Parse pagination
	page := 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 20
	if perPageStr := c.QueryParam("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	revisions, total, err := h.revisionRepo.FindByNoteID(noteID, page, perPage)
	if err != nil {
		return errors.InternalErrorWithLog(err, "NoteHandler.ListRevisions: failed to fetch revisions")
	}

	// Convert to response format
	revisionResponses := make([]RevisionResponse, len(revisions))
	for i, rev := range revisions {
		revisionResponses[i] = RevisionResponse{
			ID:        rev.ID,
			NoteID:    rev.NoteID,
			Title:     rev.Title,
			BodyMD:    rev.BodyMD,
			CreatedAt: util.FormatRFC3339(rev.CreatedAt),
		}
	}

	// Calculate total pages
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, RevisionListResponse{
		Data: revisionResponses,
		Meta: NoteMeta{
			Total:       total,
			CurrentPage: page,
			TotalPages:  totalPages,
			PerPage:     perPage,
		},
	})
}

// RestoreRevision restores a note from a revision
// POST /api/v1/notes/:id/revisions/:revision_id/restore
func (h *NoteHandler) RestoreRevision(c echo.Context) error {
	currentUser, err := GetCurrentUserOrFail(c)
	if err != nil {
		return err
	}

	noteID, err := ParseIDParam(c, "id")
	if err != nil {
		return err
	}

	revisionID, err := ParseIDParam(c, "revision_id")
	if err != nil {
		return err
	}

	note, err := h.noteService.RestoreRevision(noteID, revisionID, currentUser.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("Note", noteID)
		}
		return err
	}

	var archivedAt, trashedAt *string
	if note.ArchivedAt != nil {
		s := util.FormatRFC3339(*note.ArchivedAt)
		archivedAt = &s
	}
	if note.TrashedAt != nil {
		s := util.FormatRFC3339(*note.TrashedAt)
		trashedAt = &s
	}

	resp := NoteResponse{
		ID:           note.ID,
		Title:        note.Title,
		BodyMD:       note.BodyMD,
		Pinned:       note.Pinned,
		Archived:     note.IsArchived(),
		Trashed:      note.IsTrashed(),
		ArchivedAt:   archivedAt,
		TrashedAt:    trashedAt,
		LastEditedAt: util.FormatRFC3339(note.LastEditedAt),
		CreatedAt:    util.FormatRFC3339(note.CreatedAt),
		UpdatedAt:    util.FormatRFC3339(note.UpdatedAt),
	}

	return response.OK(c, resp)
}
