package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"todo-api/internal/model"
	"todo-api/internal/testutil"
)

// TestTagList_Success tests successful tag list retrieval
func TestTagList_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("taglist@example.com")
	color := "#FF0000"
	f.CreateTag(user.ID, "work", &color)
	f.CreateTag(user.ID, "personal", nil)

	rec, err := f.CallAuthTag(token, http.MethodGet, "/api/v1/tags", "", f.TagHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	tags := testutil.JSONArrayResponse(t, rec)
	assert.Len(t, tags, 2)

	// Should be sorted by name (personal < work)
	first := testutil.TagAt(tags, 0)
	assert.Equal(t, "personal", first["name"])
}

// TestTagList_Empty tests empty tag list
func TestTagList_Empty(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("tagempty@example.com")

	rec, err := f.CallAuthTag(token, http.MethodGet, "/api/v1/tags", "", f.TagHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	tags := testutil.JSONArrayResponse(t, rec)
	assert.Len(t, tags, 0)
}

// TestTagList_UserScope tests that users only see their own tags
func TestTagList_UserScope(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, token1 := f.CreateUser("taguser1@example.com")
	user2, _ := f.CreateUser("taguser2@example.com")

	f.CreateTag(user1.ID, "user1tag", nil)
	f.CreateTag(user2.ID, "user2tag", nil)

	rec, err := f.CallAuthTag(token1, http.MethodGet, "/api/v1/tags", "", f.TagHandler.List)
	require.NoError(t, err)

	tags := testutil.JSONArrayResponse(t, rec)
	assert.Len(t, tags, 1)
	assert.Equal(t, "user1tag", testutil.TagAt(tags, 0)["name"])
}

// TestTagCreate_Success tests successful tag creation
func TestTagCreate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("tagcreate@example.com")

	body := `{"tag":{"name":"NewTag","color":"#FF5500"}}`
	rec, err := f.CallAuthTag(token, http.MethodPost, "/api/v1/tags", body, f.TagHandler.Create)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	response := testutil.JSONResponse(t, rec)
	tag := testutil.ExtractTagFromData(response)
	assert.Equal(t, "newtag", tag["name"]) // Should be lowercase
	assert.Equal(t, "#FF5500", tag["color"])
}

// TestTagCreate_WithoutColor tests tag creation without color
func TestTagCreate_WithoutColor(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("tagnocolor@example.com")

	body := `{"tag":{"name":"NoColorTag"}}`
	rec, err := f.CallAuthTag(token, http.MethodPost, "/api/v1/tags", body, f.TagHandler.Create)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	response := testutil.JSONResponse(t, rec)
	tag := testutil.ExtractTagFromData(response)
	assert.Equal(t, "nocolortag", tag["name"])
	// Color can be nil or default value
}

// TestTagCreate_DuplicateName tests tag creation with duplicate name
func TestTagCreate_DuplicateName(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("tagdup@example.com")
	f.CreateTag(user.ID, "work", nil)

	// Try to create with same name (will be normalized to lowercase)
	body := `{"tag":{"name":"WORK"}}`
	_, err := f.CallAuthTag(token, http.MethodPost, "/api/v1/tags", body, f.TagHandler.Create)
	require.Error(t, err)
}

// TestTagCreate_NameNormalized tests that tag names are normalized to lowercase
func TestTagCreate_NameNormalized(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("tagnorm@example.com")

	body := `{"tag":{"name":"  MixedCase  "}}`
	rec, err := f.CallAuthTag(token, http.MethodPost, "/api/v1/tags", body, f.TagHandler.Create)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	tag := testutil.ExtractTagFromData(response)
	assert.Equal(t, "mixedcase", tag["name"]) // Trimmed and lowercased
}

