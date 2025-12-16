package handler_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"todo-api/internal/testutil"
)

// =============================================================================
// Comment List Tests
// =============================================================================

func TestCommentList_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentlist@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")
	f.CreateComment(user.ID, todo.ID, "Comment 1")
	f.CreateComment(user.ID, todo.ID, "Comment 2")

	rec, err := f.CallAuth(token, http.MethodGet, testutil.TodoCommentsPath(todo.ID), "", f.CommentHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	comments := testutil.JSONArrayResponse(t, rec)
	assert.Len(t, comments, 2)
}

func TestCommentList_Empty(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentlistempty@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")

	rec, err := f.CallAuth(token, http.MethodGet, testutil.TodoCommentsPath(todo.ID), "", f.CommentHandler.List)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	comments := testutil.JSONArrayResponse(t, rec)
	assert.Len(t, comments, 0)
}

func TestCommentList_TodoNotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("commentlistnotfound@example.com")

	_, err := f.CallAuth(token, http.MethodGet, testutil.TodoCommentsPath(99999), "", f.CommentHandler.List)
	require.Error(t, err)
}

func TestCommentList_OtherUserTodo(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("commentlistuser1@example.com")
	_, token2 := f.CreateUser("commentlistuser2@example.com")
	todo := f.CreateTodo(user1.ID, "User1 Todo")

	_, err := f.CallAuth(token2, http.MethodGet, testutil.TodoCommentsPath(todo.ID), "", f.CommentHandler.List)
	require.Error(t, err)
}

func TestCommentList_ExcludesSoftDeleted(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentlistdeleted@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")
	comment1 := f.CreateComment(user.ID, todo.ID, "Comment 1")
	f.CreateComment(user.ID, todo.ID, "Comment 2")

	// Soft delete comment1
	require.NoError(t, f.CommentRepo.SoftDelete(comment1.ID))

	rec, err := f.CallAuth(token, http.MethodGet, testutil.TodoCommentsPath(todo.ID), "", f.CommentHandler.List)
	require.NoError(t, err)

	comments := testutil.JSONArrayResponse(t, rec)
	assert.Len(t, comments, 1)
}

// =============================================================================
// Comment Create Tests
// =============================================================================

func TestCommentCreate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentcreate@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")

	body := `{"comment":{"content":"New comment content"}}`
	rec, err := f.CallAuth(token, http.MethodPost, testutil.TodoCommentsPath(todo.ID), body, f.CommentHandler.Create)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	response := testutil.JSONResponse(t, rec)
	data := response["data"].(map[string]interface{})
	comment := data["comment"].(map[string]interface{})

	assert.Equal(t, "New comment content", comment["content"])
	assert.Equal(t, float64(user.ID), comment["user_id"])
	assert.Equal(t, "Todo", comment["commentable_type"])
	assert.Equal(t, float64(todo.ID), comment["commentable_id"])
	assert.True(t, comment["editable"].(bool))
}

func TestCommentCreate_TodoNotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	_, token := f.CreateUser("commentcreatenotfound@example.com")

	body := `{"comment":{"content":"New comment"}}`
	_, err := f.CallAuth(token, http.MethodPost, testutil.TodoCommentsPath(99999), body, f.CommentHandler.Create)
	require.Error(t, err)
}

func TestCommentCreate_OtherUserTodo(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, _ := f.CreateUser("commentcreateuser1@example.com")
	_, token2 := f.CreateUser("commentcreateuser2@example.com")
	todo := f.CreateTodo(user1.ID, "User1 Todo")

	body := `{"comment":{"content":"New comment"}}`
	_, err := f.CallAuth(token2, http.MethodPost, testutil.TodoCommentsPath(todo.ID), body, f.CommentHandler.Create)
	require.Error(t, err)
}

func TestCommentCreate_ValidationError_EmptyContent(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentcreateempty@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")

	body := `{"comment":{"content":""}}`
	_, err := f.CallAuth(token, http.MethodPost, testutil.TodoCommentsPath(todo.ID), body, f.CommentHandler.Create)
	require.Error(t, err)
}

func TestCommentCreate_ValidationError_TooLong(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentcreatetoolong@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")

	// Create content longer than 1000 characters
	longContent := make([]byte, 1001)
	for i := range longContent {
		longContent[i] = 'a'
	}

	body := `{"comment":{"content":"` + string(longContent) + `"}}`
	_, err := f.CallAuth(token, http.MethodPost, testutil.TodoCommentsPath(todo.ID), body, f.CommentHandler.Create)
	require.Error(t, err)
}

// =============================================================================
// Comment Update Tests
// =============================================================================

func TestCommentUpdate_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentupdate@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")
	comment := f.CreateComment(user.ID, todo.ID, "Original content")

	body := `{"comment":{"content":"Updated content"}}`
	rec, err := f.CallAuth(token, http.MethodPatch, testutil.CommentPath(todo.ID, comment.ID), body, f.CommentHandler.Update)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	response := testutil.JSONResponse(t, rec)
	// Success response returns data directly (not wrapped in "data")
	updatedComment := response["comment"].(map[string]interface{})

	assert.Equal(t, "Updated content", updatedComment["content"])
}

func TestCommentUpdate_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentupdatenotfound@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")

	body := `{"comment":{"content":"Updated content"}}`
	_, err := f.CallAuth(token, http.MethodPatch, testutil.CommentPath(todo.ID, 99999), body, f.CommentHandler.Update)
	require.Error(t, err)
}

