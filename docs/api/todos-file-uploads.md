# Todo File Uploads API

Todo にファイルを添付するための API です。

## Overview

Todo には複数のファイルを添付できます。ファイルは RustFS (S3 互換ストレージ) に保存されます。

## 技術スタック

- **Storage**: RustFS (S3 互換)
- **Backend**: Go + Echo
- **Endpoint**: `http://localhost:9000` (RustFS)
- **Console**: `http://localhost:9001` (RustFS Console)

## Endpoints

### Upload Files

**POST** `/api/v1/todos/:todo_id/files`

Todo にファイルを添付します。

```bash
curl -X POST http://localhost:3001/api/v1/todos/1/files \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/document.pdf"
```

**Success Response (201 Created):**
```json
{
  "id": 1,
  "filename": "document.pdf",
  "content_type": "application/pdf",
  "byte_size": 102400,
  "url": "http://localhost:9000/todo-files/1/document.pdf"
}
```

### List Files

**GET** `/api/v1/todos/:todo_id/files`

Todo に添付されたファイル一覧を取得します。

**Success Response (200 OK):**
```json
[
  {
    "id": 1,
    "filename": "document.pdf",
    "content_type": "application/pdf",
    "byte_size": 102400,
    "url": "http://localhost:9000/todo-files/1/document.pdf"
  },
  {
    "id": 2,
    "filename": "photo.jpg",
    "content_type": "image/jpeg",
    "byte_size": 51200,
    "url": "http://localhost:9000/todo-files/1/photo.jpg"
  }
]
```

### Delete File

**DELETE** `/api/v1/todos/:todo_id/files/:file_id`

ファイルを削除します。

```bash
curl -X DELETE http://localhost:3001/api/v1/todos/1/files/123 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Success Response (204 No Content)**

## Response Format

Todo レスポンスに含まれる files 配列:

```json
{
  "id": 1,
  "title": "Todo with attachments",
  "completed": false,
  "files": [
    {
      "id": 123,
      "filename": "document.pdf",
      "content_type": "application/pdf",
      "byte_size": 102400,
      "url": "http://localhost:9000/todo-files/..."
    }
  ]
}
```

## File Object

| Property | Type | Description |
|----------|------|-------------|
| `id` | Integer | ファイル ID |
| `filename` | String | ファイル名 |
| `content_type` | String | MIME タイプ |
| `byte_size` | Integer | ファイルサイズ (bytes) |
| `url` | String | ダウンロード URL |

## File Validations

### File Size
- 最大: **10MB**
- エラー: "ファイルサイズは10MB以下にしてください"

### Allowed File Types

| Category | Types |
|----------|-------|
| Images | JPEG, PNG, GIF, WebP |
| Documents | PDF, DOC, DOCX, XLS, XLSX |
| Text | TXT, CSV |

許可されていないファイルタイプの場合:
- エラー: "許可されていないファイルタイプです"

## Frontend Implementation

### Upload File

```typescript
async function uploadFile(todoId: number, file: File) {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch(`/api/v1/todos/${todoId}/files`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
      // Content-Type は自動設定される
    },
    body: formData
  });

  return response.json();
}
```

### Display Files

```typescript
// ダウンロードリンク
<a href={file.url} download={file.filename}>
  {file.filename} をダウンロード
</a>

// 画像プレビュー
{file.content_type.startsWith('image/') && (
  <img src={file.url} alt={file.filename} />
)}
```

## Storage Configuration

### Development (Docker Compose)

```yaml
# compose.yml
rustfs:
  image: rustfs/rustfs:latest
  ports:
    - "9000:9000"  # S3 API
    - "9001:9001"  # Console
  environment:
    RUSTFS_ROOT_USER: admin
    RUSTFS_ROOT_PASSWORD: password123
  volumes:
    - rustfs_data:/data
```

### Environment Variables

```env
STORAGE_ENDPOINT=http://rustfs:9000
STORAGE_ACCESS_KEY=admin
STORAGE_SECRET_KEY=password123
STORAGE_BUCKET=todo-files
STORAGE_REGION=us-east-1
```

## Security Considerations

1. **File Type Validation**: サーバー側でファイルタイプを検証
2. **File Size Limits**: 10MB 制限
3. **Access Control**: ファイルは親 Todo と同じアクセス権限
4. **Signed URLs**: 本番環境では署名付き URL を使用

## Error Responses

### 400 Bad Request
```json
{
  "error": {
    "code": "INVALID_FILE",
    "message": "許可されていないファイルタイプです"
  }
}
```

### 413 Payload Too Large
```json
{
  "error": {
    "code": "FILE_TOO_LARGE",
    "message": "ファイルサイズは10MB以下にしてください"
  }
}
```

### 404 Not Found
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "File with ID '123' not found"
  }
}
```
