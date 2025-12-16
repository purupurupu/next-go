# Go → TypeScript/Hono マイグレーション タスクチェックリスト

## 概要
- **移行元**: Go 1.25 (Echo + GORM)
- **移行先**: TypeScript (Hono + Drizzle)
- **エンドポイント数**: 40
- **テーブル数**: 11

---

## Phase 0: 環境セットアップ

### プロジェクト初期化
- [ ] `backend-hono/` ディレクトリ作成
- [ ] `bun init` 実行
- [ ] 依存パッケージインストール
  - [ ] `hono`
  - [ ] `drizzle-orm`
  - [ ] `drizzle-kit`
  - [ ] `postgres` (node-postgres)
  - [ ] `@hono/zod-validator`
  - [ ] `zod`
  - [ ] `jose` (JWT)
  - [ ] `bcrypt` または `@node-rs/bcrypt`
  - [ ] `@aws-sdk/client-s3`
  - [ ] `sharp` (サムネイル生成)
  - [ ] `uuid`
  - [ ] `pino` (ロギング)
  - [ ] `vitest` (テスト)
  - [ ] `@types/bcrypt`

### ディレクトリ構造作成
- [ ] `src/index.ts` - エントリーポイント
- [ ] `src/routes/` - ルートハンドラー
- [ ] `src/services/` - ビジネスロジック
- [ ] `src/repositories/` - データアクセス層
- [ ] `src/models/` - Drizzle スキーマ
- [ ] `src/validators/` - Zod スキーマ
- [ ] `src/middleware/` - ミドルウェア
- [ ] `src/lib/` - ユーティリティ
- [ ] `drizzle/` - マイグレーション
- [ ] `tests/` - テスト

### Docker設定
- [ ] `backend-hono/Dockerfile` 作成
- [ ] `compose.yml` に backend-hono サービス追加
- [ ] 環境変数設定（DATABASE_URL, JWT_SECRET, S3設定等）

### 基盤コード実装
- [ ] `src/lib/db.ts` - Drizzle DB接続
- [ ] `src/lib/config.ts` - 環境変数読み込み
- [ ] `src/lib/errors.ts` - ApiError クラス定義
- [ ] `src/lib/response.ts` - レスポンスヘルパー
- [ ] `src/index.ts` - Honoアプリ + ルーティング設定

### Drizzle スキーマ定義
- [ ] `src/models/schema.ts` - 全テーブル定義
  - [ ] users テーブル
  - [ ] todos テーブル
  - [ ] categories テーブル
  - [ ] tags テーブル
  - [ ] todo_tags テーブル（中間）
  - [ ] comments テーブル
  - [ ] todo_histories テーブル
  - [ ] files テーブル
  - [ ] notes テーブル
  - [ ] note_revisions テーブル
  - [ ] jwt_denylists テーブル

### マイグレーション
- [ ] `drizzle.config.ts` 設定
- [ ] `drizzle-kit generate` でマイグレーション生成
- [ ] `drizzle-kit migrate` でマイグレーション実行

---

## Phase 1: 認証システム（最優先）

### バリデータ（Zod）
- [ ] `src/validators/auth.ts`
  - [ ] `signUpSchema` - email, password, password_confirmation, name
  - [ ] `signInSchema` - email, password

### Repository
- [ ] `src/repositories/user.ts`
  - [ ] `UserRepositoryInterface` 定義
  - [ ] `findByEmail(email: string)` - メールでユーザー検索
  - [ ] `create(user: NewUser)` - ユーザー作成
  - [ ] `findById(id: number)` - ID検索
- [ ] `src/repositories/jwt-denylist.ts`
  - [ ] `add(jti: string, exp: Date)` - トークン無効化登録
  - [ ] `exists(jti: string)` - 無効化チェック

### Service
- [ ] `src/services/auth.ts`
  - [ ] `signUp(email, password, passwordConfirmation, name)` - ユーザー登録
  - [ ] `signIn(email, password)` - ログイン
  - [ ] `signOut(jti: string, exp: Date)` - ログアウト
  - [ ] `generateToken(user)` - JWT生成（jose使用）
  - [ ] `validateToken(token: string)` - JWT検証

