# Backend Architecture (Go)

## Technology Stack

- **Language**: Go 1.23
- **Framework**: Echo v4 (Web framework)
- **ORM**: GORM
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Authentication**: JWT (golang-jwt/jwt v5)
- **Validation**: go-playground/validator v10
- **Logging**: zerolog
- **Hot Reload**: Air (development)

## Directory Structure

```
backend/
├── cmd/api/
│   └── main.go                 # エントリポイント、サーバー設定
├── internal/
│   ├── config/
│   │   └── config.go           # 環境変数からの設定読み込み
│   ├── constants/
│   │   └── constants.go        # 定数の一元管理
│   ├── handler/
│   │   ├── auth.go             # 認証ハンドラ
│   │   ├── todo.go             # Todo CRUD
│   │   ├── category.go         # Category CRUD
│   │   ├── tag.go              # Tag CRUD
│   │   └── helpers.go          # 共通ヘルパー関数
│   ├── middleware/
│   │   └── auth.go             # JWT認証ミドルウェア
│   ├── model/
│   │   ├── user.go             # Userモデル
│   │   ├── todo.go             # Todoモデル
│   │   ├── category.go         # Categoryモデル
│   │   ├── tag.go              # Tagモデル
│   │   └── jwt_denylist.go     # JWTトークン無効化リスト
│   ├── repository/
│   │   ├── interfaces.go       # リポジトリインターフェース定義
│   │   ├── user.go             # Userリポジトリ
│   │   ├── todo.go             # Todoリポジトリ
│   │   ├── category.go         # Categoryリポジトリ
│   │   ├── tag.go              # Tagリポジトリ
│   │   └── jwt_denylist.go     # JWT Denylistリポジトリ
│   ├── service/
│   │   └── auth.go             # 認証サービス（ビジネスロジック）
│   ├── testutil/
│   │   ├── fixture.go          # TestFixtureパターン
│   │   └── assertions.go       # テスト用アサーション
│   ├── validator/
│   │   └── validator.go        # カスタムバリデーション
│   └── errors/
│       └── api_error.go        # APIエラーハンドリング
└── pkg/
    ├── database/
    │   └── database.go         # DB接続管理
    ├── response/
    │   └── response.go         # 統一レスポンス形式
    └── util/
        ├── pointers.go         # ポインタユーティリティ
        └── time.go             # 時間ユーティリティ
```

---

## Architecture Design Decisions

### レイヤードアーキテクチャ

本プロジェクトでは3層アーキテクチャを採用しています：

```
┌─────────────────────────────────────────────────────────────┐
│  Handler (Presentation Layer)                                │
│  - HTTPリクエスト/レスポンス処理                              │
│  - リクエストのバインド・バリデーション                        │
│  - 認証情報の取得                                            │
│  - レスポンスの整形                                          │
├─────────────────────────────────────────────────────────────┤
│  Service (Business Logic Layer)                              │
│  - ドメインルールの実装                                       │
│  - トランザクション管理                                       │
│  - 複数リポジトリの調整                                       │
│  - 外部サービスとの連携                                       │
├─────────────────────────────────────────────────────────────┤
│  Repository (Data Access Layer)                              │
│  - データベース操作                                          │
│  - クエリ実行                                                │
│  - データの永続化                                            │
└─────────────────────────────────────────────────────────────┘
```

### Service層の使用方針

**Service層を使用するケース（推奨）:**
- 複雑なビジネスロジックがある場合
- 複数リポジトリをまたぐトランザクション処理
- 外部サービスとの連携（メール送信、通知など）
- ドメインルールの実装（例：パスワードハッシュ化、JWT生成）

**現在の実装:**
```
AuthHandler → AuthService → UserRepository, JwtDenylistRepository
  └─ 認証ロジック（パスワードハッシュ化、JWT生成/検証）があるためService経由

TodoHandler → TodoRepository, CategoryRepository (直接)
  └─ 単純なCRUD + カウンターキャッシュ更新

CategoryHandler → CategoryRepository (直接)
  └─ 単純なCRUDのみ

TagHandler → TagRepository (直接)
  └─ 単純なCRUDのみ
```

**Service層を追加するタイミング:**
- Todo完了時に通知送信
- 履歴記録（TodoHistory）
- ポイント付与などのドメインロジック追加時