func TestCommentUpdate_OtherUserComment(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, token1 := f.CreateUser("commentupdateuser1@example.com")
	user2, _ := f.CreateUser("commentupdateuser2@example.com")
	todo := f.CreateTodo(user1.ID, "User1 Todo")
	comment := f.CreateComment(user2.ID, todo.ID, "User2 comment")

	body := `{"comment":{"content":"Updated content"}}`
	_, err := f.CallAuth(token1, http.MethodPatch, testutil.CommentPath(todo.ID, comment.ID), body, f.CommentHandler.Update)
	require.Error(t, err)
}

func TestCommentUpdate_EditTimeExpired(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentupdateexpired@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")

	// Create comment with created_at 16 minutes ago
	oldTime := time.Now().Add(-16 * time.Minute)
	comment := f.CreateCommentWithCreatedAt(user.ID, todo.ID, "Old comment", oldTime)

	body := `{"comment":{"content":"Updated content"}}`
	_, err := f.CallAuth(token, http.MethodPatch, testutil.CommentPath(todo.ID, comment.ID), body, f.CommentHandler.Update)
	require.Error(t, err)
}

func TestCommentUpdate_WrongTodo(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentupdatewrongtodo@example.com")
	todo1 := f.CreateTodo(user.ID, "Todo 1")
	todo2 := f.CreateTodo(user.ID, "Todo 2")
	comment := f.CreateComment(user.ID, todo1.ID, "Comment on Todo1")

	// Try to update comment using todo2's path
	body := `{"comment":{"content":"Updated content"}}`
	_, err := f.CallAuth(token, http.MethodPatch, testutil.CommentPath(todo2.ID, comment.ID), body, f.CommentHandler.Update)
	require.Error(t, err)
}

// =============================================================================
// Comment Delete Tests
// =============================================================================

func TestCommentDelete_Success(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentdelete@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")
	comment := f.CreateComment(user.ID, todo.ID, "Comment to delete")

	rec, err := f.CallAuth(token, http.MethodDelete, testutil.CommentPath(todo.ID, comment.ID), "", f.CommentHandler.Delete)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify comment is soft deleted
	exists, err := f.CommentRepo.ExistsByID(comment.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCommentDelete_NotFound(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentdeletenotfound@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")

	_, err := f.CallAuth(token, http.MethodDelete, testutil.CommentPath(todo.ID, 99999), "", f.CommentHandler.Delete)
	require.Error(t, err)
}

func TestCommentDelete_OtherUserComment(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, token1 := f.CreateUser("commentdeleteuser1@example.com")
	user2, _ := f.CreateUser("commentdeleteuser2@example.com")
	todo := f.CreateTodo(user1.ID, "User1 Todo")
	comment := f.CreateComment(user2.ID, todo.ID, "User2 comment")

	_, err := f.CallAuth(token1, http.MethodDelete, testutil.CommentPath(todo.ID, comment.ID), "", f.CommentHandler.Delete)
	require.Error(t, err)
}

func TestCommentDelete_AlreadyDeleted(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commentdeletealready@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")
	comment := f.CreateComment(user.ID, todo.ID, "Comment to delete")

	// Soft delete first
	require.NoError(t, f.CommentRepo.SoftDelete(comment.ID))

	// Try to delete again
	_, err := f.CallAuth(token, http.MethodDelete, testutil.CommentPath(todo.ID, comment.ID), "", f.CommentHandler.Delete)
	require.Error(t, err)
}

// =============================================================================
// Editable Logic Tests
// =============================================================================

func TestCommentEditable_Within15Minutes(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commenteditable@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")
	f.CreateComment(user.ID, todo.ID, "Recent comment")

	rec, err := f.CallAuth(token, http.MethodGet, testutil.TodoCommentsPath(todo.ID), "", f.CommentHandler.List)
	require.NoError(t, err)

	comments := testutil.JSONArrayResponse(t, rec)
	comment := comments[0].(map[string]interface{})

	assert.True(t, comment["editable"].(bool))
}

func TestCommentEditable_After15Minutes(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user, token := f.CreateUser("commenteditableold@example.com")
	todo := f.CreateTodo(user.ID, "Test Todo")

	// Create comment 16 minutes ago
	oldTime := time.Now().Add(-16 * time.Minute)
	f.CreateCommentWithCreatedAt(user.ID, todo.ID, "Old comment", oldTime)

	rec, err := f.CallAuth(token, http.MethodGet, testutil.TodoCommentsPath(todo.ID), "", f.CommentHandler.List)
	require.NoError(t, err)

	comments := testutil.JSONArrayResponse(t, rec)
	comment := comments[0].(map[string]interface{})

	assert.False(t, comment["editable"].(bool))
}

func TestCommentEditable_OtherUserAlwaysFalse(t *testing.T) {
	f := testutil.SetupTestFixture(t)

	user1, token1 := f.CreateUser("commenteditableotheruser1@example.com")
	user2, _ := f.CreateUser("commenteditableotheruser2@example.com")
	todo := f.CreateTodo(user1.ID, "User1 Todo")

	// User2 creates a comment (this won't actually work because of user scope,
	// but we can test the editable logic via direct DB creation)
	f.CreateComment(user2.ID, todo.ID, "User2 comment")

	rec, err := f.CallAuth(token1, http.MethodGet, testutil.TodoCommentsPath(todo.ID), "", f.CommentHandler.List)
	require.NoError(t, err)

	comments := testutil.JSONArrayResponse(t, rec)
	comment := comments[0].(map[string]interface{})

	// Even recent comment by other user should not be editable
	assert.False(t, comment["editable"].(bool))
}