### Middleware
- [ ] `src/middleware/auth.ts`
  - [ ] `jwtAuth()` - JWT認証ミドルウェア
  - [ ] `getCurrentUser(c: Context)` - 現在のユーザー取得
  - [ ] jwt_denylistチェック統合

### Routes
- [ ] `src/routes/auth.ts`
  - [ ] `POST /auth/sign_up` - 新規登録
  - [ ] `POST /auth/sign_in` - ログイン
  - [ ] `DELETE /auth/sign_out` - ログアウト（要認証）

### CORS設定
- [ ] `@hono/cors` ミドルウェア設定
  - [ ] `origin: http://localhost:3000`
  - [ ] `credentials: true`
  - [ ] `exposeHeaders: ['Authorization']`

### テスト
- [ ] 登録テスト（成功・重複エラー・バリデーションエラー）
- [ ] ログインテスト（成功・認証エラー）
- [ ] ログアウトテスト（成功・トークン無効化確認）

### フロントエンド統合確認
- [ ] 登録→ログイン→ログアウトフロー動作確認

---

## Phase 2: Todo基本CRUD（最優先）

### バリデータ
- [ ] `src/validators/todo.ts`
  - [ ] `createTodoSchema` - title, description, priority, status, due_date, category_id, tag_ids
  - [ ] `updateTodoSchema` - 全フィールドオプショナル
  - [ ] `updateOrderSchema` - id, position の配列

### Repository
- [ ] `src/repositories/todo.ts`
  - [ ] `TodoRepositoryInterface` 定義
  - [ ] `findAllByUserId(userId: number)` - 一覧取得（position順）
  - [ ] `findById(id: number, userId: number)` - 詳細取得
  - [ ] `create(todo: NewTodo)` - 作成
  - [ ] `update(id: number, userId: number, data: UpdateTodo)` - 更新
  - [ ] `delete(id: number, userId: number)` - 削除
  - [ ] `updateOrder(updates: {id: number, position: number}[])` - 順序更新
  - [ ] `getMaxPosition(userId: number)` - 最大position取得

### Service
- [ ] `src/services/todo.ts`
  - [ ] `TodoService` クラス
  - [ ] カテゴリーカウント自動更新
  - [ ] タグ関連付け処理

### Routes
- [ ] `src/routes/todos.ts`
  - [ ] `GET /api/v1/todos` - 一覧取得
  - [ ] `POST /api/v1/todos` - 作成
  - [ ] `GET /api/v1/todos/:id` - 詳細取得
  - [ ] `PATCH /api/v1/todos/:id` - 更新
  - [ ] `DELETE /api/v1/todos/:id` - 削除
  - [ ] `PATCH /api/v1/todos/update_order` - 順序一括更新

### バリデーション
- [ ] title: 必須、1-255文字
- [ ] priority: 0-2の範囲（low/medium/high）
- [ ] status: 0-2の範囲（pending/in_progress/completed）
- [ ] due_date: ISO 8601形式

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

#### バリデータ
- [ ] `src/validators/category.ts`
  - [ ] `createCategorySchema` - name, color
  - [ ] `updateCategorySchema`

#### Repository
- [ ] `src/repositories/category.ts`
  - [ ] CRUD操作
  - [ ] `incrementTodosCount(id: number)` - カウント増加
  - [ ] `decrementTodosCount(id: number)` - カウント減少

#### Routes
- [ ] `src/routes/categories.ts`
  - [ ] `GET /api/v1/categories` - 一覧
  - [ ] `POST /api/v1/categories` - 作成
  - [ ] `GET /api/v1/categories/:id` - 詳細
  - [ ] `PATCH /api/v1/categories/:id` - 更新
  - [ ] `DELETE /api/v1/categories/:id` - 削除

#### バリデーション
- [ ] name: 必須、50文字以下、ユーザー内ユニーク
- [ ] color: 必須、HEX形式（#RRGGBB）

