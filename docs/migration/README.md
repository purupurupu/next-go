# バックエンド移行ドキュメント

このディレクトリには、バックエンド移行のための技術仕様書が含まれています。

## 移行状況

| 移行パス | 状況 |
|---------|------|
| Rails → Go | **完了** ✅ |
| Go → Hono | ガイド作成中 |

---

## Go 実装（現行）

現在のバックエンドは Go + Echo で実装されています。

### 技術スタック
- **Language**: Go 1.25
- **Framework**: Echo v4
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt v5)
- **Validation**: go-playground/validator v10
- **File Storage**: RustFS (S3互換)

### 実装ガイド
- [go-implementation-guide.md](./go-implementation-guide.md) - Go 実装パターン（1,249行）

---

## Hono 移行計画

Go バックエンドを Hono (TypeScript) に移行するためのガイドです。

### 技術スタック（移行先）
- **Runtime**: Bun / Node.js / Cloudflare Workers
- **Framework**: Hono
- **ORM**: Drizzle / Prisma
- **Validation**: Zod
- **Authentication**: JWT (hono/jwt)

### 移行ガイド
- [hono-implementation-guide.md](./hono-implementation-guide.md) - Go → Hono 移行ガイド

---

## ドキュメント一覧

| ファイル | 説明 |
|---------|------|
| [api-specification.md](./api-specification.md) | 全APIエンドポイント仕様 |
| [database-schema.md](./database-schema.md) | データベーススキーマ・ER図 |
| [authentication.md](./authentication.md) | JWT認証フロー |
| [business-logic.md](./business-logic.md) | ビジネスルール |
| [error-handling.md](./error-handling.md) | エラーコード体系 |
| [docker-setup.md](./docker-setup.md) | Docker設定 |
| [go-implementation-guide.md](./go-implementation-guide.md) | Go実装ガイド |
| [hono-implementation-guide.md](./hono-implementation-guide.md) | Hono移行ガイド |
| [migration-checklist.md](./migration-checklist.md) | 移行チェックリスト |

---

## APIエンドポイント

| リソース | エンドポイント数 |
|---------|-----------------|
| 認証（Auth） | 3 |
| Todo | 9 |
| Category | 5 |
| Tag | 5 |
| Comment | 4 |
| History | 1 |
| File | 6 |
| Note | 7 |
| **合計** | **40** |

---

## 主要なビジネスルール

1. **ユーザースコープ**: 全リソースは`user_id`でスコープ
2. **履歴追跡**: Todo変更時に`TodoHistory`に自動記録
3. **ソフトデリート**: コメントは`deleted_at`で論理削除
4. **編集制限**: コメントは作成から15分以内のみ編集可能
5. **ファイル制限**: 最大10MB、ホワイトリスト方式
6. **カウンターキャッシュ**: `categories.todos_count`
7. **ノートリビジョン**: body_md変更時に自動作成、50件制限

---

## フロントエンド連携

```
Frontend (Next.js): http://localhost:3000
Backend API:        http://localhost:3001
```

### CORS設定
- Origin: `http://localhost:3000`
- Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
- Credentials: true
- Expose: Authorization header