この方針は実用的なアプローチ（Pragmatic Approach）であり、単純なCRUDに対してService層を追加するとパススルーコードが増えるため、必要になった時点で追加します。

### Repository Interface

テスト容易性のため、リポジトリはインターフェースを定義しています：

```go
// internal/repository/interfaces.go
type UserRepositoryInterface interface {
    FindByEmail(email string) (*model.User, error)
    Create(user *model.User) error
    FindByID(id int64) (*model.User, error)
    ExistsByEmail(email string) (bool, error)
}

type TodoRepositoryInterface interface {
    FindAllByUserID(userID int64) ([]model.Todo, error)
    FindByID(id, userID int64) (*model.Todo, error)
    Create(todo *model.Todo) error
    Update(todo *model.Todo) error
    Delete(id, userID int64) error
    UpdateOrder(userID int64, todoIDs []int64) error
}

type CategoryRepositoryInterface interface {
    FindAllByUserID(userID int64) ([]model.Category, error)
    FindByID(id, userID int64) (*model.Category, error)
    Create(category *model.Category) error
    Update(category *model.Category) error
    Delete(id, userID int64) error
    ExistsByName(name string, userID int64, excludeID *int64) (bool, error)
    IncrementTodosCount(categoryID int64) error
    DecrementTodosCount(categoryID int64) error
}

type TagRepositoryInterface interface {
    FindAllByUserID(userID int64) ([]model.Tag, error)
    FindByID(id, userID int64) (*model.Tag, error)
    Create(tag *model.Tag) error
    Update(tag *model.Tag) error
    Delete(id, userID int64) error
    ExistsByName(name string, userID int64, excludeID *int64) (bool, error)
}
```

これにより：
- モック化が可能（単体テスト）
- 実装の差し替えが容易
- 依存性逆転の原則に準拠

---

## Helper Functions

コードの重複を減らすため、共通処理をヘルパー関数として抽出しています。

### Handler Helpers (`internal/handler/helpers.go`)

```go
// IDパラメータの解析
func ParseIDParam(c echo.Context, name string) (int64, error)

// リクエストのバインドとバリデーション（ジェネリクス使用）
func BindAndValidate[T any](c echo.Context, req *T) error

// 認証済みユーザーの取得
func GetCurrentUserOrFail(c echo.Context) (*middleware.CurrentUser, error)
```

### Pointer Utilities (`pkg/util/pointers.go`)

```go
// null安全な値取得
func DerefString(s *string, defaultVal string) string
func DerefInt(i *int, defaultVal int) int
func DerefInt64(i *int64, defaultVal int64) int64
func DerefBool(b *bool, defaultVal bool) bool

// ポインタ生成（ジェネリクス）
func Ptr[T any](v T) *T
```

### Time Utilities (`pkg/util/time.go`)

```go
// 日付フォーマット
func FormatDate(t *time.Time) *string
func FormatDateTime(t time.Time) string

// 日付パース
func ParseDate(s string) (*time.Time, error)

// 日付比較
func IsBeforeToday(t time.Time) bool
```

---

## Authentication Flow

### JWT Token Flow

```
1. User Login (POST /auth/sign_in)
   ↓
2. Validate credentials (bcrypt)
   ↓
3. Generate JWT token (with jti for revocation)
   ↓
4. Return token in response body
   ↓
5. Client stores token (localStorage)
   ↓
6. Client sends token in Authorization header
   ↓
7. Middleware validates token on each request
   ↓
8. Check jwt_denylist for revoked tokens
```

### Token Management
- **Generation**: AuthService.GenerateToken
- **Storage**: Client-side (localStorage)
- **Validation**: JWTAuth middleware
- **Revocation**: jwt_denylist table
- **Expiration**: Configurable (default: 24 hours)

---

## API Design

### Endpoints