### Tag

#### バリデータ
- [ ] `src/validators/tag.ts`
  - [ ] `createTagSchema` - name, color
  - [ ] `updateTagSchema`

#### Repository
- [ ] `src/repositories/tag.ts`
  - [ ] CRUD操作
  - [ ] `syncTodoTags(todoId: number, tagIds: number[])` - タグ同期

#### Routes
- [ ] `src/routes/tags.ts`
  - [ ] `GET /api/v1/tags` - 一覧
  - [ ] `POST /api/v1/tags` - 作成
  - [ ] `GET /api/v1/tags/:id` - 詳細
  - [ ] `PATCH /api/v1/tags/:id` - 更新
  - [ ] `DELETE /api/v1/tags/:id` - 削除

#### バリデーション
- [ ] name: 必須、30文字以下、ユーザー内ユニーク、正規化（小文字+trim）
- [ ] color: オプション、HEX形式

### Todo-Category/Tag連携
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

### バリデータ
- [ ] `src/validators/todo-search.ts`
  - [ ] `searchTodoSchema` - 全クエリパラメータ

### Service
- [ ] `src/services/todo-search.ts`
  - [ ] フィルター条件
    - [ ] q: タイトル・説明のILIKE検索
    - [ ] status: ステータスフィルター（複数対応）
    - [ ] priority: 優先度フィルター
    - [ ] category_id: カテゴリフィルター（-1でカテゴリなし）
    - [ ] tag_ids: タグフィルター
    - [ ] tag_mode: "all" または "any"
    - [ ] due_date_from / due_date_to: 日付範囲
  - [ ] ソート
    - [ ] sort_by: due_date, created_at, updated_at, priority, position, title, status
    - [ ] sort_order: asc, desc
    - [ ] due_dateソートでNULLを最後に配置
  - [ ] ページネーション
    - [ ] page（デフォルト: 1）
    - [ ] per_page（デフォルト: 20、最大100）

