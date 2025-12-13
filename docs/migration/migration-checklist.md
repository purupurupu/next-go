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

## Phase 2: User・Todo基本CRUD（最優先）

### モデル
- [ ] `internal/model/todo.go`
  - [ ] Todo構造体定義
  - [ ] Priority enum (0:low, 1:medium, 2:high)
  - [ ] Status enum (0:pending, 1:in_progress, 2:completed)
  - [ ] `BeforeCreate()` - position自動設定
  - [ ] リレーション定義（User, Category, Tags）

### Repository
- [ ] `internal/repository/todo.go`
  - [ ] `FindAllByUserID(userID int64)` - 一覧取得
  - [ ] `FindByID(id, userID int64)` - 詳細取得
  - [ ] `Create(todo *model.Todo)` - 作成
  - [ ] `Update(todo *model.Todo)` - 更新
  - [ ] `Delete(id, userID int64)` - 削除
  - [ ] `UpdateOrder(updates []OrderUpdate)` - 順序更新

### Handler
- [ ] `internal/handler/todo.go`
  - [ ] `GET /api/v1/todos` - 一覧取得
  - [ ] `POST /api/v1/todos` - 作成
  - [ ] `GET /api/v1/todos/:id` - 詳細取得
  - [ ] `PATCH /api/v1/todos/:id` - 更新
  - [ ] `DELETE /api/v1/todos/:id` - 削除
  - [ ] `PATCH /api/v1/todos/update_order` - 順序一括更新

### バリデーション
- [ ] title: 必須
- [ ] priority: 0-2の範囲
- [ ] status: 0-2の範囲
- [ ] due_date: 過去日付禁止（作成時）

### ユーザースコープ
- [ ] 全クエリに `user_id = ?` 条件追加
- [ ] 他ユーザーのTodoにアクセス不可を確認

### テスト
- [ ] CRUD全操作テスト
- [ ] ユーザースコープテスト（他ユーザーデータアクセス拒否）
- [ ] バリデーションエラーテスト
- [ ] 順序更新テスト

### フロントエンド統合確認
- [ ] Todo一覧表示
- [ ] Todo作成・編集・削除
- [ ] ドラッグ＆ドロップ順序変更

---

## Phase 3: Category・Tag CRUD（高優先）

### Category

#### モデル
- [ ] `internal/model/category.go`
  - [ ] Category構造体
  - [ ] todos_count カウンターキャッシュ
  - [ ] User リレーション

#### Repository
- [ ] `internal/repository/category.go`
  - [ ] CRUD操作
  - [ ] カウンターキャッシュ更新ロジック

#### Handler
- [ ] `internal/handler/category.go`
  - [ ] `GET /api/v1/categories` - 一覧
  - [ ] `POST /api/v1/categories` - 作成
  - [ ] `GET /api/v1/categories/:id` - 詳細
  - [ ] `PATCH /api/v1/categories/:id` - 更新
  - [ ] `DELETE /api/v1/categories/:id` - 削除

#### バリデーション
- [ ] name: 必須、50文字以下、ユーザー内ユニーク
- [ ] color: 必須、HEX形式（#RRGGBB）

### Tag

#### モデル
- [ ] `internal/model/tag.go`
  - [ ] Tag構造体
  - [ ] User リレーション
- [ ] `internal/model/todo_tag.go`
  - [ ] TodoTag中間テーブル

#### Repository
- [ ] `internal/repository/tag.go`
  - [ ] CRUD操作
  - [ ] Todo紐付け操作

#### Handler
- [ ] `internal/handler/tag.go`
  - [ ] `GET /api/v1/tags` - 一覧
  - [ ] `POST /api/v1/tags` - 作成
  - [ ] `GET /api/v1/tags/:id` - 詳細
  - [ ] `PATCH /api/v1/tags/:id` - 更新
  - [ ] `DELETE /api/v1/tags/:id` - 削除

#### バリデーション
- [ ] name: 必須、30文字以下、ユーザー内ユニーク、正規化（小文字+trim）
- [ ] color: 必須、HEX形式

### Todo-Category/Tag連携
- [ ] `PATCH /api/v1/todos/:id/tags` - タグ更新
- [ ] Todo作成・更新時のcategory_id設定
- [ ] Todo作成・更新時のtag_ids設定
- [ ] 他ユーザーのCategory/Tag使用禁止

### テスト
- [ ] Category CRUD テスト
- [ ] Tag CRUD テスト
- [ ] カウンターキャッシュテスト
- [ ] ユニーク制約テスト

### フロントエンド統合確認
- [ ] カテゴリ管理画面
- [ ] タグ管理画面
- [ ] Todo編集でのカテゴリ・タグ選択

---

## Phase 4: Todo検索・フィルタリング（高優先）

