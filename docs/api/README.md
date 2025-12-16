# API Documentation

## Overview

Todo アプリケーションは Go Echo フレームワークで構築された RESTful JSON API を提供します。すべてのエンドポイントは JSON レスポンスを返し、必要に応じて JSON リクエストボディを受け取ります。

## 技術スタック

- **言語**: Go 1.25
- **フレームワーク**: Echo v4
- **ORM**: GORM
- **認証**: JWT (golang-jwt/jwt v5)
- **バリデーション**: go-playground/validator v10

## Base URL

- **Development**: `http://localhost:3001/api/v1`
- **Production**: デプロイ時に設定

**Note**: API は URL ベースのバージョニングを使用します (`/api/v1`)。

## Authentication

ほとんどの API エンドポイントは JWT 認証が必要です。Authorization ヘッダーにトークンを含めてください:

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
```

## Response Format

### 単一リソース (Create, Show, Update)

オブジェクトを直接返却:

```json
{
  "id": 1,
  "title": "Complete project",
  "completed": false,
  "position": 0,
  "due_date": "2024-12-31",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### リスト (List, Search)

data と meta でラップ:

```json
{
  "data": [
    {
      "id": 1,
      "title": "Complete project",
      "completed": false
    }
  ],
  "meta": {
    "total": 100,
    "current_page": 1,
    "total_pages": 5,
    "per_page": 20
  }
}
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
    }
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
- [Categories API](./categories.md) - Category CRUD operations
- [Tags API](./tags.md) - Tag CRUD operations
- [Comments API](./comments.md) - Comment functionality for todos (15分編集制限)
- [Todo History API](./todo-histories.md) - Change tracking and audit history
- [File Uploads API](./todos-file-uploads.md) - File attachments (RustFS/S3)
- [Notes API](./notes.md) - Markdown notes with revision history

## Pagination

ページネーションは一覧・検索エンドポイントで使用:

```
GET /api/v1/todos/search?page=1&per_page=20
GET /api/v1/notes?page=1&per_page=20
```

Parameters:
- `page` - ページ番号 (default: 1)
- `per_page` - 1ページあたりの件数 (default: 20, max: 100)

## Versioning

API は URL ベースのバージョニングを使用:

- **Current Version**: `v1`
- **URL Format**: `/api/v1/{resource}`
- **Example**: `/api/v1/todos`

## CORS

CORS は以下のオリジンを許可:
- Development: `http://localhost:3000`
- Production: 環境変数で設定

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

# Create note
curl -X POST http://localhost:3001/api/v1/notes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt_token>" \
  -d '{"title":"My Note","body_md":"# Hello"}'
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