### Routes
- [ ] `src/routes/todos.ts` に追加
  - [ ] `GET /api/v1/todos/search`
  - [ ] クエリパラメータ解析（配列形式とカンマ区切り両対応）
  - [ ] ページネーションメタデータ付きレスポンス

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
  }
}
```

### テスト
- [ ] 基本検索テスト
- [ ] テキスト検索テスト
- [ ] 各フィルター条件テスト（status, priority, category, tags, date range）
- [ ] タグAND/OR検索テスト
- [ ] ソートテスト（due_date NULL最後含む）
- [ ] ページネーションテスト
- [ ] 複合条件テスト
- [ ] ユーザースコープテスト

### フロントエンド統合確認
- [ ] 検索ボックス動作
- [ ] フィルター選択
- [ ] ソート切り替え
- [ ] ページネーション

---

## Phase 5: Comment・TodoHistory（中優先）

### Comment

#### バリデータ
- [ ] `src/validators/comment.ts`
  - [ ] `createCommentSchema` - content
  - [ ] `updateCommentSchema` - content

#### Repository
- [ ] `src/repositories/comment.ts`
  - [ ] `findByTodoId(todoId: number)` - 一覧取得（deleted_at IS NULL）
  - [ ] `findById(id: number)` - ID検索
  - [ ] `create(comment: NewComment)` - 作成
  - [ ] `update(id: number, content: string)` - 更新
  - [ ] `softDelete(id: number)` - ソフトデリート

#### ヘルパー関数
- [ ] `isEditable(comment: Comment)` - 15分以内かつ未削除か判定
- [ ] `isOwnedBy(comment: Comment, userId: number)` - 所有者確認

#### Routes
- [ ] `src/routes/comments.ts`
  - [ ] `GET /api/v1/todos/:todo_id/comments` - 一覧
  - [ ] `POST /api/v1/todos/:todo_id/comments` - 作成
  - [ ] `PATCH /api/v1/todos/:todo_id/comments/:id` - 更新（15分制限）
  - [ ] `DELETE /api/v1/todos/:todo_id/comments/:id` - 削除（ソフトデリート）

#### ビジネスルール
- [ ] 作成者のみ編集・削除可能
- [ ] 作成から15分以内のみ編集可能
- [ ] 削除は論理削除（deleted_at設定）
- [ ] content: 必須、1000文字以下

### TodoHistory

#### Repository
- [ ] `src/repositories/todo-history.ts`
  - [ ] `create(history: NewTodoHistory)` - 作成
  - [ ] `findByTodoId(todoId: number, page: number, perPage: number)` - ページネーション付き取得

#### 自動記録（TodoService に統合）
- [ ] Todo作成時 → action: "created"
- [ ] Todo更新時 → action: "updated" + 変更内容（changes JSONB）
- [ ] Todo削除時 → action: "deleted"
- [ ] ステータス変更時 → action: "status_changed"
- [ ] 優先度変更時 → action: "priority_changed"

#### Routes
- [ ] `src/routes/histories.ts`
  - [ ] `GET /api/v1/todos/:todo_id/histories` - 履歴一覧（ページネーション付き）

#### human_readable_change 日本語メッセージ生成
- [ ] `generateHumanReadableChange(history: TodoHistory)` - 日本語変更メッセージ

### テスト
- [ ] Comment CRUD テスト
- [ ] 15分編集制限テスト
- [ ] ソフトデリートテスト
- [ ] 履歴自動記録テスト
- [ ] ユーザースコープテスト

### フロントエンド統合確認
- [ ] コメント表示・投稿
- [ ] コメント編集（15分以内）
- [ ] 履歴表示

---

## Phase 6: ファイルアップロード（中優先）

### バリデータ
- [ ] `src/validators/file.ts`
  - [ ] MIMEタイプチェック
  - [ ] ファイルサイズチェック（最大10MB）

### Storage
- [ ] `src/lib/storage.ts`
  - [ ] `StorageInterface` 定義
  - [ ] `upload(key: string, buffer: Buffer, contentType: string)` - アップロード
  - [ ] `download(key: string)` - ダウンロード
  - [ ] `delete(key: string)` - 削除
  - [ ] `getUrl(key: string, expiry: number)` - Pre-signed URL生成
  - [ ] `exists(key: string)` - 存在確認
- [ ] `src/lib/s3-storage.ts` - S3Storage実装（@aws-sdk/client-s3使用）
  - [ ] バケット自動作成

### Service
- [ ] `src/services/thumbnail.ts`
  - [ ] `ThumbnailService` クラス（Sharp使用）
  - [ ] thumb (300x300) 生成
  - [ ] medium (800x800) 生成
  - [ ] アスペクト比維持
  - [ ] WebP → JPEG 変換対応
- [ ] `src/services/file.ts`
  - [ ] `upload(input)` - アップロード処理
  - [ ] `download(fileId, todoId, userId)` - ダウンロード
  - [ ] `downloadThumbnail(fileId, todoId, userId, size)` - サムネイル取得
  - [ ] `delete(fileId, todoId, userId)` - 削除
  - [ ] `listByTodo(todoId, userId)` - 一覧取得

### Repository
- [ ] `src/repositories/file.ts`
  - [ ] `findById(id: number)` - ID検索
  - [ ] `findByAttachable(type: string, id: number)` - 添付先で一覧取得
  - [ ] `create(file: NewFile)` - 作成
  - [ ] `delete(id: number)` - 削除

### Routes
- [ ] `src/routes/files.ts`
  - [ ] `GET /api/v1/todos/:todo_id/files` - 一覧取得
  - [ ] `POST /api/v1/todos/:todo_id/files` - アップロード（multipart/form-data）
  - [ ] `GET /api/v1/todos/:todo_id/files/:file_id` - ダウンロード
  - [ ] `GET /api/v1/todos/:todo_id/files/:file_id/thumb` - サムネイル
  - [ ] `GET /api/v1/todos/:todo_id/files/:file_id/medium` - 中サイズ
  - [ ] `DELETE /api/v1/todos/:todo_id/files/:file_id` - 削除

### バリデーション
- [ ] ファイルサイズ: 最大10MB
- [ ] 許可MIMEタイプ:
  - [ ] image/jpeg, image/png, image/gif, image/webp
  - [ ] application/pdf, text/plain
  - [ ] MS Office (Word, Excel, PowerPoint)

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

### バリデータ
- [ ] `src/validators/note.ts`
  - [ ] `createNoteSchema` - title, body_md
  - [ ] `updateNoteSchema` - title, body_md, pinned
  - [ ] `searchNoteSchema` - archived, trashed, pinned, page, per_page

### Repository
- [ ] `src/repositories/note.ts`
  - [ ] `findAllByUserId(userId: number, filters)` - 一覧取得（フィルター・ページネーション）
  - [ ] `findById(id: number, userId: number)` - 詳細取得
  - [ ] `create(note: NewNote)` - 作成
  - [ ] `update(id: number, userId: number, data: UpdateNote)` - 更新
  - [ ] `softDelete(id: number, userId: number)` - ソフトデリート（trashed_at設定）
  - [ ] `hardDelete(id: number, userId: number)` - 完全削除
- [ ] `src/repositories/note-revision.ts`
  - [ ] `findByNoteId(noteId: number)` - リビジョン一覧
  - [ ] `findById(id: number)` - ID検索
  - [ ] `create(revision: NewNoteRevision)` - リビジョン作成
  - [ ] `deleteOldest(noteId: number, keepCount: number)` - 古いリビジョン削除

### Service
- [ ] `src/services/note.ts`
  - [ ] `create(input)` - 作成（初期リビジョン作成含む）
  - [ ] `update(id, userId, input)` - 更新（body_md変更時のみリビジョン作成）
  - [ ] `delete(id, userId, force: boolean)` - 削除（ソフト/ハードデリート）
  - [ ] `restoreRevision(noteId, revisionId, userId)` - リビジョン復元
  - [ ] `stripMarkdown(md: string)` - body_plain生成
  - [ ] `enforceRevisionLimit(noteId: number)` - 50件制限

### Routes
- [ ] `src/routes/notes.ts`
  - [ ] `GET /api/v1/notes` - 一覧（フィルター・ページネーション）
  - [ ] `POST /api/v1/notes` - 作成
  - [ ] `GET /api/v1/notes/:id` - 詳細
  - [ ] `PATCH /api/v1/notes/:id` - 更新
  - [ ] `DELETE /api/v1/notes/:id` - 削除（?force=true で完全削除）
  - [ ] `GET /api/v1/notes/:id/revisions` - リビジョン一覧
  - [ ] `POST /api/v1/notes/:id/revisions/:revision_id/restore` - リビジョン復元

### バリデーション
- [ ] title: 150文字以下（任意）
- [ ] body_md: 100,000文字以下（任意）

### テスト
- [ ] Note CRUD テスト（一覧、作成、詳細、更新、削除）
- [ ] フィルターテスト（archived, trashed, pinned）
- [ ] リビジョン作成テスト（body_md変更時のみ）
- [ ] リビジョン復元テスト
- [ ] 50件制限テスト
- [ ] ユーザースコープテスト

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
- [ ] SQLインジェクション対策確認（Drizzleパラメータ化）
- [ ] XSS対策確認
- [ ] 認可チェック漏れ確認
- [ ] レート制限実装（@hono/rate-limit）

### 運用
- [ ] ヘルスチェックエンドポイント `/health`
- [ ] グレースフルシャットダウン実装
- [ ] 構造化ログ設定（pino）
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

## 技術スタック対応表

| 項目 | Go (現行) | TypeScript/Hono (移行先) |
|------|-----------|-------------------------|
| Runtime | Go 1.25 | Bun / Node.js |
| Framework | Echo v4 | Hono |
| ORM | GORM | Drizzle |
| Validation | go-playground/validator v10 | Zod + @hono/zod-validator |
| JWT | golang-jwt/jwt v5 | jose |
| Password | bcrypt (cost=12) | bcrypt / @node-rs/bcrypt |
| S3 | aws-sdk-go-v2 | @aws-sdk/client-s3 |
| Image | disintegration/imaging | Sharp |
| Test | testing + testify | Vitest |
| Logging | zerolog | pino |
| Config | envconfig | dotenv + zod |

---

## エンドポイント一覧（40エンドポイント）

### 認証（3）
| メソッド | パス | 説明 |
|---------|------|------|
| POST | /auth/sign_up | ユーザー登録 |
| POST | /auth/sign_in | ログイン |
| DELETE | /auth/sign_out | ログアウト |

### Todo（7）
| メソッド | パス | 説明 |
|---------|------|------|
| GET | /api/v1/todos | 一覧取得 |
| GET | /api/v1/todos/search | 検索 |
| POST | /api/v1/todos | 作成 |
| GET | /api/v1/todos/:id | 詳細取得 |
| PATCH | /api/v1/todos/:id | 更新 |
| DELETE | /api/v1/todos/:id | 削除 |
| PATCH | /api/v1/todos/update_order | 順序更新 |

### Category（5）
| メソッド | パス | 説明 |
|---------|------|------|
| GET | /api/v1/categories | 一覧取得 |
| POST | /api/v1/categories | 作成 |
| GET | /api/v1/categories/:id | 詳細取得 |
| PATCH | /api/v1/categories/:id | 更新 |
| DELETE | /api/v1/categories/:id | 削除 |

### Tag（5）
| メソッド | パス | 説明 |
|---------|------|------|
| GET | /api/v1/tags | 一覧取得 |
| POST | /api/v1/tags | 作成 |
| GET | /api/v1/tags/:id | 詳細取得 |
| PATCH | /api/v1/tags/:id | 更新 |
| DELETE | /api/v1/tags/:id | 削除 |

### Comment（4）
| メソッド | パス | 説明 |
|---------|------|------|
| GET | /api/v1/todos/:todo_id/comments | 一覧取得 |
| POST | /api/v1/todos/:todo_id/comments | 作成 |
| PATCH | /api/v1/todos/:todo_id/comments/:id | 更新 |
| DELETE | /api/v1/todos/:todo_id/comments/:id | 削除 |

### TodoHistory（1）
| メソッド | パス | 説明 |
|---------|------|------|
| GET | /api/v1/todos/:todo_id/histories | 履歴一覧 |

### File（6）
| メソッド | パス | 説明 |
|---------|------|------|
| GET | /api/v1/todos/:todo_id/files | 一覧取得 |
| POST | /api/v1/todos/:todo_id/files | アップロード |
| GET | /api/v1/todos/:todo_id/files/:file_id | ダウンロード |
| GET | /api/v1/todos/:todo_id/files/:file_id/thumb | サムネイル |
| GET | /api/v1/todos/:todo_id/files/:file_id/medium | 中サイズ |
| DELETE | /api/v1/todos/:todo_id/files/:file_id | 削除 |

### Note（7）
| メソッド | パス | 説明 |
|---------|------|------|
| GET | /api/v1/notes | 一覧取得 |
| POST | /api/v1/notes | 作成 |
| GET | /api/v1/notes/:id | 詳細取得 |
| PATCH | /api/v1/notes/:id | 更新 |
| DELETE | /api/v1/notes/:id | 削除 |
| GET | /api/v1/notes/:id/revisions | リビジョン一覧 |
| POST | /api/v1/notes/:id/revisions/:revision_id/restore | リビジョン復元 |

### Health（1）
| メソッド | パス | 説明 |
|---------|------|------|
| GET | /health | ヘルスチェック |

---

## 参照ドキュメント

| ドキュメント | パス |
|-------------|------|
| Go実装ガイド | `go-implementation-guide.md` |
| Hono実装ガイド | `hono-implementation-guide.md` |
| API仕様書 | `api-specification.md` |
| DB スキーマ | `database-schema.md` |
| 認証仕様 | `authentication.md` |
| ビジネスロジック | `business-logic.md` |
| エラーハンドリング | `error-handling.md` |
| Docker設定 | `docker-setup.md` |
