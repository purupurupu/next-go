# Todo History API

## Overview

The Todo History API provides read-only access to the change history of todos. Every significant change to a todo is automatically tracked, creating an audit trail of modifications.

## Implementation Status

| Backend | Status |
|---------|--------|
| Go (Echo) | ✅ Implemented |

## Authentication Required

All history endpoints require a valid JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

## Endpoints

### List Todo History

Get all history entries for a specific todo with pagination support.

**Endpoint:** `GET /api/v1/todos/:todo_id/histories`

**URL Parameters:**
- `todo_id` (required): ID of the todo

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 20, max: 100)

**Success Response (200 OK):**
```json
{
  "histories": [
    {
      "id": 1,
      "todo_id": 1,
      "action": "created",
      "changes": {
        "title": "Complete project documentation",
        "priority": 2,
        "status": 0
      },
      "user": {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com"
      },
      "created_at": "2024-01-01T10:00:00Z",
      "human_readable_change": "Todoが作成されました"
    },
    {
      "id": 2,
      "todo_id": 1,
      "action": "updated",
      "changes": {
        "title": ["Complete project documentation", "Complete API documentation"],
        "description": [null, "Write comprehensive API docs with examples"]
      },
      "user": {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com"
      },
      "created_at": "2024-01-01T11:00:00Z",
      "human_readable_change": "タイトルが「Complete project documentation」から「Complete API documentation」に変更されました"
    },
    {
      "id": 3,
      "todo_id": 1,
      "action": "status_changed",
      "changes": {
        "status": [0, 1]
      },
      "user": {
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com"
      },
      "created_at": "2024-01-01T12:00:00Z",
      "human_readable_change": "ステータスが「未着手」から「進行中」に変更されました"
    },
    {
      "id": 4,
      "todo_id": 1,
      "action": "priority_changed",
      "changes": {
        "priority": [2, 1]
      },
      "user": {
        "id": 2,
        "name": "Jane Smith",
        "email": "jane@example.com"
      },
      "created_at": "2024-01-02T10:00:00Z",
      "human_readable_change": "優先度が「高」から「中」に変更されました"
    }
  ],
  "meta": {
    "total": 4,
    "current_page": 1,
    "total_pages": 1,
    "per_page": 20
  }
}
```

**Notes:**
- History entries are returned in reverse chronological order (newest first)
- Empty array `[]` in histories if no history exists
- `human_readable_change` provides a Japanese description of the change
- Pagination metadata is included in `meta` object

## History Entry Structure

### Fields

- `id`: Unique identifier for the history entry
- `action`: Type of action performed
- `changes`: Object containing the changed fields and their values
- `user`: User who made the change
- `created_at`: Timestamp when the change was made
- `human_readable_change`: Human-readable description in Japanese

### Action Types

| Action | Description | Changes Format |
|--------|-------------|----------------|
| `created` | Todo was created | Object with initial values |
| `updated` | Todo was updated | Object with [old_value, new_value] arrays |
| `deleted` | Todo was deleted | Object with final values |
| `status_changed` | Status was specifically changed | `{ status: [old, new] }` |
| `priority_changed` | Priority was specifically changed | `{ priority: [old, new] }` |

### Changes Object Format

For **created** actions:
```json
{
  "title": "New todo",
  "priority": "medium",
  "status": "pending"
}
```

For **updated** actions (shows before/after values):
```json
{
  "title": ["Old title", "New title"],
  "completed": [false, true],
  "due_date": [null, "2024-12-31"]
}
```

For **deleted** actions:
```json
{
  "title": "Deleted todo",
  "completed": true
}
```

## Tracked Fields

The following todo fields are tracked for changes:
- `title`
- `completed`
- `priority`
- `status`
- `description`
- `due_date`
- `category_id`
- `tag_ids`

## Human-Readable Descriptions

The `human_readable_change` field provides user-friendly descriptions in Japanese:

| Change Type | Example Description |
|-------------|-------------------|
| Created | タスクが作成されました |
| Title changed | タイトルが「{old}」から「{new}」に変更されました |
| Status changed | ステータスが「未着手」から「進行中」に変更されました |
| Priority changed | 優先度が「高」から「中」に変更されました |
| Completed | タスクが完了しました |
| Uncompleted | タスクが未完了に戻されました |
| Multiple changes | タイトル、ステータス、優先度が変更されました |

## Frontend Integration Example

```typescript
// Using ApiClient pattern from the project
class TodoHistoryApiClient extends ApiClient {
  async getHistory(todoId: number, page = 1, perPage = 20) {
    return this.get<HistoryListResponse>(
      `/todos/${todoId}/histories?page=${page}&per_page=${perPage}`
    );
  }
}

interface HistoryListResponse {
  histories: TodoHistory[];
  meta: {
    total: number;
    current_page: number;
    total_pages: number;
    per_page: number;
  };
}

// Usage in React component
function TodoHistory({ todoId }) {
  const [history, setHistory] = useState([]);
  
  useEffect(() => {
    const fetchHistory = async () => {
      const client = new TodoHistoryApiClient(httpClient);
      const entries = await client.getHistory(todoId);
      setHistory(entries);
    };
    
    fetchHistory();
  }, [todoId]);
  
  return (
    <div>
      {history.map(entry => (
        <div key={entry.id}>
          <p>{entry.human_readable_change}</p>
          <small>
            {entry.user.name} - {formatDate(entry.created_at)}
          </small>
        </div>
      ))}
    </div>
  );
}
```

## Performance Considerations

1. **Automatic Tracking**: History is created automatically via TodoService layer
2. **Read-only Access**: History entries cannot be modified or deleted via API
3. **Efficient Storage**: Only changed fields are stored in the database as JSONB
4. **User Context**: User ID is passed through JWT authentication context
5. **Pagination**: Built-in pagination support for todos with extensive history

## Notes

- History tracking is automatic and cannot be disabled
- All history entries are permanent and cannot be deleted
- The system tracks who made each change for accountability
- No duplicate history is created when updating with same values (detectChanges logic)