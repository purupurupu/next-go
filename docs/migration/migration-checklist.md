# Rails → Go マイグレーション タスクチェックリスト

## 概要
- **移行元**: Rails 7.1.3+ (Ruby 3.2.5)
- **移行先**: Go 1.22+ (Echo + GORM)
- **エンドポイント数**: 34
- **テーブル数**: 11

---

## Phase 0: 環境セットアップ

### プロジェクト初期化
- [x] `backend-go/` ディレクトリ作成
- [x] `go mod init todo-api` 実行
- [x] 依存パッケージインストール
  - [x] `github.com/labstack/echo/v4`
  - [x] `gorm.io/gorm`
  - [x] `gorm.io/driver/postgres`
  - [x] `github.com/go-playground/validator/v10`
  - [x] `github.com/golang-jwt/jwt/v5`
  - [x] `golang.org/x/crypto`
  - [x] `github.com/google/uuid`
  - [x] `github.com/joho/godotenv`
  - [x] `github.com/kelseyhightower/envconfig`
  - [x] `github.com/rs/zerolog`
  - [x] `github.com/stretchr/testify`

### ディレクトリ構造作成
- [x] `cmd/api/main.go`
- [x] `internal/config/`
- [x] `internal/handler/`
- [x] `internal/middleware/`
- [x] `internal/model/`
- [x] `internal/repository/`
- [x] `internal/service/`
- [x] `internal/validator/`
- [x] `internal/errors/`
- [x] `pkg/response/`
- [x] `pkg/database/`
- [x] `db/migrations/`

### Docker設定
- [x] `backend-go/Dockerfile` 作成
- [x] `compose.yml` に backend-go サービス追加
- [x] `.air.toml` ホットリロード設定
- [x] 環境変数ファイル設定

### 基盤コード実装
- [x] `internal/config/config.go` - 設定読み込み
- [x] `pkg/database/database.go` - DB接続
- [x] `internal/errors/api_error.go` - エラー定義
- [x] `pkg/response/response.go` - レスポンスヘルパー
- [x] `internal/validator/validator.go` - バリデーション
- [x] `cmd/api/main.go` - エントリポイント（空のルーター）

---

## Phase 1: 認証システム（最優先） ✅ 完了

### モデル
- [x] `internal/model/user.go`
  - [x] User構造体定義
  - [x] `SetPassword()` bcryptハッシュ化
  - [x] `CheckPassword()` パスワード検証
  - [x] `TableName()` テーブル名設定
- [x] `internal/model/jwt_denylist.go`
  - [x] JwtDenylist構造体定義
  - [x] `IsRevoked()` トークン無効化チェック

### Repository
- [x] `internal/repository/user.go`
  - [x] `FindByEmail(email string)` - メールでユーザー検索
  - [x] `Create(user *model.User)` - ユーザー作成
  - [x] `FindByID(id int64)` - ID検索
- [x] `internal/repository/jwt_denylist.go`
  - [x] `Add(jti string, exp time.Time)` - トークン無効化登録
  - [x] `Exists(jti string)` - 無効化チェック

### Service
- [x] `internal/service/auth.go`
  - [x] `SignUp(email, password, name)` - ユーザー登録
  - [x] `SignIn(email, password)` - ログイン
  - [x] `SignOut(jti string)` - ログアウト
  - [x] `GenerateToken(user *model.User)` - JWT生成
  - [x] `ValidateToken(token string)` - JWT検証

### Middleware
- [x] `internal/middleware/auth.go`
  - [x] `JWTAuth()` - JWT認証ミドルウェア
  - [x] `GetCurrentUser(c echo.Context)` - 現在のユーザー取得
  - [x] jwt_denylistチェック統合

### Handler
- [x] `internal/handler/auth.go`
  - [x] `POST /auth/sign_up` - 新規登録
  - [x] `POST /auth/sign_in` - ログイン
  - [x] `DELETE /auth/sign_out` - ログアウト

### CORS設定
- [x] `Origin: http://localhost:3000`
- [x] `Credentials: true`
- [x] `Expose: Authorization`

