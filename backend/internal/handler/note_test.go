package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"todo-api/internal/handler"
	"todo-api/internal/model"
	"todo-api/internal/testutil"
)

// =============================================================================
// List Tests
// =============================================================================

func TestNoteList_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	// Create some notes
	f.CreateNote(user.ID, "Note 1", "Body 1")
	f.CreateNote(user.ID, "Note 2", "Body 2")

	rec, err := f.CallAuthNote(token, http.MethodGet, "/api/v1/notes", "", f.NoteHandler.List)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handler.NoteListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, int64(2), resp.Meta.Total)
}

func TestNoteList_Empty(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	_, token := f.CreateUser("test@example.com")

	rec, err := f.CallAuthNote(token, http.MethodGet, "/api/v1/notes", "", f.NoteHandler.List)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handler.NoteListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 0)
	assert.Equal(t, int64(0), resp.Meta.Total)
}

func TestNoteList_ExcludesOtherUserNotes(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user1, token1 := f.CreateUser("user1@example.com")
	user2, _ := f.CreateUser("user2@example.com")

	f.CreateNote(user1.ID, "User1 Note", "Body")
	f.CreateNote(user2.ID, "User2 Note", "Body")

	rec, err := f.CallAuthNote(token1, http.MethodGet, "/api/v1/notes", "", f.NoteHandler.List)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handler.NoteListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "User1 Note", *resp.Data[0].Title)
}

func TestNoteList_FilterTrashed(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	// Create active and trashed notes
	f.CreateNote(user.ID, "Active Note", "Body")
	now := time.Now()
	title := "Trashed Note"
	f.CreateNoteWithOptions(user.ID, testutil.NoteOptions{
		Title:     &title,
		TrashedAt: &now,
	})

	// Default: exclude trashed
	rec, err := f.CallAuthNote(token, http.MethodGet, "/api/v1/notes", "", f.NoteHandler.List)
	require.NoError(t, err)

	var resp handler.NoteListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "Active Note", *resp.Data[0].Title)

	// trashed=true: show only trashed
	rec, err = f.CallAuthNote(token, http.MethodGet, "/api/v1/notes?trashed=true", "", f.NoteHandler.List)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "Trashed Note", *resp.Data[0].Title)
}

func TestNoteList_FilterArchived(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	// Create active and archived notes
	f.CreateNote(user.ID, "Active Note", "Body")
	now := time.Now()
	title := "Archived Note"
	f.CreateNoteWithOptions(user.ID, testutil.NoteOptions{
		Title:      &title,
		ArchivedAt: &now,
	})

	// archived=true: show only archived
	rec, err := f.CallAuthNote(token, http.MethodGet, "/api/v1/notes?archived=true", "", f.NoteHandler.List)
	require.NoError(t, err)

	var resp handler.NoteListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "Archived Note", *resp.Data[0].Title)
}

func TestNoteList_FilterPinned(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	// Create pinned and unpinned notes
	f.CreateNote(user.ID, "Normal Note", "Body")
	title := "Pinned Note"
	f.CreateNoteWithOptions(user.ID, testutil.NoteOptions{
		Title:  &title,
		Pinned: true,
	})

	// pinned=true: show only pinned
	rec, err := f.CallAuthNote(token, http.MethodGet, "/api/v1/notes?pinned=true", "", f.NoteHandler.List)
	require.NoError(t, err)

	var resp handler.NoteListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "Pinned Note", *resp.Data[0].Title)
}

// =============================================================================
// Create Tests
// =============================================================================

func TestNoteCreate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	_, token := f.CreateUser("test@example.com")

	body := `{"note":{"title":"My Note","body_md":"# Hello\n\nWorld"}}`
	rec, err := f.CallAuthNote(token, http.MethodPost, "/api/v1/notes", body, f.NoteHandler.Create)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]handler.NoteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "My Note", *resp["data"].Title)
	assert.Equal(t, "# Hello\n\nWorld", *resp["data"].BodyMD)
}

