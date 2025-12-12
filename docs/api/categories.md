# Categories API

## Overview

Categories provide a way for users to organize their todos. Each category has a name and color for visual organization, and todos can optionally be assigned to categories.

## Base URL

All endpoints are prefixed with `/api/v1`:
```
http://localhost:3001/api/v1/categories
```

## Endpoints

### Get Categories

Retrieve all categories for the authenticated user.

**Endpoint:** `GET /api/v1/categories`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "name": "Work",
    "color": "#ff4757",
    "todos_count": 5,
    "user_id": 1,
    "created_at": "2024-01-01T00:00:00.000Z",
    "updated_at": "2024-01-01T00:00:00.000Z"
  },
  {
    "id": 2,
    "name": "Personal",
    "color": "#3742fa",
    "todos_count": 3,
    "user_id": 1,
    "created_at": "2024-01-01T00:00:00.000Z",
    "updated_at": "2024-01-01T00:00:00.000Z"
  }
]
```

### Get Category

Retrieve a specific category.

**Endpoint:** `GET /api/v1/categories/:id`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200 OK):**
```json
{
  "id": 1,
  "name": "Work",
  "color": "#ff4757",
  "todos_count": 5,
  "user_id": 1,
  "created_at": "2024-01-01T00:00:00.000Z",
  "updated_at": "2024-01-01T00:00:00.000Z"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Record not found"
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
    "id": 1,
    "name": "Work",
    "color": "#ff4757",
    "todos_count": 0,
    "user_id": 1,
    "created_at": "2024-01-01T00:00:00.000Z",
    "updated_at": "2024-01-01T00:00:00.000Z"
  }
}
```

**Error Response (422 Unprocessable Entity):**
```json
{
  "errors": {
    "name": ["can't be blank", "has already been taken"],
    "color": ["can't be blank", "must be a valid hex color"]
  }
}
```

### Update Category

Update an existing category.

**Endpoint:** `PUT /api/v1/categories/:id` or `PATCH /api/v1/categories/:id`

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

**Success Response (200 OK):**
```json
{
  "message": "Category updated successfully",
  "data": {
    "id": 1,
    "name": "Personal Projects",
    "color": "#2ed573",
    "todos_count": 5,
    "user_id": 1,
    "created_at": "2024-01-01T00:00:00.000Z",
    "updated_at": "2024-01-01T12:00:00.000Z"
  }
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Record not found"
}
```

**Error Response (422 Unprocessable Entity):**
```json
{
  "errors": {
    "name": ["has already been taken"],
    "color": ["must be a valid hex color"]
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

**Success Response (200 OK):**
```json
{
  "message": "Category deleted successfully"
}
```

**Error Response (404 Not Found):**
```json
{
  "error": "Record not found"
}
```

## Category Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `id` | Integer | Read-only | Unique identifier |
| `name` | String | Yes | Category name (unique per user) |
| `color` | String | Yes | Hex color code (e.g., "#ff4757") |
| `todos_count` | Integer | Read-only | Number of todos in this category (counter cache) |
| `user_id` | Integer | Read-only | Owner of the category |
| `created_at` | String (ISO 8601) | Read-only | Creation timestamp |
| `updated_at` | String (ISO 8601) | Read-only | Last update timestamp |

## Validation Rules

### Name
- **Required**: Cannot be blank
- **Uniqueness**: Must be unique per user
- **Length**: Maximum 50 characters

### Color
- **Required**: Cannot be blank
- **Format**: Must be a valid hex color code (e.g., "#ff4757")
- **Pattern**: Must match `/^#[0-9a-fA-F]{6}$/`

## Business Rules

1. **User Scoped**: Users can only see and manage their own categories
2. **Unique Names**: Category names must be unique within a user's categories
3. **Counter Cache**: `todos_count` is automatically maintained and updated when todos are assigned/unassigned
4. **Cascade Behavior**: When a category is deleted, all todos assigned to it have their `category_id` set to `null`
5. **Default Category**: There is no default "uncategorized" category - todos can exist without a category

## Usage Examples

### Frontend Integration

```javascript
// Category API Client
class CategoryApiClient {
  async getCategories() {
    const response = await fetch('/api/v1/categories', {
      headers: { 'Authorization': `Bearer ${this.getToken()}` }
    });
    return response.json();
  }
  
  async createCategory(categoryData) {
    const response = await fetch('/api/v1/categories', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.getToken()}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ category: categoryData })
    });
    const data = await response.json();
    if (response.ok) return data;
    throw new Error(data.error || 'Failed to create category');
  }
  
  async updateCategory(id, categoryData) {
    const response = await fetch(`/api/v1/categories/${id}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${this.getToken()}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ category: categoryData })
    });
    const data = await response.json();
    if (response.ok) return data;
    throw new Error(data.error || 'Failed to update category');
  }
  
  async deleteCategory(id) {
    const response = await fetch(`/api/v1/categories/${id}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${this.getToken()}` }
    });
    if (!response.ok) {
      const data = await response.json();
      throw new Error(data.error || 'Failed to delete category');
    }
  }
}
```

### Color Validation

Categories use hex color codes for visual organization. Common color options:

```javascript
const CATEGORY_COLORS = [
  '#ff4757', // Red
  '#ff6b6b', // Light Red
  '#ff9f43', // Orange
  '#feca57', // Yellow
  '#48dbfb', // Light Blue
  '#3742fa', // Blue
  '#2f3542', // Dark
  '#57606f', // Gray
  '#2ed573', // Green
  '#5f27cd', // Purple
  '#00d2d3', // Cyan
  '#ff3838'  // Bright Red
];
```

## Error Handling

All endpoints return consistent error responses as documented in the [API Overview](./README.md). Common error scenarios:

- **401 Unauthorized**: Missing or invalid JWT token
- **404 Not Found**: Category doesn't exist or doesn't belong to the user
- **422 Unprocessable Entity**: Validation errors (duplicate name, invalid color format)

## Performance Considerations

1. **Counter Cache**: The `todos_count` field is automatically maintained using Rails counter cache, eliminating N+1 queries when loading categories with todo counts.

2. **User Scoping**: All queries are scoped to the authenticated user, ensuring data isolation and optimal query performance.

3. **Indexing**: Database indexes on `user_id` and unique constraint on `(user_id, name)` ensure fast lookups and prevent duplicates.