### テスト
- [x] 登録テスト（成功・重複エラー・バリデーションエラー）
- [x] ログインテスト（成功・認証エラー）
- [x] ログアウトテスト（成功・トークン無効化確認）

### フロントエンド統合確認
- [ ] 登録→ログイン→ログアウトフロー動作確認

---

## Phase 2: User・Todo基本CRUD（最優先） ✅ 完了

### モデル
- [x] `internal/model/todo.go`
  - [x] Todo構造体定義
  - [x] Priority enum (0:low, 1:medium, 2:high)
  - [x] Status enum (0:pending, 1:in_progress, 2:completed)
  - [x] `BeforeCreate()` - position自動設定
  - [x] リレーション定義（User, Category, Tags）

### Repository
- [x] `internal/repository/todo.go`
  - [x] `FindAllByUserID(userID int64)` - 一覧取得
  - [x] `FindByID(id, userID int64)` - 詳細取得
  - [x] `Create(todo *model.Todo)` - 作成
  - [x] `Update(todo *model.Todo)` - 更新
  - [x] `Delete(id, userID int64)` - 削除
  - [x] `UpdateOrder(updates []OrderUpdate)` - 順序更新

### Handler
- [x] `internal/handler/todo.go`
  - [x] `GET /api/v1/todos` - 一覧取得
  - [x] `POST /api/v1/todos` - 作成
  - [x] `GET /api/v1/todos/:id` - 詳細取得
  - [x] `PATCH /api/v1/todos/:id` - 更新
  - [x] `DELETE /api/v1/todos/:id` - 削除
  - [x] `PATCH /api/v1/todos/update_order` - 順序一括更新

### バリデーション
- [x] title: 必須
- [x] priority: 0-2の範囲
- [x] status: 0-2の範囲
- [x] due_date: 過去日付禁止（作成時）

### ユーザースコープ
- [x] 全クエリに `user_id = ?` 条件追加
- [x] 他ユーザーのTodoにアクセス不可を確認

### テスト
- [x] CRUD全操作テスト
- [x] ユーザースコープテスト（他ユーザーデータアクセス拒否）
- [x] バリデーションエラーテスト
- [x] 順序更新テスト

### フロントエンド統合確認
- [ ] Todo一覧表示
- [ ] Todo作成・編集・削除
- [ ] ドラッグ＆ドロップ順序変更

---

## Phase 3: Category・Tag CRUD（高優先） ✅ 完了

### Category

#### モデル
- [x] `internal/model/category.go`
  - [x] Category構造体
  - [x] todos_count カウンターキャッシュ
  - [x] User リレーション

#### Repository
- [x] `internal/repository/category.go`
  - [x] CRUD操作
  - [x] カウンターキャッシュ更新ロジック

#### Handler
- [x] `internal/handler/category.go`
  - [x] `GET /api/v1/categories` - 一覧
  - [x] `POST /api/v1/categories` - 作成
  - [x] `GET /api/v1/categories/:id` - 詳細
  - [x] `PATCH /api/v1/categories/:id` - 更新
  - [x] `DELETE /api/v1/categories/:id` - 削除

#### バリデーション
- [x] name: 必須、50文字以下、ユーザー内ユニーク
- [x] color: 必須、HEX形式（#RRGGBB）

### Tag

#### モデル
- [x] `internal/model/tag.go`
  - [x] Tag構造体
  - [x] User リレーション
- [x] `internal/model/todo_tag.go`
  - [x] TodoTag中間テーブル

#### Repository
- [x] `internal/repository/tag.go`
  - [x] CRUD操作
  - [x] Todo紐付け操作

#### Handler
- [x] `internal/handler/tag.go`
  - [x] `GET /api/v1/tags` - 一覧
  - [x] `POST /api/v1/tags` - 作成
  - [x] `GET /api/v1/tags/:id` - 詳細
  - [x] `PATCH /api/v1/tags/:id` - 更新
  - [x] `DELETE /api/v1/tags/:id` - 削除

#### バリデーション
- [x] name: 必須、30文字以下、ユーザー内ユニーク、正規化（小文字+trim）
- [x] color: 必須、HEX形式

