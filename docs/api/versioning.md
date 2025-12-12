# API Versioning

## Overview

The Todo API uses URL-based versioning to ensure backward compatibility and smooth transitions between API versions.

## Current Version

- **Version**: `v1`
- **Base URL**: `https://api.example.com/api/v1`
- **Status**: Stable

## URL-Based Versioning

All API endpoints must include the version number in the URL path:

```
https://api.example.com/api/v1/todos
https://api.example.com/api/v1/categories
https://api.example.com/api/v1/tags
```

### Examples

```bash
# Get all todos
curl -X GET https://api.example.com/api/v1/todos \
  -H "Authorization: Bearer <token>"

# Create a new todo
curl -X POST https://api.example.com/api/v1/todos \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"todo": {"title": "New task"}}'
```

## Version Response Header

All API responses include the version in the response headers:

```
HTTP/1.1 200 OK
X-API-Version: v1
X-Request-Id: 550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json
```

## Client Implementation

### JavaScript/TypeScript

```typescript
const API_BASE_URL = 'https://api.example.com';
const API_VERSION = 'v1';

class TodoApiClient {
  private baseUrl = `${API_BASE_URL}/api/${API_VERSION}`;
  
  async getTodos() {
    const response = await fetch(`${this.baseUrl}/todos`, {
      headers: {
        'Authorization': `Bearer ${this.token}`
      }
    });
    return response.json();
  }
}
```

### Configuration

It's recommended to store the API version in your application's configuration:

```javascript
// config.js
export const API_CONFIG = {
  baseUrl: process.env.API_BASE_URL || 'http://localhost:3001',
  version: 'v1',
  
  getEndpoint(path) {
    return `${this.baseUrl}/api/${this.version}${path}`;
  }
};

// Usage
fetch(API_CONFIG.getEndpoint('/todos'), {
  headers: { 'Authorization': `Bearer ${token}` }
});
```

## Future Versions

When new API versions are released:

1. The existing `v1` endpoints will continue to work
2. New features will be available in the new version (e.g., `v2`)
3. Migration guides will be provided
4. Deprecation notices will be given well in advance

## Best Practices

1. **Always include the version** in your API calls
2. **Store the version in configuration** for easy updates
3. **Monitor release notes** for new versions
4. **Test against new versions** before migrating production

## Support

For questions about API versioning:

- Documentation: This guide
- Support: api-support@example.com