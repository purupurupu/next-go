# Comments Feature

## Overview

The comments feature allows users to add discussions and notes to their todos. Comments support soft deletion, edit time limits, and are displayed in a threaded conversation format.

## Features

### 1. Comment Creation
- Add comments to any todo
- Rich text support (future enhancement)
- User attribution with timestamps
- Real-time validation

### 2. Comment Management
- **Edit Comments**: Edit your own comments within 15 minutes of creation
- **Delete Comments**: Soft delete preserves comment history
- **View History**: See deleted comments marked as "[Deleted]"

### 3. User Experience
- Comments displayed in chronological order
- User avatars and names
- Relative timestamps (e.g., "2 hours ago")
- Loading states and error handling

## Technical Implementation

### Backend

#### Model Structure
```ruby
class Comment < ApplicationRecord
  belongs_to :user
  belongs_to :commentable, polymorphic: true
  
  validates :content, presence: true
  
  scope :active, -> { where(deleted_at: nil) }
  scope :deleted, -> { where.not(deleted_at: nil) }
  
  def soft_delete!
    update!(deleted_at: Time.current)
  end
  
  def editable?
    created_at > 15.minutes.ago && deleted_at.nil?
  end
end
```

#### API Endpoints

**Base URL**: `/api/v1/todos/:todo_id/comments`

1. **List Comments**
   ```
   GET /api/v1/todos/:todo_id/comments
   ```
   Returns all comments (including soft-deleted) for a todo

2. **Create Comment**
   ```
   POST /api/v1/todos/:todo_id/comments
   {
     "comment": {
       "content": "This is a comment"
     }
   }
   ```

3. **Update Comment**
   ```
   PUT /api/v1/todos/:todo_id/comments/:id
   {
     "comment": {
       "content": "Updated comment"
     }
   }
   ```
   - Only allowed within 15 minutes of creation
   - Returns 403 if time limit exceeded

4. **Delete Comment**
   ```
   DELETE /api/v1/todos/:todo_id/comments/:id
   ```
   - Soft deletes the comment
   - Content remains visible as "[Deleted]"

### Frontend

#### Components

1. **CommentSection**
   - Container for all comment functionality
   - Manages comment state and API calls
   - Handles loading and error states

2. **CommentList**
   - Displays comments in chronological order
   - Shows user information and timestamps
   - Handles deleted comment display

3. **CommentItem**
   - Individual comment display
   - Edit/delete functionality
   - Time-based edit restrictions

4. **CommentForm**
   - Input for new comments
   - Validation and error display
   - Submit on Enter (Shift+Enter for new line)

#### Usage Example

```typescript
// In TodoDetail component
<CommentSection 
  todoId={todo.id} 
  initialComments={todo.comments}
/>

// API Integration
const commentApi = new CommentApiClient(httpClient);

// Create comment
await commentApi.create(todoId, { content: "Great task!" });

// Update comment
await commentApi.update(todoId, commentId, { content: "Updated!" });

// Delete comment
await commentApi.delete(todoId, commentId);
```

## Business Rules

1. **Edit Time Limit**: Comments can only be edited within 15 minutes of creation
2. **Soft Delete**: Comments are never hard deleted, preserving discussion history
3. **User Scope**: Users can only edit/delete their own comments
4. **Content Validation**: Comments cannot be blank

## Database Schema

```sql
CREATE TABLE comments (
  id bigserial PRIMARY KEY,
  content text NOT NULL,
  user_id bigint NOT NULL REFERENCES users(id),
  commentable_type varchar NOT NULL,
  commentable_id bigint NOT NULL,
  deleted_at timestamp,
  created_at timestamp(6) NOT NULL,
  updated_at timestamp(6) NOT NULL
);

CREATE INDEX index_comments_on_user_id ON comments(user_id);
CREATE INDEX index_comments_on_commentable ON comments(commentable_type, commentable_id);
CREATE INDEX index_comments_on_deleted_at ON comments(deleted_at);
```

## Security Considerations

1. **Authentication**: All comment operations require authentication
2. **Authorization**: Users can only modify their own comments
3. **Input Sanitization**: Content is sanitized to prevent XSS
4. **Rate Limiting**: Consider implementing rate limits for comment creation

## Performance Optimizations

1. **Eager Loading**: Comments are loaded with user information to prevent N+1 queries
2. **Pagination**: For todos with many comments, implement pagination
3. **Caching**: Consider caching comment counts

## Future Enhancements

1. **Rich Text Support**: Markdown or WYSIWYG editor
2. **Mentions**: Tag other users with @mentions
3. **Reactions**: Add emoji reactions to comments
4. **Threading**: Support nested comment threads
5. **Real-time Updates**: WebSocket support for live comments
6. **Attachments**: Allow file attachments in comments
7. **Search**: Search within comments
8. **Moderation**: Admin tools for comment moderation