### Todo-Category/Tag連携
- [x] Todo作成・更新時のcategory_id設定
- [x] Todo作成・更新時のtag_ids設定
- [x] 他ユーザーのCategory/Tag使用禁止
- [ ] `PATCH /api/v1/todos/:id/tags` - タグ更新（未実装）

### テスト
- [x] Category CRUD テスト
- [x] Tag CRUD テスト
- [x] カウンターキャッシュテスト
- [x] ユニーク制約テスト

### フロントエンド統合確認
- [ ] カテゴリ管理画面
- [ ] タグ管理画面
- [ ] Todo編集でのカテゴリ・タグ選択

---

## Phase 4: Todo検索・フィルタリング（高優先） ✅ 完了

### Service
- [x] `internal/service/todo.go` に検索機能追加
  - [x] フィルター条件
    - [x] q: タイトル・説明のILIKE検索
    - [x] status: ステータスフィルター（複数対応）
    - [x] priority: 優先度フィルター
    - [x] category_id: カテゴリフィルター（-1でカテゴリなし）
    - [x] tag_ids: タグフィルター
    - [x] tag_mode: "all" または "any"
    - [x] due_date_from / due_date_to: 日付範囲
  - [x] ソート
    - [x] sort_by: due_date, created_at, updated_at, priority, position, title, status
    - [x] sort_order: asc, desc
    - [x] due_dateソートでNULLを最後に配置
  - [x] ページネーション
    - [x] page（デフォルト: 1）
    - [x] per_page（デフォルト: 20、最大100）

### Handler
- [x] `GET /api/v1/todos/search`
  - [x] クエリパラメータ解析（配列形式とカンマ区切り両対応）
  - [x] 検索実行
  - [x] ページネーションメタデータ付きレスポンス
  - [x] 空結果時のサジェスション機能

### レスポンス形式
```json
{
  "data": [...],
  "meta": {
    "total": 100,
    "current_page": 1,
    "total_pages": 5,
    "per_page": 20,
    "filters_applied": {}
  },
  "suggestions": []
}
```

### テスト
- [x] 基本検索テスト
- [x] テキスト検索テスト
- [x] 各フィルター条件テスト（status, priority, category, tags, date range）
- [x] タグAND/OR検索テスト
- [x] ソートテスト（due_date NULL最後含む）
- [x] ページネーションテスト
- [x] 複合条件テスト
- [x] ユーザースコープテスト

### フロントエンド統合確認
- [ ] 検索ボックス動作
- [ ] フィルター選択
- [ ] ソート切り替え
- [ ] ページネーション

---

## Phase 5: Comment・TodoHistory（中優先） ✅ 完了

### Comment

#### モデル
- [x] `internal/model/comment.go`
  - [x] Comment構造体
  - [x] ポリモーフィック関連（commentable_type, commentable_id）
  - [x] deleted_at（ソフトデリート）
  - [x] `IsEditable()` - 15分以内かつ未削除か判定
  - [x] `IsOwnedBy(userID)` - 所有者確認

#### Repository
- [x] `internal/repository/comment.go`
  - [x] 一覧取得（deleted_at IS NULL）
  - [x] 作成
  - [x] 更新（15分以内チェック）
  - [x] ソフトデリート
  - [x] `ExistsByID()` - 存在確認

#### Handler
- [x] `internal/handler/comment.go`
  - [x] `GET /api/v1/todos/:todo_id/comments` - 一覧
  - [x] `POST /api/v1/todos/:todo_id/comments` - 作成
  - [x] `PATCH /api/v1/todos/:todo_id/comments/:id` - 更新
  - [x] `DELETE /api/v1/todos/:todo_id/comments/:id` - 削除
  - [x] レスポンスに `editable` フィールドを含める

#### ビジネスルール
- [x] 作成者のみ編集・削除可能
- [x] 作成から15分以内のみ編集可能
- [x] 削除は論理削除（deleted_at設定）
- [x] content: 必須、1000文字以下

### TodoHistory

#### モデル
- [x] `internal/model/todo_history.go`
  - [x] TodoHistory構造体
  - [x] action enum (created, updated, deleted, status_changed, priority_changed)
  - [x] changes JSONB