### Service
- [ ] `internal/service/todo_search.go`
  - [ ] フィルター条件
    - [ ] q: タイトル・説明のILIKE検索
    - [ ] status: ステータスフィルター
    - [ ] priority: 優先度フィルター
    - [ ] category_id: カテゴリフィルター
    - [ ] tag_ids: タグフィルター
    - [ ] tag_mode: "all" または "any"
    - [ ] due_date_from / due_date_to: 日付範囲
  - [ ] ソート
    - [ ] sort_by: due_date, created_at, updated_at, priority, position, title
    - [ ] sort_order: asc, desc
  - [ ] ページネーション
    - [ ] page
    - [ ] per_page（最大100）

### Handler
- [ ] `GET /api/v1/todos/search`
  - [ ] クエリパラメータ解析
  - [ ] 検索実行
  - [ ] ページネーションメタデータ付きレスポンス

### レスポンス形式
```json
{
  "todos": [...],
  "meta": {
    "total": 100,
    "current_page": 1,
    "total_pages": 5,
    "per_page": 20
  }
}
```

### テスト
- [ ] 各フィルター条件テスト
- [ ] ソートテスト
- [ ] ページネーションテスト
- [ ] 複合条件テスト

### フロントエンド統合確認
- [ ] 検索ボックス動作
- [ ] フィルター選択
- [ ] ソート切り替え
- [ ] ページネーション

---

## Phase 5: Comment・TodoHistory（中優先）

### Comment

#### モデル
- [ ] `internal/model/comment.go`
  - [ ] Comment構造体
  - [ ] ポリモーフィック関連（commentable_type, commentable_id）
  - [ ] deleted_at（ソフトデリート）

#### Repository
- [ ] `internal/repository/comment.go`
  - [ ] 一覧取得（deleted_at IS NULL）
  - [ ] 作成
  - [ ] 更新（15分以内チェック）
  - [ ] ソフトデリート

#### Handler
- [ ] `internal/handler/comment.go`
  - [ ] `GET /api/v1/todos/:todo_id/comments` - 一覧
  - [ ] `POST /api/v1/todos/:todo_id/comments` - 作成
  - [ ] `PATCH /api/v1/todos/:todo_id/comments/:id` - 更新
  - [ ] `DELETE /api/v1/todos/:todo_id/comments/:id` - 削除

#### ビジネスルール
- [ ] 作成者のみ編集・削除可能
- [ ] 作成から15分以内のみ編集可能
- [ ] 削除は論理削除（deleted_at設定）
- [ ] content: 必須、1000文字以下

### TodoHistory

#### モデル
- [ ] `internal/model/todo_history.go`
  - [ ] TodoHistory構造体
  - [ ] action enum (created, updated, deleted, status_changed, priority_changed)
  - [ ] changes JSONB

#### 自動記録
- [ ] Todo作成時 → action: "created"
- [ ] Todo更新時 → action: "updated" + 変更内容
- [ ] Todo削除時 → action: "deleted"
- [ ] ステータス変更時 → action: "status_changed"
- [ ] 優先度変更時 → action: "priority_changed"

#### Handler
- [ ] `GET /api/v1/todos/:todo_id/histories` - 履歴一覧

### テスト
- [ ] Comment CRUD テスト
- [ ] 15分編集制限テスト
- [ ] ソフトデリートテスト
- [ ] 履歴自動記録テスト

### フロントエンド統合確認
- [ ] コメント表示・投稿
- [ ] コメント編集（15分以内）
- [ ] 履歴表示

---

## Phase 6: ファイルアップロード（中優先）

### ストレージ設定
- [ ] ローカルストレージまたはS3設定
- [ ] アップロードディレクトリ設定

### モデル拡張
- [ ] Todo.Files関連追加
- [ ] ファイルメタデータ構造体

### Handler
- [ ] `POST /api/v1/todos/:id/files` - ファイルアップロード
- [ ] `DELETE /api/v1/todos/:id/files/:file_id` - ファイル削除
- [ ] `GET /api/v1/todos/:id/files/:file_id` - ファイルダウンロード

### バリデーション
- [ ] ファイルサイズ: 最大10MB
- [ ] 許可MIMEタイプ:
  - [ ] image/jpeg, image/png, image/gif, image/webp
  - [ ] application/pdf
  - [ ] text/plain
  - [ ] application/msword
  - [ ] application/vnd.openxmlformats-officedocument.wordprocessingml.document

### テスト
- [ ] アップロードテスト
- [ ] サイズ制限テスト
- [ ] MIMEタイプ制限テスト
- [ ] 削除テスト

### フロントエンド統合確認
- [ ] ファイルアップロードUI
- [ ] ファイル一覧表示
- [ ] ファイルダウンロード
- [ ] ファイル削除

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