// TestTagCreate_ValidationError tests tag creation with validation errors
func TestTagCreate_ValidationError(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("tagvalid@example.com")

	tests := []struct {
		name string
		body string
	}{
		{name: "missing name", body: `{"tag":{"color":"#FF0000"}}`},
		{name: "empty name", body: `{"tag":{"name":""}}`},
		{name: "blank name", body: `{"tag":{"name":"   "}}`},
		{name: "invalid color", body: `{"tag":{"name":"test","color":"invalid"}}`},
		{name: "name too long", body: `{"tag":{"name":"` + strings.Repeat("a", 31) + `"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.CallAuthTag(token, http.MethodPost, "/api/v1/tags", tt.body, f.TagHandler.Create)
			require.Error(t, err)
		})
	}
}

// TestTagShow_Success tests successful tag retrieval by ID
func TestTagShow_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("tagshow@example.com")
	tag := f.CreateTag(user.ID, "showme", nil)

	rec, err := f.CallAuthTag(token, http.MethodGet, testutil.TagPath(tag.ID), "", f.TagHandler.Show)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	tagResp := testutil.ExtractTag(response)
	assert.Equal(t, "showme", tagResp["name"])
}

// TestTagShow_NotFound tests tag not found error
func TestTagShow_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("tagnotfound@example.com")

	_, err := f.CallAuthTag(token, http.MethodGet, "/api/v1/tags/99999", "", f.TagHandler.Show)
	require.Error(t, err)
}

// TestTagShow_OtherUser tests accessing other user's tag
func TestTagShow_OtherUser(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("tagowner@example.com")
	_, token2 := f.CreateUser("tagother@example.com")

	tag := f.CreateTag(user1.ID, "privatetag", nil)

	_, err := f.CallAuthTag(token2, http.MethodGet, testutil.TagPath(tag.ID), "", f.TagHandler.Show)
	require.Error(t, err)
}

// TestTagUpdate_Success tests successful tag update
func TestTagUpdate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("tagupdate@example.com")
	tag := f.CreateTag(user.ID, "original", nil)

	body := `{"tag":{"name":"Updated","color":"#00FF00"}}`
	rec, err := f.CallAuthTag(token, http.MethodPatch, testutil.TagPath(tag.ID), body, f.TagHandler.Update)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	tagResp := testutil.ExtractTag(response)
	assert.Equal(t, "updated", tagResp["name"]) // Normalized to lowercase
	assert.Equal(t, "#00FF00", tagResp["color"])
}

// TestTagUpdate_PartialUpdate tests partial tag update
func TestTagUpdate_PartialUpdate(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("tagpartial@example.com")
	color := "#FF0000"
	tag := f.CreateTag(user.ID, "original", &color)

	body := `{"tag":{"color":"#00FF00"}}`
	rec, err := f.CallAuthTag(token, http.MethodPatch, testutil.TagPath(tag.ID), body, f.TagHandler.Update)
	require.NoError(t, err)

	response := testutil.JSONResponse(t, rec)
	tagResp := testutil.ExtractTag(response)
	assert.Equal(t, "original", tagResp["name"]) // Name unchanged
	assert.Equal(t, "#00FF00", tagResp["color"])
}

// TestTagUpdate_DuplicateName tests tag update with duplicate name
func TestTagUpdate_DuplicateName(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("tagdupupdate@example.com")
	f.CreateTag(user.ID, "existing", nil)
	tag := f.CreateTag(user.ID, "toupdate", nil)

	body := `{"tag":{"name":"existing"}}`
	_, err := f.CallAuthTag(token, http.MethodPatch, testutil.TagPath(tag.ID), body, f.TagHandler.Update)
	require.Error(t, err)
}

// TestTagDelete_Success tests successful tag deletion
func TestTagDelete_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("tagdelete@example.com")
	tag := f.CreateTag(user.ID, "deleteme", nil)

	rec, err := f.CallAuthTag(token, http.MethodDelete, testutil.TagPath(tag.ID), "", f.TagHandler.Delete)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify tag is deleted
	var count int64
	f.DB.Model(&model.Tag{}).Where("id = ?", tag.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestTagDelete_OtherUser tests deleting other user's tag
func TestTagDelete_OtherUser(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("tagdelowner@example.com")
	_, token2 := f.CreateUser("tagdelother@example.com")

	tag := f.CreateTag(user1.ID, "notyours", nil)

	_, err := f.CallAuthTag(token2, http.MethodDelete, testutil.TagPath(tag.ID), "", f.TagHandler.Delete)
	require.Error(t, err)

	// Verify tag still exists
	var count int64
	f.DB.Model(&model.Tag{}).Where("id = ?", tag.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}
