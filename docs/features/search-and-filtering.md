# Search and Filtering

## Overview

The todo application provides comprehensive search and filtering capabilities to help users find and organize their tasks efficiently. The search functionality is implemented through a dedicated API endpoint that supports full-text search, multiple filter criteria, sorting, and pagination.

## Features

### 1. Full-Text Search

Search across todo titles and descriptions:
- **Case-insensitive** partial matching
- **Highlight support** showing exact match positions
- **Real-time search** with debouncing for optimal performance

Example:
```
GET /api/todos/search?q=documentation
```

### 2. Category Filtering

Filter todos by category:
- Filter by specific category ID
- Use `-1` to find uncategorized todos
- Combine with other filters

Example:
```
GET /api/todos/search?category_id=1
GET /api/todos/search?category_id=-1  # Uncategorized todos
```

### 3. Status Filtering

Filter by one or multiple statuses:
- `pending` - Not started tasks
- `in_progress` - Tasks being worked on
- `completed` - Finished tasks

Example:
```
GET /api/todos/search?status[]=pending&status[]=in_progress
```

### 4. Priority Filtering

Filter by priority levels:
- `high` - High priority tasks
- `medium` - Medium priority tasks
- `low` - Low priority tasks

Example:
```
GET /api/todos/search?priority[]=high&priority[]=medium
```

### 5. Tag Filtering

Advanced tag filtering with two modes:
- **ANY mode** (default): Match todos with any of the specified tags
- **ALL mode**: Match todos with all specified tags

Example:
```
GET /api/todos/search?tag_ids[]=1&tag_ids[]=2&tag_mode=any   # Has tag 1 OR tag 2
GET /api/todos/search?tag_ids[]=1&tag_ids[]=2&tag_mode=all   # Has tag 1 AND tag 2
```

### 6. Date Range Filtering

Filter todos by due date:
- `due_date_from` - Start of date range
- `due_date_to` - End of date range

Example:
```
GET /api/todos/search?due_date_from=2024-01-01&due_date_to=2024-12-31
```

### 7. Sorting Options

Sort results by various fields:
- `position` - Default drag-and-drop order
- `created_at` - Creation timestamp
- `updated_at` - Last modification time
- `due_date` - Task deadline
- `title` - Alphabetical by title
- `priority` - Priority level (high → low)
- `status` - Status order

Each field supports both ascending (`asc`) and descending (`desc`) order:
```
GET /api/todos/search?sort_by=due_date&sort_order=asc
GET /api/todos/search?sort_by=priority&sort_order=desc
```

### 8. Pagination

Handle large result sets efficiently:
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 20, max: 100)

Example:
```
GET /api/todos/search?page=2&per_page=50
```

## Response Format

### Successful Search

```json
{
  "todos": [
    {
      "id": 1,
      "title": "Complete API documentation",
      "description": "Write comprehensive documentation with examples",
      // ... other todo fields ...
      "highlights": {
        "title": [
          {
            "start": 12,
            "end": 25,
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
      "category_id": 1,
      "status": ["pending", "in_progress"]
    }
  }
}
```

### No Results with Suggestions

When no results are found, helpful suggestions are provided:

```json
{
  "todos": [],
  "meta": {
    "total": 0,
    "current_page": 1,
    "total_pages": 0,
    "per_page": 20
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

## Frontend Implementation

### Search Components

1. **SearchBar**
   - Real-time search input with debouncing
   - Clear button for quick reset
   - Keyboard-friendly (Escape to clear)

2. **AdvancedFilters**
   - Collapsible panel for detailed filters
   - Category selector with "uncategorized" option
   - Multi-select for status and priority
   - Tag selector with AND/OR toggle
   - Date range pickers
   - Sort options dropdown

3. **FilterBadges**
   - Visual display of active filters
   - Individual filter removal
   - Clear all filters option

4. **HighlightedText**
   - Renders search matches with highlighting
   - Used in todo titles and descriptions

### Search Hooks

1. **useSearchParams**
   - Centralized search parameter management
   - URL synchronization for shareable searches
   - Filter state management

2. **useTodoSearch**
   - API integration with automatic debouncing
   - Loading and error states
   - Search result caching

3. **useDebounce**
   - Generic debouncing utility
   - Prevents excessive API calls

### Implementation Example

```typescript
// Using the search functionality
const { 
  searchParams, 
  updateSearchQuery, 
  updateCategory,
  updateStatus,
  clearFilters 
} = useSearchParams();

const { 
  todos, 
  loading, 
  searchResponse 
} = useTodoSearch(searchParams);

// Render search UI
<SearchBar 
  value={searchParams.q || ''} 
  onChange={updateSearchQuery} 
/>
<AdvancedFilters 
  searchParams={searchParams}
  onUpdateCategory={updateCategory}
  onUpdateStatus={updateStatus}
  // ... other handlers
/>
```

## Performance Optimization

### Backend Optimizations

1. **Database Indexes**
   - Index on `todos.title` for text search
   - Index on `todos.description` for text search
   - Composite indexes for common filter combinations
   - Indexes on foreign keys (user_id, category_id)

2. **Query Optimization**
   - Efficient use of ActiveRecord scopes
   - N+1 query prevention with `includes`
   - Selective field loading where appropriate

3. **Caching Strategy**
   - Consider caching frequent searches
   - Cache category and tag lists
   - Implement ETags for conditional requests

### Frontend Optimizations

1. **Debouncing**
   - 300ms delay for search input
   - Prevents rapid API calls while typing

2. **Memoization**
   - Cache search parameters with `useMemo`
   - Prevent unnecessary re-renders

3. **Optimistic Updates**
   - Update UI immediately on user actions
   - Sync with server in background

4. **Progressive Enhancement**
   - Show skeleton loaders during search
   - Maintain previous results while loading

## Best Practices

### For Users

1. **Start with broad searches** and narrow down with filters
2. **Use the clear button** to reset complex filter combinations
3. **Combine filters** for precise results
4. **Save common searches** by bookmarking the URL

### For Developers

1. **Always debounce** search inputs
2. **Show loading states** for better UX
3. **Handle errors gracefully** with fallback UI
4. **Test edge cases** like empty results, single results, pagination boundaries
5. **Monitor performance** with search query analytics

## Technical Details

### Search Implementation

The search functionality is powered by:
- **Backend**: `TodoSearchService` class using ActiveRecord scopes
- **Database**: PostgreSQL with ILIKE for case-insensitive search
- **Frontend**: React hooks with TypeScript for type safety

### Security Considerations

- All searches are scoped to the authenticated user
- Input sanitization prevents SQL injection
- Rate limiting should be considered for production

### Limitations

- Search is limited to title and description fields
- No fuzzy matching or typo correction
- No search history or saved searches (yet)
- Maximum 100 items per page

## Future Enhancements

Potential improvements for the search functionality:

1. **Full-text search** with PostgreSQL's built-in FTS
2. **Elasticsearch integration** for advanced search features
3. **Search suggestions** and autocomplete
4. **Saved searches** and search history
5. **Export search results** to CSV/PDF
6. **Smart filters** (e.g., "Due this week", "Overdue")
7. **Natural language search** (e.g., "high priority tasks due tomorrow")