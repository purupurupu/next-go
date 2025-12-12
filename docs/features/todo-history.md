# Todo History Feature

## Overview

The todo history feature provides a complete audit trail of all changes made to todos. Every creation, update, and deletion is tracked automatically, providing transparency and accountability for task management.

## Features

### 1. Automatic Tracking
- All todo changes are recorded automatically
- No user action required
- Cannot be disabled or bypassed

### 2. Tracked Actions
- **Created**: Initial todo creation
- **Updated**: Any field changes
- **Deleted**: Todo deletion
- **Status Changed**: Specific tracking for status transitions
- **Priority Changed**: Specific tracking for priority changes

### 3. Change Details
- Stores before/after values for all fields
- Tracks which user made the change
- Precise timestamps for each action
- JSON storage for flexible field tracking

## Technical Implementation

### Backend

#### Model Structure
```ruby
class TodoHistory < ApplicationRecord
  belongs_to :todo
  belongs_to :user
  
  enum action: {
    created: 0,
    updated: 1,
    deleted: 2,
    status_changed: 3,
    priority_changed: 4
  }
  
  # Changes stored as JSONB
  # Format: { field_name: { from: old_value, to: new_value } }
end

class Todo < ApplicationRecord
  has_many :todo_histories, dependent: :destroy
  
  after_create :track_creation
  after_update :track_update
  after_destroy :track_deletion
  
  private
  
  def track_creation
    todo_histories.create!(
      user: Current.user,
      action: 'created',
      changes: attributes.slice('title', 'status', 'priority', 'due_date')
    )
  end
  
  def track_update
    changes_to_track = saved_changes.except('updated_at', 'position')
    return if changes_to_track.empty?
    
    action = determine_action(changes_to_track)
    
    todo_histories.create!(
      user: Current.user,
      action: action,
      changes: format_changes(changes_to_track)
    )
  end
end
```

#### API Endpoints

**Base URL**: `/api/v1/todos/:todo_id/histories`

1. **List History**
   ```
   GET /api/v1/todos/:todo_id/histories
   ```
   
   Response:
   ```json
   {
     "histories": [
       {
         "id": 1,
         "action": "created",
         "changes": {
           "title": { "to": "New task" },
           "status": { "to": "pending" },
           "priority": { "to": "medium" }
         },
         "user": {
           "id": 1,
           "name": "John Doe",
           "email": "john@example.com"
         },
         "created_at": "2024-01-15T10:30:00Z"
       },
       {
         "id": 2,
         "action": "status_changed",
         "changes": {
           "status": { "from": "pending", "to": "in_progress" }
         },
         "user": {
           "id": 1,
           "name": "John Doe",
           "email": "john@example.com"
         },
         "created_at": "2024-01-15T11:00:00Z"
       }
     ]
   }
   ```

### Frontend

#### Components

1. **TodoHistory**
   - Main container for history display
   - Fetches and displays history entries
   - Collapsible panel design

2. **HistoryTimeline**
   - Visual timeline of changes
   - Groups changes by date
   - Shows user avatars

3. **HistoryEntry**
   - Individual history item
   - Formats changes for display
   - Color-coded by action type

4. **ChangeDetail**
   - Shows field-level changes
   - Before/after comparison
   - Human-readable formatting

#### Usage Example

```typescript
// In TodoDetail component
const [showHistory, setShowHistory] = useState(false);
const { histories, loading } = useTodoHistory(todo.id);

<Collapsible open={showHistory} onOpenChange={setShowHistory}>
  <CollapsibleTrigger>
    <History className="h-4 w-4" />
    View History ({todo.history_count})
  </CollapsibleTrigger>
  <CollapsibleContent>
    <TodoHistory histories={histories} loading={loading} />
  </CollapsibleContent>
</Collapsible>

// API Integration
const historyApi = new TodoHistoryApiClient(httpClient);
const histories = await historyApi.list(todoId);
```

## Change Formatting

### Field Names
- `title` → "Title"
- `status` → "Status"
- `priority` → "Priority"
- `due_date` → "Due Date"
- `completed` → "Completed"
- `description` → "Description"
- `category_id` → "Category"
- `tag_ids` → "Tags"

### Value Formatting
- **Status**: `pending` → "Pending", `in_progress` → "In Progress", `completed` → "Completed"
- **Priority**: `low` → "Low", `medium` → "Medium", `high` → "High"
- **Dates**: ISO format → "Jan 15, 2024"
- **Boolean**: `true` → "Yes", `false` → "No"
- **Arrays**: Show added/removed items

## Database Schema

```sql
CREATE TABLE todo_histories (
  id bigserial PRIMARY KEY,
  todo_id bigint NOT NULL REFERENCES todos(id),
  user_id bigint NOT NULL REFERENCES users(id),
  action integer NOT NULL,
  changes jsonb DEFAULT '{}',
  created_at timestamp(6) NOT NULL,
  updated_at timestamp(6) NOT NULL
);

CREATE INDEX index_todo_histories_on_todo_id ON todo_histories(todo_id);
CREATE INDEX index_todo_histories_on_user_id ON todo_histories(user_id);
CREATE INDEX index_todo_histories_on_action ON todo_histories(action);
```

## Display Examples

### Creation Entry
```
John Doe created this todo
• Title: "Complete project documentation"
• Status: Pending
• Priority: High
• Due Date: Dec 31, 2024
```

### Update Entry
```
Jane Smith updated this todo
• Title: "Complete project documentation" → "Complete API documentation"
• Priority: High → Medium
```

### Status Change Entry
```
John Doe changed status
• Status: Pending → In Progress
```

## Business Rules

1. **Immutable**: History entries cannot be modified or deleted
2. **Automatic**: All changes are tracked without user intervention
3. **Complete**: Every field change is recorded
4. **User Attribution**: Every change is linked to the user who made it

## Performance Considerations

1. **Lazy Loading**: History is only loaded when requested
2. **Pagination**: For todos with extensive history
3. **Indexing**: Proper indexes for efficient queries
4. **JSONB**: Efficient storage and querying of change data

## Security Considerations

1. **Read-Only**: History is read-only via API
2. **User Scoped**: Users can only see history for their own todos
3. **No Sensitive Data**: Passwords or tokens are never stored

## Future Enhancements

1. **Diff Visualization**: Better visualization of text changes
2. **Filtering**: Filter history by action type or date range
3. **Export**: Export history to CSV/PDF
4. **Notifications**: Subscribe to changes on specific todos
5. **Comparison**: Compare todo state at different points in time
6. **Restore**: Ability to restore previous versions
7. **Bulk History**: View history across multiple todos
8. **Analytics**: Insights from history data (e.g., average completion time)