```
# Auth (Public)
POST   /auth/sign_up     # ユーザー登録
POST   /auth/sign_in     # ログイン
DELETE /auth/sign_out    # ログアウト（要認証）

# API v1 (Protected)
# Todos
GET    /api/v1/todos              # Todo一覧
POST   /api/v1/todos              # Todo作成
GET    /api/v1/todos/:id          # Todo詳細
PATCH  /api/v1/todos/:id          # Todo更新
DELETE /api/v1/todos/:id          # Todo削除
PATCH  /api/v1/todos/update_order # 順序更新

# Categories
GET    /api/v1/categories         # Category一覧
POST   /api/v1/categories         # Category作成
GET    /api/v1/categories/:id     # Category詳細
PATCH  /api/v1/categories/:id     # Category更新
DELETE /api/v1/categories/:id     # Category削除

# Tags
GET    /api/v1/tags               # Tag一覧
POST   /api/v1/tags               # Tag作成
GET    /api/v1/tags/:id           # Tag詳細
PATCH  /api/v1/tags/:id           # Tag更新
DELETE /api/v1/tags/:id           # Tag削除
```

### Response Format

**Success Response:**
```json
{
  "status": {
    "code": 200,
    "message": "Success"
  },
  "data": { ... }
}
```

**Error Response:**
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation failed",
    "details": {
      "validation_errors": {
        "title": ["required"]
      }
    }
  }
}
```

---

## Configuration

環境変数による設定管理（`internal/config/config.go`）：

| 環境変数 | 説明 | デフォルト |
|---------|------|-----------|
| `PORT` | サーバーポート | 3000 |
| `DATABASE_URL` | PostgreSQL接続文字列 | (required) |
| `JWT_SECRET` | JWT署名キー | (required) |
| `JWT_EXPIRATION_HOURS` | JWT有効期限（時間） | 24 |
| `ENV` | 環境 (development/production) | development |
| `CORS_ALLOW_ORIGINS` | 許可オリジン（カンマ区切り） | http://localhost:3000 |
| `CORS_MAX_AGE` | CORSプリフライトキャッシュ秒数 | 86400 |

---

## Testing

### TestFixture Pattern

テストセットアップの重複を排除するため、TestFixtureパターンを採用：

```go
// internal/testutil/fixture.go
type TestFixture struct {
    T               *testing.T
    DB              *gorm.DB
    Echo            *echo.Echo
    UserRepo        *repository.UserRepository
    TodoRepo        *repository.TodoRepository
    CategoryRepo    *repository.CategoryRepository
    TagRepo         *repository.TagRepository
    AuthHandler     *handler.AuthHandler
    TodoHandler     *handler.TodoHandler
    CategoryHandler *handler.CategoryHandler
    TagHandler      *handler.TagHandler
}

// 使用例
func TestTodoCreate(t *testing.T) {
    f := testutil.SetupTestFixture(t)
    user, token := f.CreateUser("test@example.com")

    rec, _ := f.CallAuth(token, "POST", "/api/v1/todos",
        `{"title":"Test"}`, f.TodoHandler.Create)

    assert.Equal(t, http.StatusCreated, rec.Code)
}
```

### Assertion Helpers

```go
// internal/testutil/assertions.go
func JSONResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]any
func ExtractStatusCode(response map[string]any) int
func ExtractData(response map[string]any) map[string]any

// ジェネリックヘルパー（Todo, Category, Tag共通）
func ExtractResource[T string](response map[string]any, key T) map[string]any
func ExtractResourceFromData[T string](response map[string]any, key T) map[string]any
func ExtractResources[T string](response map[string]any, key T) []any
func ResourceAt(resources []any, index int) map[string]any
```

---

## Security

1. **Authentication**: 全API v1エンドポイントはJWTトークン必須
2. **Authorization**: ユーザーは自分のデータのみアクセス可能
3. **CORS**: 環境変数で許可オリジンを設定
4. **Password**: bcryptでハッシュ化
5. **SQL Injection**: GORMのパラメータ化クエリで防止

---

## Performance

1. **Database Indexes**:
   - `todos`: user_id, position, category_id
   - `categories`: (user_id, name) 複合ユニークインデックス
   - `tags`: (user_id, name) 複合ユニークインデックス
   - `todo_tags`: (todo_id, tag_id) 複合ユニークインデックス
2. **Counter Cache**: `categories.todos_count` をIncrement/Decrementで効率的に更新
3. **Connection Pool**: GORMのSetMaxOpenConns, SetMaxIdleConns設定
4. **Eager Loading**: Preload()でN+1問題を回避
5. **Graceful Shutdown**: 10秒のシャットダウンタイムアウト
