# API Error Handling

This document describes the error handling system used by the Todo API.

## Error Response Format

All API errors follow a consistent JSON structure:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      // Additional context-specific information
    },
    "request_id": "unique-request-identifier",
    "timestamp": "2025-01-29T10:00:00Z"
  }
}
```

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `error.code` | string | Machine-readable error code for programmatic handling |
| `error.message` | string | Human-readable error message |
| `error.details` | object | Optional additional context about the error |
| `error.request_id` | string | Unique identifier for tracking the request |
| `error.timestamp` | string | ISO 8601 timestamp when the error occurred |

## Error Codes

### Authentication Errors (401)

| Code | Description | Example |
|------|-------------|---------|
| `AUTHENTICATION_REQUIRED` | No authentication token provided | Missing Authorization header |
| `INVALID_TOKEN` | Token is malformed or invalid | Corrupted JWT token |
| `TOKEN_EXPIRED` | Authentication token has expired | JWT token past expiration |
| `TOKEN_REVOKED` | Token has been revoked | User logged out |
| `INVALID_CREDENTIALS` | Email/password combination is incorrect | Login failure |

### Authorization Errors (403)

| Code | Description | Example |
|------|-------------|---------|
| `AUTHORIZATION_ERROR` | User lacks permission for this action | Accessing another user's resource |
| `FORBIDDEN` | Action is not allowed | Modifying system resources |

### Validation Errors (422)

| Code | Description | Example |
|------|-------------|---------|
| `VALIDATION_ERROR` | Request data failed validation | Missing required field |
| `INVALID_PARAMETER` | Parameter has invalid format or value | Invalid email format |
| `MISSING_PARAMETER` | Required parameter is missing | Missing title for todo |
| `DUPLICATE_RESOURCE` | Resource already exists | Duplicate category name |

### Resource Errors (404)

| Code | Description | Example |
|------|-------------|---------|
| `RESOURCE_NOT_FOUND` | Requested resource doesn't exist | Todo with ID not found |
| `ENDPOINT_NOT_FOUND` | API endpoint doesn't exist | Invalid URL path |

### Business Logic Errors (400)

| Code | Description | Example |
|------|-------------|---------|
| `INVALID_STATUS_TRANSITION` | Status change is not allowed | Completing a deleted todo |
| `LIMIT_EXCEEDED` | Resource limit has been exceeded | Too many todos created |
| `INVALID_OPERATION` | Operation cannot be performed | Invalid bulk operation |

### Server Errors (500)

| Code | Description | Example |
|------|-------------|---------|
| `INTERNAL_SERVER_ERROR` | Unexpected server error | Database connection failure |
| `SERVICE_UNAVAILABLE` | Service temporarily unavailable | Maintenance mode |

### Rate Limiting (429)

| Code | Description | Example |
|------|-------------|---------|
| `RATE_LIMIT_EXCEEDED` | Too many requests | API rate limit hit |

## Error Examples

### Authentication Error

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json
X-Request-Id: 550e8400-e29b-41d4-a716-446655440000

{
  "error": {
    "code": "AUTHENTICATION_REQUIRED",
    "message": "Authentication required for this endpoint",
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2025-01-29T10:15:30Z"
  }
}
```

### Validation Error

```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json
X-Request-Id: 660e8400-e29b-41d4-a716-446655440001

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "fields": {
        "title": ["can't be blank", "is too short (minimum is 1 character)"],
        "priority": ["is not included in the list"]
      }
    },
    "request_id": "660e8400-e29b-41d4-a716-446655440001",
    "timestamp": "2025-01-29T10:16:45Z"
  }
}
```

### Resource Not Found

```http
HTTP/1.1 404 Not Found
Content-Type: application/json
X-Request-Id: 770e8400-e29b-41d4-a716-446655440002

{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Todo with ID '123' not found",
    "details": {
      "resource": "Todo",
      "id": "123"
    },
    "request_id": "770e8400-e29b-41d4-a716-446655440002",
    "timestamp": "2025-01-29T10:18:00Z"
  }
}
```

### Business Logic Error

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json
X-Request-Id: 880e8400-e29b-41d4-a716-446655440003

{
  "error": {
    "code": "INVALID_STATUS_TRANSITION",
    "message": "Cannot transition from 'completed' to 'pending'",
    "details": {
      "current_status": "completed",
      "requested_status": "pending",
      "allowed_transitions": ["in_progress"]
    },
    "request_id": "880e8400-e29b-41d4-a716-446655440003",
    "timestamp": "2025-01-29T10:20:15Z"
  }
}
```

## Request ID Tracking

Every API response includes a `X-Request-Id` header and the same ID in error responses. This ID can be used to:

- Track requests through logs
- Debug issues with support
- Correlate frontend and backend events

Example:
```javascript
// Frontend error handling
fetch('/api/v1/todos')
  .then(response => {
    const requestId = response.headers.get('X-Request-Id');
    if (!response.ok) {
      return response.json().then(error => {
        console.error(`Request ${requestId} failed:`, error);
        throw error;
      });
    }
    return response.json();
  });
```

## Error Handling Best Practices

### Frontend Integration

1. **Always check response status**
   ```javascript
   if (!response.ok) {
     const error = await response.json();
     // Handle error based on error.code
   }
   ```

2. **Use error codes for logic**
   ```javascript
   switch (error.error.code) {
     case 'AUTHENTICATION_REQUIRED':
     case 'TOKEN_EXPIRED':
       // Redirect to login
       break;
     case 'VALIDATION_ERROR':
       // Show field-specific errors
       break;
     default:
       // Show generic error message
   }
   ```

3. **Log request IDs for debugging**
   ```javascript
   console.error(`API Error [${error.error.request_id}]:`, error.error.message);
   ```

### Retry Logic

Some errors are retryable:
- `500` Internal Server Error (with exponential backoff)
- `503` Service Unavailable
- Network timeouts

Others should not be retried:
- `401` Authentication errors
- `403` Authorization errors
- `422` Validation errors
- `404` Not found errors

### Security Considerations

- Error messages in production are sanitized to avoid leaking sensitive information
- Detailed error information is logged server-side with the request ID
- Stack traces are never exposed in production API responses

## Common Error Scenarios

### Expired Authentication Token

**Scenario**: User's session has expired

**Response**:
```json
{
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "Your session has expired. Please log in again.",
    "request_id": "...",
    "timestamp": "..."
  }
}
```

**Frontend handling**: Redirect to login page

### Concurrent Update Conflict

**Scenario**: Two users trying to update the same resource

**Response**:
```json
{
  "error": {
    "code": "CONFLICT",
    "message": "The resource has been modified by another user",
    "details": {
      "current_version": 5,
      "your_version": 3
    },
    "request_id": "...",
    "timestamp": "..."
  }
}
```

**Frontend handling**: Reload resource and retry or show merge conflict UI

### Rate Limiting

**Scenario**: Too many API requests

**Response**:
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "details": {
      "retry_after": 60,
      "limit": 100,
      "window": "1 minute"
    },
    "request_id": "...",
    "timestamp": "..."
  }
}
```

**Frontend handling**: Wait for `retry_after` seconds before retrying