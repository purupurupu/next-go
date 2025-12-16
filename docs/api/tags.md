# Tags API

## Overview

Tags provide a flexible way to label and organize todos using a many-to-many relationship. Each tag has a name and optional color for visual distinction.

## Base URL

All endpoints are prefixed with `/api/v1`:
```
http://localhost:3001/api/v1/tags
```

## Endpoints

### List Tags

Retrieve all tags for the authenticated user, sorted by name.

**Endpoint:** `GET /api/v1/tags`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "important",
    "color": "#EF4444",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  {
    "id": 2,
    "name": "work",
    "color": "#3B82F6",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

### Get Tag

Retrieve a specific tag by ID.

**Endpoint:** `GET /api/v1/tags/:id`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200 OK):**
```json
{
  "id": 1,
  "name": "work",
  "color": "#3B82F6",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Tag with id 123 not found"
  }
}
```

### Create Tag

Create a new tag for the authenticated user.

**Endpoint:** `POST /api/v1/tags`

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Important",
  "color": "#EF4444"
}
```

**Success Response (201 Created):**
```json
{
  "id": 1,
  "name": "important",
  "color": "#EF4444",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

**Note:** Tag names are normalized to lowercase before saving (e.g., "Important" becomes "important").

**Error Response (422 Unprocessable Entity):**
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "details": {
      "validation_errors": {
        "name": ["required", "notblank"],
        "color": ["hexcolor"]
      }
    }
  }
}
```

**Error Response (409 Conflict - Duplicate Name):**
```json
{
  "error": {
    "code": "DUPLICATE_RESOURCE",
    "message": "Tag with this name already exists"
  }
}
```

### Update Tag

Update an existing tag.

**Endpoint:** `PATCH /api/v1/tags/:id`

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "High Priority",
  "color": "#DC2626"
}
```

Both fields are optional for partial updates.

**Success Response (200 OK):**
```json
{
  "id": 1,
  "name": "high priority",
  "color": "#DC2626",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Tag with id 123 not found"
  }
}
```

### Delete Tag

Delete a tag. This will also remove the tag from all associated todos via the todo_tags join table.

**Endpoint:** `DELETE /api/v1/tags/:id`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (204 No Content):**
```
(empty body)
```

**Error Response (404 Not Found):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Tag with id 123 not found"
  }
}
```

## Tag Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | Integer | Read-only | Unique identifier |
| `name` | String | Yes | Tag name (unique per user, stored lowercase) |
| `color` | String | No | Hex color code (default: "#6B7280") |
| `created_at` | String (RFC3339) | Read-only | Creation timestamp |
| `updated_at` | String (RFC3339) | Read-only | Last update timestamp |

## Validation Rules

### Name
- **Required**: Cannot be blank
- **Not Blank**: Cannot be only whitespace
- **Uniqueness**: Must be unique per user (case-insensitive)
- **Length**: Maximum 30 characters
- **Normalization**: Trimmed and converted to lowercase before saving

### Color
- **Optional**: Can be omitted (defaults to "#6B7280")
- **Format**: Must be a valid hex color code if provided (e.g., "#EF4444")
- **Pattern**: Must match `/^#[0-9a-fA-F]{6}$/`

## Business Rules

1. **User Scoped**: Users can only see and manage their own tags
2. **Unique Names**: Tag names must be unique within a user's tags (case-insensitive due to lowercase normalization)
3. **Many-to-Many**: Tags have a many-to-many relationship with todos through the `todo_tags` join table
4. **Cascade Behavior**: When a tag is deleted, all entries in `todo_tags` referencing it are removed
5. **Optional Color**: Tags can exist without a color (uses default gray)

## Error Handling

All endpoints return consistent error responses. Common error scenarios:

- **401 Unauthorized**: Missing or invalid JWT token
- **404 Not Found**: Tag doesn't exist or doesn't belong to the user
- **409 Conflict**: Duplicate tag name
- **422 Unprocessable Entity**: Validation errors (invalid color format, blank name)

## Database Schema

### Tags Table
```sql
CREATE TABLE tags (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  name VARCHAR(30) NOT NULL,
  color VARCHAR(7) DEFAULT '#6B7280',
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  UNIQUE(user_id, name)
);
```

### Todo_Tags Join Table
```sql
CREATE TABLE todo_tags (
  id BIGSERIAL PRIMARY KEY,
  todo_id BIGINT NOT NULL REFERENCES todos(id) ON DELETE CASCADE,
  tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  UNIQUE(todo_id, tag_id)
);
```
