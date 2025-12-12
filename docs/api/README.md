# API Documentation

## Overview

The Todo application provides a RESTful JSON API built with Rails API-only mode. All endpoints return JSON responses and require JSON request bodies where applicable.

## Base URL

- **Development**: `http://localhost:3001/api/v1`
- **Production**: Configure in deployment

**Note**: The API uses URL-based versioning (`/api/v1`). All endpoints must include the version in the URL path.

## Authentication

Most API endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

## Common Headers

### Request Headers
```
Content-Type: application/json
Accept: application/json
Authorization: Bearer <jwt_token>
```

### Response Headers
```
Content-Type: application/json
X-Request-Id: <unique_request_id>
X-API-Version: v1
```

## Response Format (UNIFIED)

### Success Response with Data
```json
{
  "message": "Todo created successfully",
  "data": {
    "id": 1,
    "title": "Complete project",
    "completed": false,
    "position": 0,
    "due_date": "2024-12-31",
    "created_at": "2024-01-01T00:00:00.000Z",
    "updated_at": "2024-01-01T00:00:00.000Z"
  }
}
```

### Success Response (Message Only)
```json
{
  "message": "Logged out successfully"
}
```

### Resource List Response
```json
[
  {
    "id": 1,
    "title": "Complete project",
    "completed": false,
    "position": 0,
    "due_date": "2024-12-31"
  }
]
```

### Error Response
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Todo with ID '123' not found",
    "details": {
      "resource": "Todo",
      "id": "123"
    },
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2025-01-29T10:00:00Z"
  }
}
```

See [Error Handling](./errors.md) for complete error documentation.

## HTTP Status Codes

| Status Code | Description |
|------------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 204 | No Content - Request successful, no content to return |
| 400 | Bad Request - Invalid request parameters |
| 401 | Unauthorized - Missing or invalid authentication |
| 404 | Not Found - Resource not found |
| 422 | Unprocessable Entity - Validation errors |
| 500 | Internal Server Error - Server error |

## API Endpoints

### Core Documentation
- [Error Handling](./errors.md) - Error codes, formats, and troubleshooting
- [API Versioning](./versioning.md) - Version support and migration guides

### Authentication
- [Authentication API](./authentication.md) - User registration, login, and logout

### Resources
- [Todos API](./todos.md) - Todo CRUD operations, search, and batch updates
  - Basic CRUD operations (GET, POST, PUT, DELETE)
  - Advanced search and filtering (GET /api/todos/search)
  - Bulk position updates for drag-and-drop
  - File attachments support
- [Categories API](./categories.md) - Category CRUD operations
- [Tags API](./tags.md) - Tag CRUD operations for flexible todo organization
- [Comments API](./comments.md) - Comment functionality for todos
- [Todo History API](./todo-histories.md) - Change tracking and audit history for todos

## Pagination

Pagination is implemented for the search endpoint:
```
GET /api/todos/search?page=1&per_page=20
```

Pagination parameters:
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 20, max: 100)

Pagination metadata is returned in the response:
```json
{
  "todos": [...],
  "meta": {
    "total": 100,
    "current_page": 1,
    "total_pages": 5,
    "per_page": 20
  }
}
```

## Rate Limiting

Currently not implemented. Future versions may include rate limiting headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1640995200
```

## Versioning

The API uses URL-based versioning. All endpoints must include the version number in the path:

- **Current Version**: `v1`
- **URL Format**: `/api/v1/{resource}`
- **Example**: `/api/v1/todos`

## CORS

CORS is configured to allow requests from:
- Development: `http://localhost:3000`
- Production: Configure allowed origins

## Request Examples

### Using cURL
```bash
# Login
curl -X POST http://localhost:3001/auth/sign_in \
  -H "Content-Type: application/json" \
  -d '{"user":{"email":"user@example.com","password":"password"}}'

# Get todos
curl -X GET http://localhost:3001/api/v1/todos \
  -H "Authorization: Bearer <jwt_token>"
```

### Using JavaScript (Fetch)
```javascript
// Login
const response = await fetch('http://localhost:3001/auth/sign_in', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    user: { email: 'user@example.com', password: 'password' }
  })
});

// Get todos
const todos = await fetch('http://localhost:3001/api/v1/todos', {
  headers: { 'Authorization': `Bearer ${token}` }
}).then(res => res.json());
```

## Error Handling Best Practices

1. Always check response status before parsing
2. Use error codes for programmatic handling
3. Log request IDs for debugging
4. Handle authentication errors by redirecting to login
5. Implement retry logic for 5xx errors only
6. Display user-friendly messages based on error codes

See [Error Handling](./errors.md) for detailed error handling guidelines.

## Security Considerations

1. Always use HTTPS in production
2. Store JWT tokens securely (HttpOnly cookies recommended)
3. Implement token refresh mechanism
4. Validate all input on both client and server
5. Use CORS to restrict origins
6. Implement rate limiting in production