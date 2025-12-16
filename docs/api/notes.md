# Notes API

Markdown ノート機能の API 仕様です。リビジョン管理による履歴追跡をサポートしています。

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/notes` | ノート一覧取得 |
| POST | `/api/v1/notes` | ノート作成 |
| GET | `/api/v1/notes/:id` | ノート詳細取得 |
| PATCH | `/api/v1/notes/:id` | ノート更新 |
| DELETE | `/api/v1/notes/:id` | ノート削除 |
| GET | `/api/v1/notes/:id/revisions` | リビジョン一覧 |
| POST | `/api/v1/notes/:id/revisions/:revision_id/restore` | リビジョン復元 |

---

## Note Object

```json
{
  "id": 1,
  "title": "開発メモ",
  "body_md": "# Hello\n\nMarkdown content here.",
  "pinned": false,
  "archived": false,
  "trashed": false,
  "archived_at": null,
  "trashed_at": null,
  "last_edited_at": "2024-01-01T10:00:00Z",
  "created_at": "2024-01-01T09:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

---

## List Notes

```
GET /api/v1/notes
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| q | string | 検索クエリ（title, body_plain を検索） |
| pinned | boolean | ピン留めでフィルタ |
| archived | boolean | アーカイブでフィルタ |
| trashed | boolean | ゴミ箱でフィルタ |
| page | integer | ページ番号（default: 1） |
| per_page | integer | 1ページあたりの件数（default: 20, max: 100） |

### Response

```json
{
  "data": [
    {
      "id": 1,
      "title": "開発メモ",
      "body_md": "# Hello...",
      "pinned": true,
      "archived": false,
      "trashed": false,
      "archived_at": null,
      "trashed_at": null,
      "last_edited_at": "2024-01-01T10:00:00Z",
      "created_at": "2024-01-01T09:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ],
  "meta": {
    "total": 10,
    "current_page": 1,
    "total_pages": 1,
    "per_page": 20
  }
}
```

---

## Create Note

```
POST /api/v1/notes
```

### Request Body

```json
{
  "note": {
    "title": "新しいノート",
    "body_md": "# 本文\n\nMarkdown で記述",
    "pinned": false
  }
}
```

### Response (201 Created)

```json
{
  "id": 1,
  "title": "新しいノート",
  "body_md": "# 本文\n\nMarkdown で記述",
  "pinned": false,
  "archived": false,
  "trashed": false,
  "archived_at": null,
  "trashed_at": null,
  "last_edited_at": "2024-01-01T10:00:00Z",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

**Note**: 作成時に初期リビジョンが自動作成されます。

---

## Show Note

```
GET /api/v1/notes/:id
```

### Response (200 OK)

```json
{
  "id": 1,
  "title": "開発メモ",
  "body_md": "# Hello\n\nContent...",
  "pinned": true,
  "archived": false,
  "trashed": false,
  "archived_at": null,
  "trashed_at": null,
  "last_edited_at": "2024-01-01T10:00:00Z",
  "created_at": "2024-01-01T09:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

---

## Update Note

```
PATCH /api/v1/notes/:id
```

### Request Body

```json
{
  "note": {
    "title": "更新されたタイトル",
    "body_md": "# 更新された本文",
    "pinned": true,
    "archived": false,
    "trashed": false
  }
}
```

### Response (200 OK)

Note オブジェクトを返却。

### Business Rules

- `body_md` が変更された場合のみ、新しいリビジョンが作成されます
- `title` のみの変更ではリビジョンは作成されません
- `pinned`, `archived`, `trashed` の変更では `last_edited_at` は更新されません
- リビジョンが 50 件を超えると、古いものから自動削除されます

---

## Delete Note

```
DELETE /api/v1/notes/:id
DELETE /api/v1/notes/:id?force=true
```

### Parameters

| Parameter | Description |
|-----------|-------------|
| force | `true`: 完全削除、`false`/省略: ゴミ箱へ移動 |

### Response (204 No Content)

---

## List Revisions

```
GET /api/v1/notes/:id/revisions
```

### Query Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| page | integer | ページ番号（default: 1） |
| per_page | integer | 1ページあたりの件数（default: 20, max: 100） |

### Response

```json
{
  "data": [
    {
      "id": 5,
      "note_id": 1,
      "title": "開発メモ",
      "body_md": "# Hello\n\n更新版",
      "created_at": "2024-01-01T12:00:00Z"
    },
    {
      "id": 4,
      "note_id": 1,
      "title": "開発メモ",
      "body_md": "# Hello\n\n前のバージョン",
      "created_at": "2024-01-01T11:00:00Z"
    }
  ],
  "meta": {
    "total": 5,
    "current_page": 1,
    "total_pages": 1,
    "per_page": 20
  }
}
```

---

## Restore Revision

```
POST /api/v1/notes/:id/revisions/:revision_id/restore
```

### Response (200 OK)

復元後の Note オブジェクトを返却。

### Business Rules

- 復元時、現在の状態が新しいリビジョンとして保存されます
- 指定したリビジョンの `title` と `body_md` がノートに適用されます

---

## Error Responses

### 404 Not Found

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Note with ID '123' not found"
  }
}
```

### 422 Unprocessable Entity

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "title": "must be at most 150 characters"
    }
  }
}
```

---

## Validation Rules

| Field | Rule |
|-------|------|
| title | 最大 150 文字（nullable） |
| body_md | 最大 100,000 文字（nullable） |