#### 自動記録
- [x] Todo作成時 → action: "created"
- [x] Todo更新時 → action: "updated" + 変更内容
- [x] Todo削除時 → action: "deleted"
- [x] ステータス変更時 → action: "status_changed"
- [x] 優先度変更時 → action: "priority_changed"

#### Handler
- [x] `GET /api/v1/todos/:todo_id/histories` - 履歴一覧
- [x] ページネーション対応
- [x] `human_readable_change` 日本語メッセージ生成

### テスト
- [x] Comment CRUD テスト（22テストケース）
- [x] 15分編集制限テスト
- [x] ソフトデリートテスト
- [x] 履歴自動記録テスト（10テストケース）
- [x] ユーザースコープテスト

### フロントエンド統合確認
- [ ] コメント表示・投稿
- [ ] コメント編集（15分以内）
- [ ] 履歴表示

---

## Phase 6: ファイルアップロード（中優先） ✅ 完了

### インフラ設定
- [x] `compose.yml` に RustFS (S3互換ストレージ) サービス追加
- [x] `internal/config/config.go` に S3 設定追加
  - [x] S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY
  - [x] S3_USE_PATH_STYLE (パススタイルURL対応)

### モデル
- [x] `internal/model/file.go`
  - [x] File構造体（ポリモーフィック関連: attachable_type, attachable_id）
  - [x] FileType enum (image, document, other)
  - [x] AllowedMimeTypes マップ
  - [x] MaxFileSize 定数 (10MB)
  - [x] `IsImage()` - 画像判定
  - [x] `IsOwnedBy(userID)` - 所有者確認
  - [x] `GetFileType(contentType)` - MIMEタイプからFileType判定

### Storage
- [x] `internal/storage/storage.go`
  - [x] Storage インターフェース定義
  - [x] Upload, Download, Delete, GetURL, Exists
- [x] `internal/storage/s3.go`
  - [x] S3Storage 実装 (aws-sdk-go-v2)
  - [x] バケット自動作成

### Repository
- [x] `internal/repository/file.go`
  - [x] `FindByID(id)` - ID検索
  - [x] `FindByAttachable(type, id)` - 添付先で一覧取得
  - [x] `Create(file)` - 作成
  - [x] `Delete(id)` - 削除

### Service
- [x] `internal/service/thumbnail.go`
  - [x] ThumbnailService (github.com/disintegration/imaging)
  - [x] thumb (300x300) 生成
  - [x] medium (800x800) 生成
  - [x] アスペクト比維持
  - [x] WebP → JPEG 変換対応
- [x] `internal/service/file.go`
  - [x] `Upload(ctx, input)` - アップロード処理
  - [x] `Download(ctx, fileID, todoID, userID)` - ダウンロード
  - [x] `DownloadThumbnail(ctx, fileID, todoID, userID, size)` - サムネイル取得
  - [x] `Delete(ctx, fileID, todoID, userID)` - 削除
  - [x] `ListByTodo(ctx, todoID, userID)` - 一覧取得

### Handler
- [x] `internal/handler/file.go`
  - [x] `GET /api/v1/todos/:todo_id/files` - 一覧取得
  - [x] `POST /api/v1/todos/:todo_id/files` - アップロード
  - [x] `GET /api/v1/todos/:todo_id/files/:file_id` - ダウンロード
  - [x] `GET /api/v1/todos/:todo_id/files/:file_id/thumb` - サムネイル
  - [x] `GET /api/v1/todos/:todo_id/files/:file_id/medium` - 中サイズ
  - [x] `DELETE /api/v1/todos/:todo_id/files/:file_id` - 削除

### バリデーション
- [x] ファイルサイズ: 最大10MB
- [x] 許可MIMEタイプ:
  - [x] image/jpeg, image/png, image/gif, image/webp
  - [x] application/pdf, text/plain
  - [x] MS Office (Word, Excel, PowerPoint)