func TestNoteCreate_Empty(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	_, token := f.CreateUser("test@example.com")

	body := `{"note":{}}`
	rec, err := f.CallAuthNote(token, http.MethodPost, "/api/v1/notes", body, f.NoteHandler.Create)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestNoteCreate_CreatesInitialRevision(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	body := `{"note":{"title":"My Note","body_md":"Content"}}`
	rec, err := f.CallAuthNote(token, http.MethodPost, "/api/v1/notes", body, f.NoteHandler.Create)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]handler.NoteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	noteID := resp["data"].ID

	// Check revision was created
	revisions, total, err := f.NoteRevisionRepo.FindByNoteID(noteID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "My Note", *revisions[0].Title)
	assert.Equal(t, user.ID, revisions[0].UserID)
}

func TestNoteCreate_ValidationError_TitleTooLong(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	_, token := f.CreateUser("test@example.com")

	// Title > 150 characters
	longTitle := make([]byte, 151)
	for i := range longTitle {
		longTitle[i] = 'a'
	}
	body := `{"note":{"title":"` + string(longTitle) + `"}}`
	rec, _ := f.CallAuthNote(token, http.MethodPost, "/api/v1/notes", body, f.NoteHandler.Create)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// =============================================================================
// Show Tests
// =============================================================================

func TestNoteShow_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "My Note", "Body")

	rec, err := f.CallAuthNote(token, http.MethodGet, testutil.NotePath(note.ID), "", f.NoteHandler.Show)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]handler.NoteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "My Note", *resp["data"].Title)
}

func TestNoteShow_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	_, token := f.CreateUser("test@example.com")

	rec, _ := f.CallAuthNote(token, http.MethodGet, testutil.NotePath(99999), "", f.NoteHandler.Show)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestNoteShow_OtherUserNote(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	_, token1 := f.CreateUser("user1@example.com")
	user2, _ := f.CreateUser("user2@example.com")

	note := f.CreateNote(user2.ID, "User2 Note", "Body")

	rec, _ := f.CallAuthNote(token1, http.MethodGet, testutil.NotePath(note.ID), "", f.NoteHandler.Show)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// =============================================================================
// Update Tests
// =============================================================================

func TestNoteUpdate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Original", "Body")

	body := `{"note":{"title":"Updated Title"}}`
	rec, err := f.CallAuthNote(token, http.MethodPatch, testutil.NotePath(note.ID), body, f.NoteHandler.Update)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]handler.NoteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Updated Title", *resp["data"].Title)
}

func TestNoteUpdate_BodyMD_CreatesRevision(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Note", "Original Body")

	// Initial revision exists
	revisions, _, _ := f.NoteRevisionRepo.FindByNoteID(note.ID, 1, 10)
	initialCount := len(revisions)

	// Update body
	body := `{"note":{"body_md":"Updated Body"}}`
	_, err := f.CallAuthNote(token, http.MethodPatch, testutil.NotePath(note.ID), body, f.NoteHandler.Update)
	require.NoError(t, err)

	// Check new revision created
	revisions, _, _ = f.NoteRevisionRepo.FindByNoteID(note.ID, 1, 10)
	assert.Equal(t, initialCount+1, len(revisions))
}

func TestNoteUpdate_TitleOnly_NoRevision(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Note", "Body")

	// Get initial revision count
	revisions, _, _ := f.NoteRevisionRepo.FindByNoteID(note.ID, 1, 10)
	initialCount := len(revisions)

	// Update title only
	body := `{"note":{"title":"New Title"}}`
	_, err := f.CallAuthNote(token, http.MethodPatch, testutil.NotePath(note.ID), body, f.NoteHandler.Update)
	require.NoError(t, err)

	// Check no new revision
	revisions, _, _ = f.NoteRevisionRepo.FindByNoteID(note.ID, 1, 10)
	assert.Equal(t, initialCount, len(revisions))
}

func TestNoteUpdate_Archive(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Note", "Body")

	body := `{"note":{"archived":true}}`
	rec, err := f.CallAuthNote(token, http.MethodPatch, testutil.NotePath(note.ID), body, f.NoteHandler.Update)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]handler.NoteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp["data"].Archived)
	assert.NotNil(t, resp["data"].ArchivedAt)
}

