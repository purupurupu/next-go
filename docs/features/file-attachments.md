# File Attachments Feature

## Overview

ファイル添付機能により、ユーザーは Todo にファイルを添付できます。Go バックエンドと RustFS (S3互換ストレージ) を使用して実装されています。

## Features

### 1. File Upload
- Todo ごとに複数ファイル添付可能
- ファイルタイプ検証
- サイズ制限（10MB）
- 画像のサムネイル自動生成

### 2. File Management
- ファイル一覧表示
- ダウンロード
- 個別削除
- サムネイル・中間サイズ画像

### 3. Supported File Types
- **Documents**: PDF, DOC, DOCX, TXT, MD
- **Images**: JPG, JPEG, PNG, GIF, WebP
- **Spreadsheets**: XLS, XLSX, CSV
- **Archives**: ZIP
- **Code**: JSON, XML

## Technical Implementation

### Backend (Go)

#### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/todos/:todo_id/files` | ファイル一覧 |
| POST | `/api/v1/todos/:todo_id/files` | アップロード |
| GET | `/api/v1/todos/:todo_id/files/:file_id` | ダウンロード |
| GET | `/api/v1/todos/:todo_id/files/:file_id/thumb` | サムネイル |
| GET | `/api/v1/todos/:todo_id/files/:file_id/medium` | 中間サイズ |
| DELETE | `/api/v1/todos/:todo_id/files/:file_id` | 削除 |

#### Response Format

```json
{
  "id": 1,
  "original_name": "document.pdf",
  "content_type": "application/pdf",
  "file_size": 245000,
  "file_type": "document",
  "download_url": "/api/v1/todos/1/files/1",
  "thumb_url": null,
  "medium_url": null,
  "created_at": "2024-01-01T00:00:00Z"
}
```

画像の場合:
```json
{
  "id": 2,
  "original_name": "photo.jpg",
  "content_type": "image/jpeg",
  "file_size": 1200000,
  "file_type": "image",
  "download_url": "/api/v1/todos/1/files/2",
  "thumb_url": "/api/v1/todos/1/files/2/thumb",
  "medium_url": "/api/v1/todos/1/files/2/medium",
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### Architecture

```
backend/internal/
├── handler/file.go      # HTTP handlers
├── service/
│   ├── file.go          # Business logic
│   └── thumbnail.go     # Image processing
├── storage/s3.go        # S3 client abstraction
├── model/file.go        # File model
└── repository/file.go   # Database access
```

#### Model

```go
type File struct {
    ID           int64     `gorm:"primaryKey"`
    TodoID       int64     `gorm:"not null;index"`
    OriginalName string    `gorm:"not null;size:255"`
    StoragePath  string    `gorm:"not null;size:500"`
    ContentType  string    `gorm:"not null;size:100"`
    FileSize     int64     `gorm:"not null"`
    FileType     FileType  `gorm:"not null;size:20"`
    ThumbPath    *string   `gorm:"size:500"`
    MediumPath   *string   `gorm:"size:500"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Storage Configuration

#### Docker Compose (RustFS)

```yaml
rustfs:
  image: rustfs/rustfs:latest
  ports:
    - "9000:9000"
    - "9001:9001"
  environment:
    - RUSTFS_ACCESS_KEY=${S3_ACCESS_KEY:-rustfs-dev-access}
    - RUSTFS_SECRET_KEY=${S3_SECRET_KEY:-rustfs-dev-secret-key}
  volumes:
    - ./data/rustfs:/data
```

#### Backend Environment

```yaml
environment:
  - S3_ENDPOINT=http://rustfs:9000
  - S3_REGION=us-east-1
  - S3_BUCKET=todo-files
  - S3_ACCESS_KEY=${RUSTFS_ACCESS_KEY:-rustfs-dev-access}
  - S3_SECRET_KEY=${RUSTFS_SECRET_KEY:-rustfs-dev-secret-key}
  - S3_USE_PATH_STYLE=true
```

### Frontend

#### Upload Example

```typescript
// features/file/lib/api.ts
export async function uploadFile(todoId: number, file: File): Promise<FileAttachment> {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch(`${API_BASE}/todos/${todoId}/files`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: formData
  });

  return response.json();
}
```

#### Display Attachments

```typescript
// Show thumbnail for images
{file.thumb_url && (
  <img src={file.thumb_url} alt={file.original_name} />
)}

// Download link
<a href={file.download_url}>Download</a>
```

## Image Processing

### Thumbnail Generation

画像アップロード時に自動生成:
- **Thumbnail**: 150x150px (アスペクト比維持)
- **Medium**: 800x800px (アスペクト比維持)

処理フロー:
1. ファイルアップロード
2. Content-Type が画像の場合、サムネイルサービスを呼び出し
3. オリジナル・サムネイル・中間サイズを S3 に保存
4. パスをデータベースに記録

## Security Considerations

1. **File Type Validation**: Content-Type のホワイトリスト検証
2. **Size Limits**: 最大 10MB/ファイル
3. **Access Control**: Todo の所有者のみアクセス可能
4. **Signed URLs**: プロダクション環境では署名付き URL を使用

## Limitations

1. **File Size**: 最大 10MB/ファイル
2. **File Types**: ホワイトリストのみ許可
3. **Storage**: 開発環境は RustFS、本番は S3 推奨
