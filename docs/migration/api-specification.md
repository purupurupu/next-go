# API仕様書

## 概要

- **ベースURL**: `http://localhost:3001`
- **認証**: Bearer Token (JWT)
- **コンテンツタイプ**: `application/json`
- **APIバージョン**: v1（レスポンスヘッダー: `X-API-Version: v1`）

## 共通仕様

### 認証ヘッダー
```
Authorization: Bearer <jwt_token>
```

### レスポンスヘッダー
```
X-API-Version: v1
X-Request-Id: <uuid>
```

### 成功レスポンス形式
```json
{
  "status": {
    "code": 200,
    "message": "Success"
  },
  "data": { ... }
}
```

### ページネーション付きレスポンス
```json
{
  "status": {
    "code": 200,
    "message": "Success"
  },
  "data": [ ... ],
  "meta": {
    "total": 100,
    "current_page": 1,
    "total_pages": 5,
    "per_page": 20
  }
}
```

### エラーレスポンス形式
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": { ... },
    "request_id": "uuid",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

---

## 1. 認証API

### POST /auth/sign_up
ユーザー登録

**リクエスト**
```json
{
  "user": {
    "email": "user@example.com",
    "password": "password123",
    "password_confirmation": "password123",
    "name": "John Doe"
  }
}
```

**成功レスポンス** (201 Created)
```json
{
  "status": {
    "code": 201,
    "message": "Signed up successfully."
  },
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```
**レスポンスヘッダー**: `Authorization: Bearer <jwt_token>`

---

### POST /auth/sign_in
ユーザーログイン

**リクエスト**
```json
{
  "user": {
    "email": "user@example.com",
    "password": "password123"
  }
}
```

**成功レスポンス** (200 OK)
```json
{
  "status": {
    "code": 200,
    "message": "Logged in successfully."
  },
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```
**レスポンスヘッダー**: `Authorization: Bearer <jwt_token>`

---

### DELETE /auth/sign_out
ログアウト

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**成功レスポンス** (200 OK)
```json
{
  "status": {
    "code": 200,
    "message": "Logged out successfully."
  }
}
```

---

## 2. Todo API

