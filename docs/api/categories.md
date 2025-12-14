# Categories API

## Overview

Categories provide a way for users to organize their todos. Each category has a name and color for visual organization, and todos can optionally be assigned to categories.

## Base URL

All endpoints are prefixed with `/api/v1`:
```
http://localhost:3001/api/v1/categories
```

## Endpoints

### List Categories

Retrieve all categories for the authenticated user, sorted by name.

**Endpoint:** `GET /api/v1/categories`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200 OK):**
```json
{
  "categories": [
    {
      "id": 1,
      "user_id": 1,
      "name": "personal",
      "color": "#3742fa",
      "todo_count": 3,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "user_id": 1,
      "name": "work",
      "color": "#ff4757",
      "todo_count": 5,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Get Category

Retrieve a specific category by ID.

**Endpoint:** `GET /api/v1/categories/:id`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200 OK):**
```json
{
  "category": {
    "id": 1,
    "user_id": 1,
    "name": "work",
    "color": "#ff4757",
    "todo_count": 5,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**Error Response (404 Not Found):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Category with id 123 not found"
  }
}
```

### Create Category

Create a new category for the authenticated user.

**Endpoint:** `POST /api/v1/categories`

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "category": {
    "name": "Work",
    "color": "#ff4757"
  }
}
```

**Success Response (201 Created):**
```json
{
  "message": "Category created successfully",
  "data": {
    "category": {
      "id": 1,
      "user_id": 1,
      "name": "work",
      "color": "#ff4757",
      "todo_count": 0,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

**Note:** Category names are normalized to lowercase before saving (e.g., "Work" becomes "work").

**Error Response (422 Unprocessable Entity):**
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "details": {
      "validation_errors": {
        "name": ["required", "notblank"],
        "color": ["required", "hexcolor"]
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
    "message": "Category with this name already exists"
  }
}
```

### Update Category

Update an existing category.

**Endpoint:** `PATCH /api/v1/categories/:id`

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "category": {
    "name": "Personal Projects",
    "color": "#2ed573"
  }
}
```

Both fields are optional for partial updates.

**Success Response (200 OK):**
```json
{
  "category": {
    "id": 1,
    "user_id": 1,
    "name": "personal projects",
    "color": "#2ed573",
    "todo_count": 5,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**Error Response (404 Not Found):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Category with id 123 not found"
  }
}
```

### Delete Category

Delete a category. All todos assigned to this category will have their category_id set to null.

**Endpoint:** `DELETE /api/v1/categories/:id`

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
    "message": "Category with id 123 not found"
  }
}
```

## Category Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | Integer | Read-only | Unique identifier |
| `user_id` | Integer | Read-only | Owner of the category |
| `name` | String | Yes | Category name (unique per user, stored lowercase) |
| `color` | String | Yes | Hex color code (e.g., "#ff4757") |
| `todo_count` | Integer | Read-only | Number of todos in this category |
| `created_at` | String (RFC3339) | Read-only | Creation timestamp |
| `updated_at` | String (RFC3339) | Read-only | Last update timestamp |

## Validation Rules

### Name
- **Required**: Cannot be blank
- **Not Blank**: Cannot be only whitespace
- **Uniqueness**: Must be unique per user (case-insensitive)
- **Length**: Maximum 50 characters
- **Normalization**: Trimmed and converted to lowercase before saving

### Color
- **Required**: Cannot be blank
- **Format**: Must be a valid hex color code (e.g., "#ff4757")
- **Pattern**: Must match `/^#[0-9a-fA-F]{6}$/`

## Business Rules

1. **User Scoped**: Users can only see and manage their own categories
2. **Unique Names**: Category names must be unique within a user's categories (case-insensitive due to lowercase normalization)
3. **Todo Count Sync**: `todo_count` is automatically updated when todos are created, updated, or deleted
4. **Cascade Behavior**: When a category is deleted, all todos assigned to it have their `category_id` set to `null`
5. **No Default Category**: Todos can exist without a category

## Error Handling

All endpoints return consistent error responses. Common error scenarios:

- **401 Unauthorized**: Missing or invalid JWT token
- **404 Not Found**: Category doesn't exist or doesn't belong to the user
- **409 Conflict**: Duplicate category name
- **422 Unprocessable Entity**: Validation errors (invalid color format, blank name)

## Performance Considerations

1. **Counter Cache**: The `todo_count` field is maintained through increment/decrement operations when todos change categories.

2. **User Scoping**: All queries are scoped to the authenticated user, ensuring data isolation.

3. **Indexing**: Composite unique index on `(user_id, name)` ensures fast lookups and prevents duplicates at the database level.