### フロントエンド実装
- [x] `types/todo.ts` - TodoFile 型を Go バックエンド形式に更新
- [x] `lib/api-client.ts` - ファイル操作 API メソッド追加
  - [x] getFiles, uploadTodoFile, deleteFile
  - [x] downloadFile, downloadThumbnail
- [x] `hooks/useFileUpload.ts` - アップロード状態管理フック
- [x] `components/FileThumbnail.tsx` - サムネイル表示
- [x] `components/FilePreviewModal.tsx` - 画像プレビュー
  - [x] ズーム機能 (0.5x - 3x)
  - [x] 矢印キーナビゲーション
  - [x] ダウンロードボタン
- [x] `components/AttachmentList.tsx` - 添付ファイル一覧
  - [x] 画像グリッド表示
  - [x] プレビューモーダル統合

### テスト
- [ ] アップロードテスト
- [ ] サイズ制限テスト
- [ ] MIMEタイプ制限テスト
- [ ] サムネイル生成テスト
- [ ] 削除テスト
- [ ] ユーザースコープテスト

### フロントエンド統合確認
- [ ] ファイルアップロードUI動作確認
- [ ] サムネイル表示確認
- [ ] 画像プレビュー動作確認
- [ ] ファイルダウンロード確認
- [ ] ファイル削除確認

---

## Phase 7: Note・NoteRevision（低優先）

### モデル
- [ ] `internal/model/note.go`
  - [ ] Note構造体
  - [ ] body_md（Markdown本文）
  - [ ] body_plain（プレーンテキスト変換）
- [ ] `internal/model/note_revision.go`
  - [ ] NoteRevision構造体
  - [ ] リビジョン管理（最大50件）

### Repository
- [ ] `internal/repository/note.go`
  - [ ] CRUD操作
  - [ ] リビジョン保存
  - [ ] リビジョン復元
  - [ ] 古いリビジョン削除（50件超過時）

### Handler
- [ ] `internal/handler/note.go`
  - [ ] `GET /api/v1/notes` - 一覧
  - [ ] `POST /api/v1/notes` - 作成
  - [ ] `GET /api/v1/notes/:id` - 詳細
  - [ ] `PATCH /api/v1/notes/:id` - 更新
  - [ ] `DELETE /api/v1/notes/:id` - 削除
  - [ ] `GET /api/v1/notes/:id/revisions` - リビジョン一覧
  - [ ] `POST /api/v1/notes/:id/revisions/:revision_id/restore` - リビジョン復元

### バリデーション
- [ ] title: 必須
- [ ] body_md: 100,000文字以下

### テスト
- [ ] Note CRUD テスト
- [ ] リビジョン作成テスト
- [ ] リビジョン復元テスト
- [ ] 50件制限テスト

### フロントエンド統合確認
- [ ] ノート一覧
- [ ] ノート編集（Markdown）
- [ ] リビジョン履歴表示
- [ ] リビジョン復元

---

## 最終確認・本番準備

### パフォーマンス
- [ ] N+1クエリ確認・解消
- [ ] インデックス確認
- [ ] コネクションプール設定

### セキュリティ
- [ ] SQLインジェクション対策確認
- [ ] XSS対策確認
- [ ] 認可チェック漏れ確認
- [ ] レート制限実装

### 運用
- [ ] ヘルスチェックエンドポイント `/health`
- [ ] グレースフルシャットダウン実装
- [ ] 構造化ログ設定
- [ ] エラートラッキング設定

### ドキュメント
- [ ] API変更点ドキュメント
- [ ] 環境構築手順
- [ ] デプロイ手順

### マイグレーション戦略
- [ ] 並行運用期間の計画
- [ ] データ移行スクリプト（必要な場合）
- [ ] フロントエンド切り替え手順
- [ ] ロールバック計画

---

## 参照ドキュメント

| ドキュメント | パス |
|-------------|------|
| Go実装ガイド | `go-implementation-guide.md` |
| API仕様書 | `api-specification.md` |
| DB スキーマ | `database-schema.md` |
| 認証仕様 | `authentication.md` |
| ビジネスロジック | `business-logic.md` |
| エラーハンドリング | `error-handling.md` |
| Docker設定 | `docker-setup.md` |
