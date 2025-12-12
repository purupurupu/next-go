# Rails → Rust/Go バックエンド移管ドキュメント

このディレクトリには、現在のRailsバックエンドをRustまたはGoに移管するための技術仕様書が含まれています。

## ドキュメント一覧

| ファイル | 説明 |
|---------|------|
| [api-specification.md](./api-specification.md) | 全APIエンドポイント仕様（34エンドポイント） |
| [database-schema.md](./database-schema.md) | データベーススキーマ・ER図・インデックス |
| [authentication.md](./authentication.md) | JWT認証フロー・トークン仕様 |
| [business-logic.md](./business-logic.md) | バリデーション・コールバック・ビジネスルール |
| [error-handling.md](./error-handling.md) | エラーコード体系・レスポンス形式 |
| [docker-setup.md](./docker-setup.md) | Docker設定・Dockerfile・compose.yml |
| [rust-implementation-guide.md](./rust-implementation-guide.md) | Rust向け実装ガイド（Axum + SeaORM） |
| [go-implementation-guide.md](./go-implementation-guide.md) | Go向け実装ガイド（Echo + GORM） |

## 現在のシステム概要

### 技術スタック（Rails）
- **Runtime**: Ruby 3.2.5
- **Framework**: Rails 7.1.3+ (API-only)
- **Database**: PostgreSQL 15
- **Authentication**: Devise + devise-jwt
- **File Storage**: Active Storage
- **Serialization**: JSONAPI::Serializer

### エンドポイント数
| リソース | エンドポイント数 |
|---------|-----------------|
| 認証（Auth） | 3 |
| Todo | 9 |
| Category | 5 |
| Tag | 5 |
| Comment | 4 |
| History | 1 |
| Note | 7 |
| **合計** | **34** |

### データベーステーブル
- `users` - ユーザー
- `todos` - タスク
- `categories` - カテゴリ
- `tags` - タグ
- `todo_tags` - タスク-タグ関連
- `comments` - コメント（ポリモーフィック）
- `todo_histories` - 変更履歴
- `notes` - ノート
- `note_revisions` - ノートリビジョン
- `jwt_denylist` - 無効化トークン
- `active_storage_*` - ファイル添付

## 推奨技術スタック比較

| 領域 | Rails現状 | Rust推奨 | Go推奨 |
|------|----------|----------|--------|
| Runtime | Ruby 3.2.5 | Rust (stable) | Go 1.22+ |
| Framework | Rails 7.1.3 | Axum | Echo |
| ORM | ActiveRecord | SeaORM | GORM |
| Validation | ActiveModel | validator | go-playground/validator |
| JWT | devise-jwt | jsonwebtoken | golang-jwt/jwt |
| File Upload | Active Storage | tower-http + S3 | multipart + S3 |
| Serialization | JSONAPI::Serializer | serde | encoding/json |

## 移行優先順位

| Phase | 内容 | 優先度 | 依存関係 |
|-------|------|--------|----------|
| 1 | 認証システム | 最高 | なし |
| 2 | User・Todo基本CRUD | 最高 | Phase 1 |
| 3 | Category・Tag CRUD | 高 | Phase 2 |
| 4 | Todo検索・フィルタリング | 高 | Phase 2 |
| 5 | Comment・TodoHistory | 中 | Phase 2 |
| 6 | ファイルアップロード | 中 | Phase 2 |
| 7 | Note・NoteRevision | 低 | Phase 1 |

## 主要なビジネスルール

1. **ユーザースコープ**: 全リソースは`user_id`でスコープされ、他ユーザーのデータにアクセス不可
2. **履歴追跡**: Todo変更時に自動的に`TodoHistory`に記録
3. **ソフトデリート**: コメントは`deleted_at`で論理削除
4. **編集制限**: コメントは作成から15分以内のみ編集可能
5. **ファイル制限**: 最大10MB、特定MIMEタイプのみ許可
6. **カウンターキャッシュ**: `categories.todos_count`で効率的な集計

## フロントエンドとの連携

現在のフロントエンド（Next.js）は以下のベースURLでAPIを呼び出しています：

```
http://localhost:3001
```

移行後も同じエンドポイント構造を維持することで、フロントエンドの変更を最小限に抑えられます。

### CORS設定
- Origin: `http://localhost:3000`
- Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD
- Credentials: true
- Expose: Authorization header

## 使用方法

1. 各ドキュメントを参照し、実装要件を理解
2. 移行先の言語（Rust/Go）に応じた実装ガイドを参照
3. Phase順に実装を進める
4. 各フェーズでフロントエンドとの統合テストを実施
