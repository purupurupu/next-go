# Todos API

## Overview

The Todos API provides CRUD operations for managing todo items. All endpoints require authentication and return only the authenticated user's todos.

## Authentication Required

All todo endpoints require a valid JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Base URL

All endpoints are prefixed with `/api/v1`:
```
http://localhost:3001/api/v1/todos
```

## Endpoints

### List Todos

Get all todos for the authenticated user.

**Endpoint:** `GET /api/v1/todos`

**Query Parameters:** None

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "title": "Complete project documentation",
    "completed": false,
    "position": 0,
    "priority": "high",
    "status": "in_progress",
    "description": "Write comprehensive API documentation with examples",
    "due_date": "2024-12-31",
    "category": {
      "id": 1,
      "name": "Work",
      "color": "#3B82F6"
    },
    "tags": [],
    "files": [],
    "comments_count": 2,
    "latest_comments": [],
    "history_count": 5,
    "created_at": "2024-01-01T00:00:00.000Z",
    "updated_at": "2024-01-01T00:00:00.000Z"
  },
  {
    "id": 2,
    "title": "Review pull requests",
    "completed": true,
    "position": 1,
    "priority": "medium",
    "status": "completed",
    "description": null,
    "due_date": null,
    "category": null,
    "tags": [
      {
        "id": 1,
        "name": "urgent",
        "color": "#EF4444"
      }
    ],
    "files": [
      {
        "id": 123,
        "filename": "code_review.pdf",
        "content_type": "application/pdf",
        "byte_size": 204800,
        "url": "http://localhost:9000/todo-files/..."
      }
    ],
    "comments_count": 0,
    "latest_comments": [],
    "history_count": 3,
    "created_at": "2024-01-01T00:00:00.000Z",
    "updated_at": "2024-01-02T00:00:00.000Z"
  }
]
```

**Notes:**
- Todos are returned ordered by `position`
- Empty array `[]` if no todos exist
- `comments_count` shows the total number of comments on the todo
- `latest_comments` may contain recent comments for preview (currently empty)
- `history_count` shows the total number of change history entries

### Get Single Todo

Get a specific todo by ID.

**Endpoint:** `GET /api/v1/todos/:id`

**URL Parameters:**
- `id` (required): Todo ID

**Success Response (200 OK):**
```json
{
  "id": 1,
  "title": "Complete project documentation",
  "completed": false,
  "position": 0,
  "priority": "high",
  "status": "in_progress",
  "description": "Write comprehensive API documentation with examples",
  "due_date": "2024-12-31",
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

### Create Todo

Create a new todo item.

**Endpoint:** `POST /api/v1/todos`

**Request Body:**
```json
{
  "title": "New task",
  "priority": "high",
  "status": "pending",
  "description": "Detailed task description",
  "due_date": "2024-12-31",
  "category_id": 2,
  "tag_ids": [1, 3]
}
```

**Parameters:**
- `title` (required): Todo description
- `priority` (optional): Priority level - `"low"`, `"medium"`, `"high"`. Defaults to `"medium"`
- `status` (optional): Task status - `"pending"`, `"in_progress"`, `"completed"`. Defaults to `"pending"`
- `description` (optional): Detailed description of the task
- `due_date` (optional): Due date in YYYY-MM-DD format
- `category_id` (optional): ID of the category to assign this todo to
- `tag_ids` (optional): Array of tag IDs to assign to this todo
- `files` (optional): File attachments (use multipart/form-data for file uploads)
- `completed` (optional): Defaults to `false`

**Success Response (201 Created):**
```json
{
  "id": 3,
  "title": "New task",
  "completed": false,
  "position": 2,
  "priority": "high",
  "status": "pending",
  "description": "Detailed task description",
  "due_date": "2024-12-31",
  "created_at": "2024-01-03T00:00:00.000Z",
  "updated_at": "2024-01-03T00:00:00.000Z"
}
```

**Error Response (422 Unprocessable Entity):**
```json
{
  "errors": {
    "title": ["can't be blank"],
    "priority": ["is not included in the list"],
    "status": ["is not included in the list"],
    "due_date": ["must be in the future"]
  }
}
```

### Update Todo

Update an existing todo.

**Endpoint:** `PUT /api/v1/todos/:id` or `PATCH /api/v1/todos/:id`

**URL Parameters:**
- `id` (required): Todo ID

**Request Body:**
```json
{
  "title": "Updated task",
  "completed": true,
  "priority": "low",
  "status": "completed",
  "description": "Updated description",
  "due_date": "2024-12-31",
  "category_id": 3,
  "tag_ids": [2, 4]
}
```

**Parameters:**
- `title` (optional): New title
- `completed` (optional): Completion status
- `priority` (optional): Priority level - `"low"`, `"medium"`, `"high"`
- `status` (optional): Task status - `"pending"`, `"in_progress"`, `"completed"`
- `description` (optional): Updated description
- `due_date` (optional): New due date
- `category_id` (optional): ID of the category to assign (use null to remove category)
- `tag_ids` (optional): Array of tag IDs to assign (empty array to remove all tags)
- `files` (optional): New file attachments (use multipart/form-data)

**Success Response (200 OK):**
```json
{
  "id": 1,
  "title": "Updated task",
  "completed": true,
  "position": 0,
  "priority": "low",
  "status": "completed",
  "description": "Updated description",
  "due_date": "2024-12-31",
  "created_at": "2024-01-01T00:00:00.000Z",
  "updated_at": "2024-01-03T00:00:00.000Z"
}
```

### Delete Todo

Delete a todo item.

**Endpoint:** `DELETE /api/v1/todos/:id`

**URL Parameters:**
- `id` (required): Todo ID

**Success Response (204 No Content):**
No response body

**Error Response (404 Not Found):**
```json
{
  "error": "Record not found"
}
```

### Search Todos

Search and filter todos with advanced filtering options.

**Endpoint:** `GET /api/v1/todos/search`

**Query Parameters:**
- `q` (optional): Search query for title and description
- `category_id` (optional): Filter by category ID. Use `-1` for uncategorized todos
- `status` (optional): Filter by status. Can be single value or array
- `priority` (optional): Filter by priority. Can be single value or array
- `tag_ids[]` (optional): Filter by tag IDs (array)
- `tag_mode` (optional): Tag matching mode - `"any"` (default) or `"all"`
- `due_date_from` (optional): Filter todos with due date from this date (YYYY-MM-DD)
- `due_date_to` (optional): Filter todos with due date until this date (YYYY-MM-DD)
- `sort_by` (optional): Sort field - `"position"` (default), `"created_at"`, `"updated_at"`, `"due_date"`, `"title"`, `"priority"`, `"status"`
- `sort_order` (optional): Sort direction - `"asc"` (default) or `"desc"`
- `page` (optional): Page number for pagination (default: 1)
- `per_page` (optional): Items per page (default: 20, max: 100)

**Example Request:**
```
GET /api/v1/todos/search?q=documentation&status[]=pending&status[]=in_progress&priority=high&tag_ids[]=1&tag_ids[]=2&tag_mode=all&sort_by=due_date&sort_order=asc&page=1&per_page=20
```

**Success Response (200 OK):**
```json
{
  "todos": [
    {
      "id": 1,
      "title": "Complete project documentation",
      "completed": false,
      "position": 0,
      "priority": "high",
      "status": "in_progress",
      "description": "Write comprehensive API documentation with examples",
      "due_date": "2024-12-31",
      "category": {
        "id": 1,
        "name": "Work",
        "color": "#3B82F6"
      },
      "tags": [
        {
          "id": 1,
          "name": "urgent",
          "color": "#EF4444"
        }
      ],
      "files": [],
      "comments_count": 2,
      "latest_comments": [],
      "history_count": 5,
      "created_at": "2024-01-01T00:00:00.000Z",
      "updated_at": "2024-01-01T00:00:00.000Z",
      "highlights": {
        "title": [
          {
            "start": 17,
            "end": 30,
            "matched_text": "documentation"
          }
        ],
        "description": [
          {
            "start": 20,
            "end": 33,
            "matched_text": "documentation"
          }
        ]
      }
    }
  ],
  "meta": {
    "total": 42,
    "current_page": 1,
    "total_pages": 3,
    "per_page": 20,
    "search_query": "documentation",
    "filters_applied": {
      "search": "documentation",
      "status": ["pending", "in_progress"],
      "priority": ["high"],
      "tag_ids": [1, 2]
    }
  },
  "suggestions": [
    {
      "type": "spelling",
      "message": "検索キーワードのスペルを確認してください。"
    },
    {
      "type": "reduce_filters",
      "message": "フィルター条件を減らしてみてください。",
      "current_filters": ["search", "status", "priority", "tag_ids"]
    }
  ]
}
```

**Response Fields:**
- `todos`: Array of todo items matching the search criteria
- `highlights`: Contains match positions for search query in title and description
- `meta`: Pagination and search metadata
  - `total`: Total number of matching todos
  - `current_page`: Current page number
  - `total_pages`: Total number of pages
  - `per_page`: Number of items per page
  - `search_query`: The search query used
  - `filters_applied`: Active filters summary
- `suggestions`: Array of suggestions when no results found (optional)

**Notes:**
- Search is case-insensitive and matches partial words
- Multiple status/priority values create an OR condition
- Tag filtering supports both ANY (match any tag) and ALL (match all tags) modes
- Results include highlight information for search query matches
- Empty results include helpful suggestions

### Update Todo Order

Bulk update todo positions for drag-and-drop reordering.

**Endpoint:** `PATCH /api/v1/todos/update_order`

**Request Body:**
```json
{
  "todos": [
    { "id": 3, "position": 0 },
    { "id": 1, "position": 1 },
    { "id": 2, "position": 2 }
  ]
}
```

**Success Response (200 OK):**
```json
{
  "message": "Order updated successfully"
}
```

**Error Response (422 Unprocessable Entity):**
```json
{
  "error": "Invalid todo IDs"
}
```

**Notes:**
- All todos must belong to the authenticated user
- Invalid IDs will cause the entire operation to fail
- Positions should be sequential starting from 0
- Updates are performed in a transaction for data consistency

### Update Todo Tags

Update tags for a specific todo.

**Endpoint:** `PATCH /api/v1/todos/:id/tags`

**Request Body:**
```json
{
  "tag_ids": [1, 2, 3]
}
```

**Success Response (200 OK):**
Returns the updated todo with new tags.

**Error Response (422 Unprocessable Entity):**
```json
{
  "error": "Invalid tag IDs"
}
```

**Notes:**
- All tags must belong to the authenticated user
- Empty array removes all tags from the todo
- Invalid tag IDs will cause the operation to fail

### File Attachments

Todos support multiple file attachments. See [Todo File Uploads API](./todos-file-uploads.md) for detailed documentation.

**Delete File Attachment:** `DELETE /api/v1/todos/:todo_id/files/:file_id`

Removes a specific file from a todo.

**Success Response (200 OK):**
Returns the updated todo without the deleted file.

**Error Response (404 Not Found):**
```json
{
  "error": "File not found"
}
```

### Comments

Todos support commenting functionality. See [Comments API](./comments.md) for detailed documentation.

**Endpoints:**
- `GET /api/v1/todos/:todo_id/comments` - List all comments for a todo
- `POST /api/v1/todos/:todo_id/comments` - Create a new comment
- `PUT /api/v1/todos/:todo_id/comments/:id` - Update a comment
- `DELETE /api/v1/todos/:todo_id/comments/:id` - Soft delete a comment

### History

Todo changes are automatically tracked. See [Todo History API](./todo-histories.md) for detailed documentation.

**Endpoints:**
- `GET /api/v1/todos/:todo_id/histories` - List all history entries for a todo

**Tracked Actions:**
- Todo creation
- Todo updates (title, status, priority, etc.)
- Todo deletion
- Status changes
- Priority changes

## Data Validation

### Title
- Required field
- Cannot be empty
- Maximum length: 255 characters

### Priority
- Optional field
- Allowed values: `"low"`, `"medium"`, `"high"`
- Defaults to `"medium"` on creation
- Used for task prioritization and visual indicators

### Status
- Optional field
- Allowed values: `"pending"`, `"in_progress"`, `"completed"`
- Defaults to `"pending"` on creation
- Tracks task progress workflow

### Description
- Optional field
- Text field for detailed task information
- No length limit (database TEXT type)
- Can be `null` or empty string

### Due Date
- Optional field
- Format: YYYY-MM-DD
- Must be today or in the future (on creation)
- Can be any date on update

### Position
- Automatically assigned on creation
- Should be unique among user's todos
- Used for ordering in the UI

## Filtering and Sorting

### Basic List (GET /api/v1/todos)
- Returns all todos ordered by position
- No server-side filtering available
- Suitable for small todo lists

### Advanced Search (GET /api/v1/todos/search)
Use the search endpoint for:
- **Full-text search**: Search in title and description
- **Category filtering**: Filter by category or uncategorized todos
- **Status filtering**: Filter by `"pending"`, `"in_progress"`, `"completed"`
- **Priority filtering**: Filter by `"low"`, `"medium"`, `"high"`
- **Tag filtering**: Filter by multiple tags with AND/OR logic
- **Date range filtering**: Filter by due date range
- **Custom sorting**: Sort by various fields in ascending/descending order
- **Pagination**: Handle large result sets efficiently

### Performance Tips
- Use the search endpoint when you need filtering or custom sorting
- The basic list endpoint is faster for displaying all todos without filters
- Search results include highlight information for better UX
- Implement debouncing for real-time search to reduce API calls

## Frontend Integration Example

```javascript
class TodoApiClient {
  constructor(token) {
    this.token = token;
    this.baseURL = '/api/v1/todos';
  }

  async getAll() {
    const response = await fetch(this.baseURL, {
      headers: {
        'Authorization': `Bearer ${this.token}`
      }
    });
    return response.json();
  }

  async create(todoData) {
    // todoData can include: title, priority, status, description, due_date
    const response = await fetch(this.baseURL, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(todoData)
    });
    return response.json();
  }

  async updateOrder(todos) {
    const response = await fetch(`${this.baseURL}/update_order`, {
      method: 'PATCH',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ todos })
    });
    return response.json();
  }
}
```

## Performance Considerations

1. **Batch Operations**: Use `update_order` for bulk position updates
2. **Caching**: Consider caching todo list on frontend
3. **Optimistic Updates**: Update UI before API confirmation
4. **Pagination**: Use the search endpoint with `page` and `per_page` parameters
5. **Search Debouncing**: Implement debouncing (300-500ms) for real-time search
6. **Efficient Filtering**: Use server-side filtering via search endpoint for large datasets
7. **Indexes**: The database has indexes on searchable fields for optimal performance