func TestNoteUpdate_Pin(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Note", "Body")

	body := `{"note":{"pinned":true}}`
	rec, err := f.CallAuthNote(token, http.MethodPatch, testutil.NotePath(note.ID), body, f.NoteHandler.Update)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]handler.NoteResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp["data"].Pinned)
}

func TestNoteUpdate_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	_, token := f.CreateUser("test@example.com")

	body := `{"note":{"title":"Updated"}}`
	rec, _ := f.CallAuthNote(token, http.MethodPatch, testutil.NotePath(99999), body, f.NoteHandler.Update)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestNoteDelete_SoftDelete(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Note", "Body")

	rec, err := f.CallAuthNote(token, http.MethodDelete, testutil.NotePath(note.ID), "", f.NoteHandler.Delete)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Note should still exist with trashed_at set
	var updatedNote model.Note
	err = f.DB.First(&updatedNote, note.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, updatedNote.TrashedAt)
}

func TestNoteDelete_ForceDelete(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Note", "Body")

	rec, _ := f.CallAuthNote(token, http.MethodDelete, testutil.NotePath(note.ID)+"?force=true", "", f.NoteHandler.Delete)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Note should be completely deleted
	var count int64
	f.DB.Model(&model.Note{}).Where("id = ?", note.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestNoteDelete_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	_, token := f.CreateUser("test@example.com")

	rec, _ := f.CallAuthNote(token, http.MethodDelete, testutil.NotePath(99999), "", f.NoteHandler.Delete)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// =============================================================================
// Revision Tests
// =============================================================================

func TestNoteListRevisions_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Note", "Body")

	rec, err := f.CallAuthNote(token, http.MethodGet, testutil.NoteRevisionsPath(note.ID), "", f.NoteHandler.ListRevisions)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handler.RevisionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.GreaterOrEqual(t, len(resp.Data), 1) // At least initial revision
}

func TestNoteListRevisions_NoteNotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	_, token := f.CreateUser("test@example.com")

	rec, _ := f.CallAuthNote(token, http.MethodGet, testutil.NoteRevisionsPath(99999), "", f.NoteHandler.ListRevisions)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestNoteRestoreRevision_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	// Create note and update body to create revisions
	note := f.CreateNote(user.ID, "Original Title", "Original Body")

	// Update body to create a new revision
	updateBody := `{"note":{"body_md":"Updated Body"}}`
	_, err := f.CallAuthNote(token, http.MethodPatch, testutil.NotePath(note.ID), updateBody, f.NoteHandler.Update)
	require.NoError(t, err)

	// Get revisions
	revisions, _, _ := f.NoteRevisionRepo.FindByNoteID(note.ID, 1, 10)
	require.GreaterOrEqual(t, len(revisions), 1)

	// Restore to first revision (oldest)
	oldestRevision := revisions[len(revisions)-1]

	rec, err := f.CallAuthNote(token, http.MethodPost, testutil.RestoreRevisionPath(note.ID, oldestRevision.ID), "", f.NoteHandler.RestoreRevision)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestNoteRestoreRevision_RevisionNotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Note", "Body")

	rec, _ := f.CallAuthNote(token, http.MethodPost, testutil.RestoreRevisionPath(note.ID, 99999), "", f.NoteHandler.RestoreRevision)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// =============================================================================
// Revision Limit Tests
// =============================================================================

func TestNote_RevisionLimit_EnforcedAt50(t *testing.T) {
	f := testutil.SetupTestFixture(t)
	user, token := f.CreateUser("test@example.com")

	note := f.CreateNote(user.ID, "Note", "Body 0")

	// Create 55 revisions by updating body
	for i := 1; i <= 55; i++ {
		body := `{"note":{"body_md":"Body ` + string(rune('0'+i%10)) + `"}}`
		_, err := f.CallAuthNote(token, http.MethodPatch, testutil.NotePath(note.ID), body, f.NoteHandler.Update)
		require.NoError(t, err)
	}

	// Check revision count is <= 50
	count, err := f.NoteRevisionRepo.CountByNoteID(note.ID)
	require.NoError(t, err)
	assert.LessOrEqual(t, count, int64(model.MaxRevisionsPerNote))
}