### GET /api/v1/todos
Todo一覧取得（ユーザースコープ）

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**成功レスポンス** (200 OK)
```json
{
  "status": {
    "code": 200,
    "message": "Success"
  },
  "data": [
    {
      "id": 1,
      "title": "Buy groceries",
      "description": "Milk, eggs, bread",
      "completed": false,
      "position": 1,
      "priority": "medium",
      "status": "pending",
      "due_date": "2025-01-20",
      "user_id": 1,
      "category": {
        "id": 1,
        "name": "Shopping",
        "color": "#FF5733"
      },
      "tags": [
        { "id": 1, "name": "urgent", "color": "#FF0000" }
      ],
      "files": [
        {
          "id": "abc123",
          "filename": "list.pdf",
          "content_type": "application/pdf",
          "byte_size": 102400,
          "url": "http://localhost:3001/rails/active_storage/blobs/..."
        }
      ],
      "comments_count": 3,
      "history_count": 5,
      "latest_comments": [
        {
          "id": 1,
          "content": "Don't forget the milk!",
          "user": { "id": 1, "email": "user@example.com", "name": "John" },
          "created_at": "2025-01-15T10:30:00Z"
        }
      ],
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

### GET /api/v1/todos/search
Todo検索・フィルタリング

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**クエリパラメータ**
| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `q` | string | 検索クエリ（title, descriptionで検索） |
| `status` | string | ステータス（pending/in_progress/completed、カンマ区切りで複数可） |
| `priority` | string | 優先度（low/medium/high） |
| `category_id` | integer | カテゴリID（-1または'null'でカテゴリなし） |
| `tag_ids` | string | タグID（カンマ区切り） |
| `tag_mode` | string | タグ検索モード（any=OR / all=AND）デフォルト: any |
| `due_date_from` | date | 期限開始日（YYYY-MM-DD） |
| `due_date_to` | date | 期限終了日（YYYY-MM-DD） |
| `sort_by` | string | ソートフィールド（created_at/updated_at/due_date/title/priority/status） |
| `sort_order` | string | ソート順（asc/desc）デフォルト: desc |
| `page` | integer | ページ番号 デフォルト: 1 |
| `per_page` | integer | 1ページあたりの件数（1-100）デフォルト: 20 |

**例**
```
GET /api/v1/todos/search?q=meeting&status=pending,in_progress&priority=high&sort_by=due_date&page=1
```

**成功レスポンス** (200 OK)
```json
{
  "status": {
    "code": 200,
    "message": "Success"
  },
  "data": [ ... ],
  "meta": {
    "total": 45,
    "current_page": 1,
    "total_pages": 3,
    "per_page": 20,
    "active_filters": {
      "search": "meeting",
      "status": ["pending", "in_progress"],
      "priority": "high"
    }
  }
}
```

**結果が空の場合のサジェスション**
```json
{
  "data": [],
  "meta": {
    "total": 0,
    "suggestions": [
      { "type": "spelling", "message": "Check your spelling" },
      { "type": "reduce_filters", "message": "Try removing some filters" }
    ]
  }
}
```

---

### GET /api/v1/todos/:id
Todo詳細取得

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**成功レスポンス** (200 OK)
```json
{
  "status": {
    "code": 200,
    "message": "Success"
  },
  "data": {
    "id": 1,
    "title": "Buy groceries",
    ...
  }
}
```

---

### POST /api/v1/todos
Todo作成

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**リクエスト** (application/json)
```json
{
  "todo": {
    "title": "New task",
    "description": "Task description",
    "priority": "high",
    "status": "pending",
    "due_date": "2025-01-20",
    "category_id": 1,
    "tag_ids": [1, 2, 3]
  }
}
```

**ファイル添付時** (multipart/form-data)
```
todo[title]: New task
todo[description]: Task description
todo[files][]: <file1>
todo[files][]: <file2>
```

**成功レスポンス** (201 Created)
```json
{
  "status": {
    "code": 201,
    "message": "Todo created successfully"
  },
  "data": {
    "id": 2,
    "title": "New task",
    "position": 2,
    ...
  }
}
```

---

### PATCH /api/v1/todos/:id
Todo更新

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**リクエスト**
```json
{
  "todo": {
    "title": "Updated title",
    "completed": true,
    "status": "completed"
  }
}
```

**成功レスポンス** (200 OK)

---

### DELETE /api/v1/todos/:id
Todo削除

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**成功レスポンス** (204 No Content)

---

### PATCH /api/v1/todos/update_order
Todo順序一括更新（ドラッグ&ドロップ用）

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**リクエスト**
```json
{
  "todos": [
    { "id": 3, "position": 1 },
    { "id": 1, "position": 2 },
    { "id": 2, "position": 3 }
  ]
}
```

**成功レスポンス** (200 OK)
```json
{
  "status": {
    "code": 200,
    "message": "Order updated successfully"
  },
  "data": [ ... ]
}
```

---

### PATCH /api/v1/todos/:id/tags
Todoのタグ更新

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**リクエスト**
```json
{
  "todo": {
    "tag_ids": [1, 2, 3]
  }
}
```

**成功レスポンス** (200 OK)

---

### DELETE /api/v1/todos/:id/files/:file_id
添付ファイル削除

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**成功レスポンス** (200 OK)
```json
{
  "status": {
    "code": 200,
    "message": "File removed successfully"
  }
}
```

---

## 3. Category API

### GET /api/v1/categories
カテゴリ一覧取得

**ヘッダー**: `Authorization: Bearer <jwt_token>`

**成功レスポンス** (200 OK)
```json
{
  "status": {
    "code": 200,
    "message": "Success"
  },
  "data": [
    {
      "id": 1,
      "name": "Work",
      "color": "#FF5733",
      "todo_count": 5,
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

### GET /api/v1/categories/:id
カテゴリ詳細取得

---

### POST /api/v1/categories
カテゴリ作成

**リクエスト**
```json
{
  "category": {
    "name": "Personal",
    "color": "#3498DB"
  }
}
```

**成功レスポンス** (201 Created)

---

### PATCH /api/v1/categories/:id
カテゴリ更新

---

### DELETE /api/v1/categories/:id
カテゴリ削除

**注意**: カテゴリ削除時、関連するTodoの`category_id`は`null`に設定される（nullify）

---

## 4. Tag API

### GET /api/v1/tags
タグ一覧取得（名前順でソート）

**成功レスポンス** (200 OK)
```json
{
  "data": [
    {
      "id": 1,
      "name": "bug",
      "color": "#E74C3C",
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

### GET /api/v1/tags/:id
タグ詳細取得

---

### POST /api/v1/tags
タグ作成

**リクエスト**
```json
{
  "tag": {
    "name": "feature",
    "color": "#2ECC71"
  }
}
```

---

### PATCH /api/v1/tags/:id
タグ更新

---

### DELETE /api/v1/tags/:id
タグ削除

---

## 5. Comment API

### GET /api/v1/todos/:todo_id/comments
コメント一覧取得（時系列順）

**成功レスポンス** (200 OK)
```json
{
  "data": [
    {
      "id": 1,
      "content": "This is a comment",
      "user": {
        "id": 1,
        "email": "user@example.com",
        "name": "John"
      },
      "editable": true,
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

### POST /api/v1/todos/:todo_id/comments
コメント作成

**リクエスト**
```json
{
  "comment": {
    "content": "This is a new comment"
  }
}
```

---

### PATCH /api/v1/todos/:todo_id/comments/:id
コメント更新

**制限**:
- 作成者のみ編集可能（403 Forbidden）
- 作成から15分以内のみ編集可能（422 Unprocessable Entity）

**リクエスト**
```json
{
  "comment": {
    "content": "Updated comment content"
  }
}
```

---

### DELETE /api/v1/todos/:todo_id/comments/:id
コメント削除（ソフトデリート）

**注意**: 物理削除ではなく、`deleted_at`に現在時刻を設定

---

## 6. History API

### GET /api/v1/todos/:todo_id/histories
Todo変更履歴取得（最新50件）

**成功レスポンス** (200 OK)
```json
{
  "data": [
    {
      "id": 1,
      "field_name": "status",
      "old_value": "pending",
      "new_value": "completed",
      "action": "status_changed",
      "user": {
        "id": 1,
        "email": "user@example.com",
        "name": "John"
      },
      "human_readable_change": "ステータスを「未着手」から「完了」に変更",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

## 7. Note API

### GET /api/v1/notes
Note一覧取得

**クエリパラメータ**
| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `q` | string | 検索クエリ（title, body_plainで検索） |
| `archived` | boolean | アーカイブ済みを表示 |
| `trashed` | boolean | ゴミ箱を表示 |
| `pinned` | boolean | ピン留めのみ表示 |
| `page` | integer | ページ番号 |
| `per_page` | integer | 1ページあたりの件数 |

**成功レスポンス** (200 OK)
```json
{
  "data": [
    {
      "id": 1,
      "title": "Meeting Notes",
      "body_md": "# Meeting Notes\n\n- Point 1\n- Point 2",
      "pinned": true,
      "archived": false,
      "trashed": false,
      "last_edited_at": "2025-01-15T10:30:00Z",
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

---

### GET /api/v1/notes/:id
Note詳細取得

---

### POST /api/v1/notes
Note作成

**リクエスト**
```json
{
  "note": {
    "title": "New Note",
    "body_md": "# Content\n\nMarkdown content here",
    "pinned": false
  }
}
```

---

### PATCH /api/v1/notes/:id
Note更新

---

### DELETE /api/v1/notes/:id
Note削除

**クエリパラメータ**
| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `force` | boolean | true: 完全削除 / false: ゴミ箱へ移動 |

---

### GET /api/v1/notes/:id/revisions
Noteリビジョン一覧取得

**成功レスポンス** (200 OK)
```json
{
  "data": [
    {
      "id": 1,
      "note_id": 1,
      "title": "Previous Title",
      "body_md": "Previous content...",
      "created_at": "2025-01-14T10:30:00Z"
    }
  ]
}
```

---

### POST /api/v1/notes/:id/revisions/:revision_id/restore
リビジョンから復元

**成功レスポンス** (200 OK)
```json
{
  "status": {
    "code": 200,
    "message": "Note restored from revision"
  },
  "data": { ... }
}
```

---

## エラーコード一覧

| コード | HTTPステータス | 説明 |
|--------|---------------|------|
| `AUTHENTICATION_FAILED` | 401 | 認証失敗 |
| `AUTHORIZATION_FAILED` | 403 | 権限なし |
| `RESOURCE_NOT_FOUND` | 404 | リソースが見つからない |
| `VALIDATION_FAILED` | 422 | バリデーションエラー |
| `PARAMETER_MISSING` | 400 | 必須パラメータ不足 |
| `RATE_LIMIT_EXCEEDED` | 429 | レート制限超過 |
| `INTERNAL_ERROR` | 500 | サーバーエラー |
