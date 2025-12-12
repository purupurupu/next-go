# Tags API

The Tags API allows users to create and manage tags for organizing their todos. Tags provide a flexible way to categorize and filter todos with a many-to-many relationship.

## Overview

- **Base URL**: `/api/v1/tags`
- **Authentication**: Required (JWT token in Authorization header)
- **Content-Type**: `application/json`
- **User Scope**: All tag operations are scoped to the authenticated user

## Tag Object

### Attributes

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| `id` | integer | Unique identifier | Auto-generated |
| `name` | string | Tag name | Required, 1-30 characters, unique per user |
| `color` | string | Hex color code | Required, format: `#RRGGBB` |
| `created_at` | string | ISO 8601 timestamp | Auto-generated |
| `updated_at` | string | ISO 8601 timestamp | Auto-updated |

### Example Tag Object

```json
{
  "id": 1,
  "name": "urgent",
  "color": "#EF4444",
  "created_at": "2024-01-15T10:30:00.000Z",
  "updated_at": "2024-01-15T10:30:00.000Z"
}
```

## Endpoints

### List Tags

Get all tags for the authenticated user.

```
GET /api/v1/tags
```

#### Response

```json
{
  "tags": [
    {
      "id": 1,
      "name": "urgent",
      "color": "#EF4444",
      "created_at": "2024-01-15T10:30:00.000Z",
      "updated_at": "2024-01-15T10:30:00.000Z"
    },
    {
      "id": 2,
      "name": "work",
      "color": "#3B82F6",
      "created_at": "2024-01-15T10:31:00.000Z",
      "updated_at": "2024-01-15T10:31:00.000Z"
    }
  ]
}
```

**Notes**:
- Tags are returned ordered by name (ascending)
- Returns only tags belonging to the authenticated user

### Get Tag

Get a specific tag by ID.

```
GET /api/v1/tags/:id
```

#### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | integer | Tag ID |

#### Response

```json
{
  "tag": {
    "id": 1,
    "name": "urgent",
    "color": "#EF4444",
    "created_at": "2024-01-15T10:30:00.000Z",
    "updated_at": "2024-01-15T10:30:00.000Z"
  }
}
```

#### Error Responses

- `404 Not Found` - Tag not found or belongs to another user

### Create Tag

Create a new tag.

```
POST /api/v1/tags
```

#### Request Body

```json
{
  "tag": {
    "name": "important",
    "color": "#F59E0B"
  }
}
```

#### Parameters

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `name` | string | Yes | Tag name | 1-30 characters, trimmed |
| `color` | string | No | Hex color | Default: `#6B7280`, format: `#RRGGBB` |

#### Response

```json
{
  "tag": {
    "id": 3,
    "name": "important",
    "color": "#F59E0B",
    "created_at": "2024-01-15T10:35:00.000Z",
    "updated_at": "2024-01-15T10:35:00.000Z"
  }
}
```

#### Error Responses

- `422 Unprocessable Entity` - Validation errors

```json
{
  "errors": {
    "name": ["has already been taken"],
    "color": ["is invalid"]
  }
}
```

**Notes**:
- Tag names are normalized (trimmed of whitespace)
- Color values are normalized to uppercase
- Names must be unique per user (case-insensitive)

### Update Tag

Update an existing tag.

```
PATCH /api/v1/tags/:id
PUT /api/v1/tags/:id
```

#### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | integer | Tag ID |

#### Request Body

```json
{
  "tag": {
    "name": "high-priority",
    "color": "#DC2626"
  }
}
```

#### Response

```json
{
  "tag": {
    "id": 1,
    "name": "high-priority",
    "color": "#DC2626",
    "created_at": "2024-01-15T10:30:00.000Z",
    "updated_at": "2024-01-15T10:40:00.000Z"
  }
}
```

#### Error Responses

- `404 Not Found` - Tag not found or belongs to another user
- `422 Unprocessable Entity` - Validation errors

### Delete Tag

Delete a tag. This will also remove the tag from all associated todos.

```
DELETE /api/v1/tags/:id
```

#### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | integer | Tag ID |

#### Response

```
204 No Content
```

#### Error Responses

- `404 Not Found` - Tag not found or belongs to another user

**Notes**:
- Deleting a tag removes it from all todos but does not delete the todos themselves
- This operation cannot be undone

## Tag Search

Search for tags by name (useful for autocomplete).

```
GET /api/v1/tags?search=work
```

#### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `search` | string | Search term for tag names |

#### Response

Returns tags where the name contains the search term (case-insensitive).

## Using Tags with Todos

### Adding Tags to Todos

When creating or updating todos, include tag IDs:

```json
{
  "todo": {
    "title": "Complete project",
    "tag_ids": [1, 2, 3]
  }
}
```

### Filtering Todos by Tag

```
GET /api/v1/todos?tag_id=1
```

Returns all todos that have the specified tag.

## Best Practices

1. **Limit Tags per User**: Consider implementing a maximum number of tags per user to prevent abuse
2. **Tag Colors**: Use a consistent color palette for better UX
3. **Tag Names**: Keep tag names short and descriptive
4. **Cleanup**: Periodically review and remove unused tags

## Rate Limiting

Tag operations follow the same rate limiting rules as other API endpoints. See the main API documentation for details.