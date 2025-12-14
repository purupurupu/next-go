# Comments API

## Overview

The Comments API provides functionality for adding, viewing, updating, and soft-deleting comments on todos. Comments use polymorphic associations, allowing them to be attached to different types of resources (currently only todos).

## Implementation Status

| Backend | Status |
|---------|--------|
| Go (Echo) | ✅ Implemented |

## Authentication Required

All comment endpoints require a valid JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Endpoints

### List Comments

Get all comments for a specific todo.

**Endpoint:** `GET /api/v1/todos/:todo_id/comments`

**URL Parameters:**
- `todo_id` (required): ID of the todo

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "content": "This task needs more details about the API specifications",
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com"
    },
    "created_at": "2024-01-01T10:00:00.000Z",
    "updated_at": "2024-01-01T10:00:00.000Z",
    "editable": true
  },
  {
    "id": 2,
    "content": "I've added the specifications to the description",
    "user": {
      "id": 2,
      "name": "Jane Smith",
      "email": "jane@example.com"
    },
    "created_at": "2024-01-01T11:00:00.000Z",
    "updated_at": "2024-01-01T11:00:00.000Z",
    "editable": false
  }
]
```

**Notes:**
- Comments are returned in chronological order (oldest first)
- Empty array `[]` if no comments exist
- `editable` field indicates if the current user can edit/delete the comment

### Create Comment

Add a new comment to a todo.

**Endpoint:** `POST /api/v1/todos/:todo_id/comments`

**URL Parameters:**
- `todo_id` (required): ID of the todo

**Request Body:**
```json
{
  "comment": {
    "content": "This is a new comment"
  }
}
```

**Parameters:**
- `content` (required): The comment text

**Success Response (201 Created):**
```json
{
  "id": 3,
  "content": "This is a new comment",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  },
  "created_at": "2024-01-02T10:00:00.000Z",
  "updated_at": "2024-01-02T10:00:00.000Z",
  "editable": true
}
```

**Error Response (422 Unprocessable Entity):**
```json
{
  "errors": {
    "content": ["can't be blank"]
  }
}
```

### Update Comment

Update an existing comment. **Note: Comments can only be edited within 15 minutes of creation.**

**Endpoint:** `PATCH /api/v1/todos/:todo_id/comments/:id`

**URL Parameters:**
- `todo_id` (required): ID of the todo
- `id` (required): ID of the comment

**Request Body:**
```json
{
  "comment": {
    "content": "Updated comment text"
  }
}
```

**Parameters:**
- `content` (required): The updated comment text

**Success Response (200 OK):**
```json
{
  "id": 1,
  "content": "Updated comment text",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  },
  "created_at": "2024-01-01T10:00:00.000Z",
  "updated_at": "2024-01-02T11:00:00.000Z",
  "editable": true
}
```

**Error Responses:**
- **403 Forbidden:** User is not authorized to update this comment
- **404 Not Found:** Comment not found
- **410 Gone:** Edit time expired (15 minutes limit exceeded)
- **422 Unprocessable Entity:** Validation errors

### Delete Comment

Soft delete a comment. The comment is not permanently removed but marked as deleted.

**Endpoint:** `DELETE /api/v1/todos/:todo_id/comments/:id`

**URL Parameters:**
- `todo_id` (required): ID of the todo
- `id` (required): ID of the comment

**Success Response (204 No Content):**
No response body

**Error Responses:**
- **403 Forbidden:** User is not authorized to delete this comment
- **404 Not Found:** Comment not found

**Notes:**
- Comments are soft-deleted (marked with `deleted_at` timestamp)
- Deleted comments are excluded from list responses
- Deletion preserves comment history for audit purposes

## Data Validation

### Content
- Required field
- Cannot be empty
- Minimum length: 1 character
- Maximum length: 1000 characters

## Authorization Rules

1. **Viewing**: Authenticated users can view comments on their own todos
2. **Creating**: Authenticated users can create comments on their own todos
3. **Updating**: Users can only update their own comments **within 15 minutes of creation**
4. **Deleting**: Users can only delete their own comments

## Edit Time Limitation

Comments can only be edited within **15 minutes** of creation. After this period:
- The `editable` field in the response becomes `false`
- Update requests will return a **410 Gone** error with the message "Edit time has expired"

This design encourages thoughtful commenting while still allowing quick corrections to typos or mistakes.

## Frontend Integration Example

```typescript
// Using ApiClient pattern from the project
class CommentApiClient extends ApiClient {
  async getComments(todoId: number) {
    return this.get<{ comments: Comment[] }>(`/todos/${todoId}/comments`);
  }

  async createComment(todoId: number, content: string) {
    return this.post<{ data: { comment: Comment } }>(`/todos/${todoId}/comments`, {
      comment: { content }
    });
  }

  async updateComment(todoId: number, commentId: number, content: string) {
    return this.patch<{ comment: Comment }>(`/todos/${todoId}/comments/${commentId}`, {
      comment: { content }
    });
  }

  async deleteComment(todoId: number, commentId: number) {
    return this.delete(`/todos/${todoId}/comments/${commentId}`);
  }
}
```

## Performance Considerations

1. **Soft Delete**: Comments are soft-deleted to preserve history
2. **N+1 Queries**: User information is included to avoid additional queries
3. **Polymorphic Association**: Designed to support comments on other resources in the future
4. **Real-time Updates**: Consider implementing WebSocket support for live comment updates