# Go実装ガイド

## 概要

このドキュメントでは、RailsバックエンドをGoに移行する際の推奨技術スタックと実装パターンを説明します。

---

## 推奨技術スタック

| 領域 | 推奨ライブラリ | 代替案 |
|------|--------------|--------|
| Webフレームワーク | Echo | Gin, Chi, Fiber |
| ORM | GORM | Ent, sqlx |
| バリデーション | go-playground/validator | ozzo-validation |
| JWT | golang-jwt/jwt | - |
| パスワードハッシュ | golang.org/x/crypto/bcrypt | argon2 |
| 環境変数 | godotenv + envconfig | viper |
| ログ | zerolog | zap, logrus |
| テスト | testing + testify | - |
| モック | gomock | mockery |
| ホットリロード | air | - |

---

## go.mod

```go
module todo-api

go 1.22

require (
    github.com/labstack/echo/v4 v4.11.4
    gorm.io/gorm v1.25.5
    gorm.io/driver/postgres v1.5.4
    github.com/go-playground/validator/v10 v10.17.0
    github.com/golang-jwt/jwt/v5 v5.2.0
    golang.org/x/crypto v0.18.0
    github.com/google/uuid v1.5.0
    github.com/joho/godotenv v1.5.1
    github.com/kelseyhightower/envconfig v1.4.0
    github.com/rs/zerolog v1.31.0
    github.com/stretchr/testify v1.8.4
)
```

---

## ディレクトリ構成

```
cmd/
└── api/
    └── main.go              # エントリポイント
internal/
├── config/
│   └── config.go            # 設定管理
├── handler/
│   ├── auth.go              # 認証ハンドラ
│   ├── todo.go              # Todo CRUD
│   ├── category.go          # Category CRUD
│   ├── tag.go               # Tag CRUD
│   ├── comment.go           # Comment CRUD
│   └── note.go              # Note CRUD
├── middleware/
│   ├── auth.go              # JWT認証ミドルウェア
│   ├── error.go             # エラーハンドリング
│   └── request_id.go        # リクエストID
├── model/
│   ├── user.go
│   ├── todo.go
│   ├── category.go
│   ├── tag.go
│   ├── todo_tag.go
│   ├── comment.go
│   ├── todo_history.go
│   ├── note.go
│   ├── note_revision.go
│   └── jwt_denylist.go
├── repository/
│   ├── user.go
│   ├── todo.go
│   ├── category.go
│   ├── tag.go
│   ├── comment.go
│   └── note.go
├── service/
│   ├── auth.go
│   └── todo_search.go
├── validator/
│   └── validator.go
└── errors/
    └── api_error.go
pkg/
├── response/
│   └── response.go
└── database/
    └── database.go
```

---

## 実装パターン

### cmd/api/main.go

```go
package main

import (
    "log"
    "os"

    "github.com/joho/godotenv"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"

    "todo-api/internal/config"
    "todo-api/internal/handler"
    authMiddleware "todo-api/internal/middleware"
    "todo-api/pkg/database"
)

func main() {
    // 環境変数読み込み
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // 設定読み込み
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // DB接続
    db, err := database.Connect(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Echo初期化
    e := echo.New()

    // ミドルウェア
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.RequestID())
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
        AllowHeaders:     []string{echo.HeaderAuthorization, echo.HeaderContentType},
        ExposeHeaders:    []string{echo.HeaderAuthorization},
        AllowCredentials: true,
    }))

    // ハンドラ初期化
    authHandler := handler.NewAuthHandler(db, cfg)
    todoHandler := handler.NewTodoHandler(db)
    categoryHandler := handler.NewCategoryHandler(db)
    tagHandler := handler.NewTagHandler(db)
    commentHandler := handler.NewCommentHandler(db)
    noteHandler := handler.NewNoteHandler(db)

    // 認証ルート（公開）
    auth := e.Group("/auth")
    auth.POST("/sign_up", authHandler.SignUp)
    auth.POST("/sign_in", authHandler.SignIn)
    auth.DELETE("/sign_out", authHandler.SignOut, authMiddleware.JWTAuth(cfg))

    // API v1（認証必須）
    api := e.Group("/api/v1", authMiddleware.JWTAuth(cfg))

    // Todos
    api.GET("/todos", todoHandler.List)
    api.GET("/todos/search", todoHandler.Search)
    api.POST("/todos", todoHandler.Create)
    api.GET("/todos/:id", todoHandler.Get)
    api.PATCH("/todos/:id", todoHandler.Update)
    api.DELETE("/todos/:id", todoHandler.Delete)
    api.PATCH("/todos/update_order", todoHandler.UpdateOrder)
    api.PATCH("/todos/:id/tags", todoHandler.UpdateTags)
    api.DELETE("/todos/:id/files/:file_id", todoHandler.DeleteFile)

    // Comments
    api.GET("/todos/:todo_id/comments", commentHandler.List)
    api.POST("/todos/:todo_id/comments", commentHandler.Create)
    api.PATCH("/todos/:todo_id/comments/:id", commentHandler.Update)
    api.DELETE("/todos/:todo_id/comments/:id", commentHandler.Delete)

    // Categories
    api.GET("/categories", categoryHandler.List)
    api.POST("/categories", categoryHandler.Create)
    api.GET("/categories/:id", categoryHandler.Get)
    api.PATCH("/categories/:id", categoryHandler.Update)
    api.DELETE("/categories/:id", categoryHandler.Delete)

    // Tags
    api.GET("/tags", tagHandler.List)
    api.POST("/tags", tagHandler.Create)
    api.GET("/tags/:id", tagHandler.Get)
    api.PATCH("/tags/:id", tagHandler.Update)
    api.DELETE("/tags/:id", tagHandler.Delete)

    // Notes
    api.GET("/notes", noteHandler.List)
    api.POST("/notes", noteHandler.Create)
    api.GET("/notes/:id", noteHandler.Get)
    api.PATCH("/notes/:id", noteHandler.Update)
    api.DELETE("/notes/:id", noteHandler.Delete)
    api.GET("/notes/:id/revisions", noteHandler.ListRevisions)
    api.POST("/notes/:id/revisions/:revision_id/restore", noteHandler.RestoreRevision)

    // サーバー起動
    port := os.Getenv("PORT")
    if port == "" {
        port = "3001"
    }
    e.Logger.Fatal(e.Start(":" + port))
}
```

### internal/config/config.go

```go
package config

import (
    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    DatabaseURL         string `envconfig:"DATABASE_URL" required:"true"`
    JWTSecret           string `envconfig:"JWT_SECRET" required:"true"`
    JWTExpirationHours  int    `envconfig:"JWT_EXPIRATION_HOURS" default:"24"`
}

func Load() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

### internal/errors/api_error.go

```go
package errors

import (
    "net/http"
    "time"

    "github.com/google/uuid"
    "github.com/labstack/echo/v4"
)

type ApiError struct {
    Code       string      `json:"code"`
    Message    string      `json:"message"`
    Details    interface{} `json:"details,omitempty"`
    RequestID  string      `json:"request_id"`
    Timestamp  string      `json:"timestamp"`
    StatusCode int         `json:"-"`
}

func (e *ApiError) Error() string {
    return e.Message
}

func NewApiError(code string, message string, statusCode int, details interface{}) *ApiError {
    return &ApiError{
        Code:       code,
        Message:    message,
        Details:    details,
        RequestID:  uuid.New().String(),
        Timestamp:  time.Now().UTC().Format(time.RFC3339),
        StatusCode: statusCode,
    }
}

func AuthenticationFailed(message string) *ApiError {
    if message == "" {
        message = "Authentication failed"
    }
    return NewApiError("AUTHENTICATION_FAILED", message, http.StatusUnauthorized, nil)
}

func AuthorizationFailed(resource, action string) *ApiError {
    return NewApiError("AUTHORIZATION_FAILED", "Not authorized", http.StatusForbidden, map[string]string{
        "resource": resource,
        "action":   action,
    })
}

func NotFound(resource string, id interface{}) *ApiError {
    return NewApiError("RESOURCE_NOT_FOUND", "Resource not found", http.StatusNotFound, map[string]interface{}{
        "resource": resource,
        "id":       id,
    })
}

func ValidationFailed(errors interface{}) *ApiError {
    return NewApiError("VALIDATION_FAILED", "Validation failed. Please check your input.", http.StatusUnprocessableEntity, map[string]interface{}{
        "validation_errors": errors,
    })
}

func InternalError() *ApiError {
    return NewApiError("INTERNAL_ERROR", "An unexpected error occurred", http.StatusInternalServerError, nil)
}

// Echoエラーハンドラ
func ErrorHandler(err error, c echo.Context) {
    if apiErr, ok := err.(*ApiError); ok {
        c.JSON(apiErr.StatusCode, map[string]interface{}{
            "error": apiErr,
        })
        return
    }

    // 予期しないエラー
    c.JSON(http.StatusInternalServerError, map[string]interface{}{
        "error": InternalError(),
    })
}
```

### internal/model/todo.go

```go
package model

import (
    "time"

    "gorm.io/gorm"
)

type Priority int

const (
    PriorityLow    Priority = 0
    PriorityMedium Priority = 1
    PriorityHigh   Priority = 2
)

type Status int

const (
    StatusPending    Status = 0
    StatusInProgress Status = 1
    StatusCompleted  Status = 2
)

type Todo struct {
    ID          int64          `gorm:"primaryKey" json:"id"`
    UserID      int64          `gorm:"not null;index" json:"user_id"`
    CategoryID  *int64         `gorm:"index" json:"category_id"`
    Title       string         `gorm:"not null" json:"title"`
    Description *string        `json:"description"`
    Completed   bool           `gorm:"default:false" json:"completed"`
    Position    *int           `json:"position"`
    Priority    Priority       `gorm:"default:1;not null" json:"priority"`
    Status      Status         `gorm:"default:0;not null" json:"status"`
    DueDate     *time.Time     `gorm:"type:date" json:"due_date"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`

    // リレーション
    User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Category   *Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
    Tags       []Tag       `gorm:"many2many:todo_tags" json:"tags,omitempty"`
    Comments   []Comment   `gorm:"polymorphic:Commentable" json:"comments,omitempty"`
    Histories  []TodoHistory `gorm:"foreignKey:TodoID" json:"histories,omitempty"`
}

func (Todo) TableName() string {
    return "todos"
}

// BeforeCreate: position自動設定
func (t *Todo) BeforeCreate(tx *gorm.DB) error {
    if t.Position == nil {
        var maxPosition int
        tx.Model(&Todo{}).Where("user_id = ?", t.UserID).Select("COALESCE(MAX(position), 0)").Scan(&maxPosition)
        newPosition := maxPosition + 1
        t.Position = &newPosition
    }
    return nil
}
```

### internal/model/user.go

```go
package model

import (
    "time"

    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type User struct {
    ID                 int64     `gorm:"primaryKey" json:"id"`
    Email              string    `gorm:"uniqueIndex;not null" json:"email"`
    EncryptedPassword  string    `gorm:"not null" json:"-"`
    ResetPasswordToken *string   `gorm:"uniqueIndex" json:"-"`
    Name               *string   `json:"name"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`

    // リレーション
    Todos      []Todo      `gorm:"foreignKey:UserID" json:"-"`
    Categories []Category  `gorm:"foreignKey:UserID" json:"-"`
    Tags       []Tag       `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string {
    return "users"
}

// パスワードハッシュ化
func (u *User) SetPassword(password string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.EncryptedPassword = string(hash)
    return nil
}

// パスワード検証
func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password))
    return err == nil
}
```

### internal/middleware/auth.go

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/golang-jwt/jwt/v5"
    "github.com/labstack/echo/v4"

    "todo-api/internal/config"
    "todo-api/internal/errors"
)

type JWTClaims struct {
    Sub string `json:"sub"`
    Jti string `json:"jti"`
    jwt.RegisteredClaims
}

type CurrentUser struct {
    ID    int64
    Email string
    Name  string
}

const CurrentUserKey = "current_user"

func JWTAuth(cfg *config.Config) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Authorizationヘッダー取得
            authHeader := c.Request().Header.Get("Authorization")
            if authHeader == "" {
                return errors.AuthenticationFailed("Missing authorization header")
            }

            // Bearerトークン抽出
            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) != 2 || parts[0] != "Bearer" {
                return errors.AuthenticationFailed("Invalid authorization format")
            }
            tokenString := parts[1]

            // JWT検証
            token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
                return []byte(cfg.JWTSecret), nil
            })

            if err != nil {
                return errors.AuthenticationFailed("Invalid token")
            }

            claims, ok := token.Claims.(*JWTClaims)
            if !ok || !token.Valid {
                return errors.AuthenticationFailed("Invalid token claims")
            }

            // ユーザーID取得
            var userID int64
            if _, err := fmt.Sscanf(claims.Sub, "%d", &userID); err != nil {
                return errors.AuthenticationFailed("Invalid user id in token")
            }

            // TODO: DBからユーザー取得 & jwt_denylistチェック

            // コンテキストに設定
            c.Set(CurrentUserKey, &CurrentUser{
                ID: userID,
            })

            return next(c)
        }
    }
}

func GetCurrentUser(c echo.Context) *CurrentUser {
    user, ok := c.Get(CurrentUserKey).(*CurrentUser)
    if !ok {
        return nil
    }
    return user
}
```

### internal/handler/todo.go

```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
    "gorm.io/gorm"

    "todo-api/internal/errors"
    "todo-api/internal/middleware"
    "todo-api/internal/model"
    "todo-api/pkg/response"
)

type TodoHandler struct {
    db *gorm.DB
}

func NewTodoHandler(db *gorm.DB) *TodoHandler {
    return &TodoHandler{db: db}
}

type CreateTodoRequest struct {
    Title       string  `json:"title" validate:"required"`
    Description *string `json:"description"`
    Priority    *int    `json:"priority" validate:"omitempty,min=0,max=2"`
    Status      *int    `json:"status" validate:"omitempty,min=0,max=2"`
    DueDate     *string `json:"due_date"`
    CategoryID  *int64  `json:"category_id"`
    TagIDs      []int64 `json:"tag_ids"`
}

func (h *TodoHandler) List(c echo.Context) error {
    currentUser := middleware.GetCurrentUser(c)

    var todos []model.Todo
    result := h.db.
        Where("user_id = ?", currentUser.ID).
        Preload("Category").
        Preload("Tags").
        Order("position ASC").
        Find(&todos)

    if result.Error != nil {
        return errors.InternalError()
    }

    return response.Success(c, todos)
}

func (h *TodoHandler) Get(c echo.Context) error {
    currentUser := middleware.GetCurrentUser(c)

    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return errors.NotFound("Todo", c.Param("id"))
    }

    var todo model.Todo
    result := h.db.
        Where("id = ? AND user_id = ?", id, currentUser.ID).
        Preload("Category").
        Preload("Tags").
        First(&todo)

    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            return errors.NotFound("Todo", id)
        }
        return errors.InternalError()
    }

    return response.Success(c, todo)
}

func (h *TodoHandler) Create(c echo.Context) error {
    currentUser := middleware.GetCurrentUser(c)

    var req CreateTodoRequest
    if err := c.Bind(&req); err != nil {
        return errors.ValidationFailed(map[string][]string{
            "body": {"Invalid request body"},
        })
    }

    // バリデーション
    if err := c.Validate(req); err != nil {
        return errors.ValidationFailed(formatValidationErrors(err))
    }

    todo := model.Todo{
        UserID:      currentUser.ID,
        Title:       req.Title,
        Description: req.Description,
        CategoryID:  req.CategoryID,
    }

    if req.Priority != nil {
        todo.Priority = model.Priority(*req.Priority)
    }
    if req.Status != nil {
        todo.Status = model.Status(*req.Status)
    }

    // トランザクション
    err := h.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(&todo).Error; err != nil {
            return err
        }

        // タグ紐付け
        if len(req.TagIDs) > 0 {
            var tags []model.Tag
            tx.Where("id IN ? AND user_id = ?", req.TagIDs, currentUser.ID).Find(&tags)
            if err := tx.Model(&todo).Association("Tags").Replace(tags); err != nil {
                return err
            }
        }

        return nil
    })

    if err != nil {
        return errors.InternalError()
    }

    return response.Created(c, todo, "Todo created successfully")
}

func (h *TodoHandler) Update(c echo.Context) error {
    // ... 実装
    return nil
}

func (h *TodoHandler) Delete(c echo.Context) error {
    currentUser := middleware.GetCurrentUser(c)

    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return errors.NotFound("Todo", c.Param("id"))
    }

    result := h.db.Where("id = ? AND user_id = ?", id, currentUser.ID).Delete(&model.Todo{})
    if result.RowsAffected == 0 {
        return errors.NotFound("Todo", id)
    }

    return c.NoContent(http.StatusNoContent)
}

func (h *TodoHandler) Search(c echo.Context) error {
    // TODO: 検索サービス実装
    return nil
}

func (h *TodoHandler) UpdateOrder(c echo.Context) error {
    // TODO: 順序更新実装
    return nil
}

func (h *TodoHandler) UpdateTags(c echo.Context) error {
    // TODO: タグ更新実装
    return nil
}

func (h *TodoHandler) DeleteFile(c echo.Context) error {
    // TODO: ファイル削除実装
    return nil
}
```

### pkg/response/response.go

```go
package response

import (
    "net/http"

    "github.com/labstack/echo/v4"
)

type Response struct {
    Status Status      `json:"status"`
    Data   interface{} `json:"data"`
    Meta   *Meta       `json:"meta,omitempty"`
}

type Status struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

type Meta struct {
    Total       int64 `json:"total"`
    CurrentPage int   `json:"current_page"`
    TotalPages  int   `json:"total_pages"`
    PerPage     int   `json:"per_page"`
}

func Success(c echo.Context, data interface{}) error {
    return c.JSON(http.StatusOK, Response{
        Status: Status{
            Code:    http.StatusOK,
            Message: "Success",
        },
        Data: data,
    })
}

func Created(c echo.Context, data interface{}, message string) error {
    return c.JSON(http.StatusCreated, Response{
        Status: Status{
            Code:    http.StatusCreated,
            Message: message,
        },
        Data: data,
    })
}

func Paginated(c echo.Context, data interface{}, total int64, page, perPage int) error {
    totalPages := int(total) / perPage
    if int(total)%perPage > 0 {
        totalPages++
    }

    return c.JSON(http.StatusOK, Response{
        Status: Status{
            Code:    http.StatusOK,
            Message: "Success",
        },
        Data: data,
        Meta: &Meta{
            Total:       total,
            CurrentPage: page,
            TotalPages:  totalPages,
            PerPage:     perPage,
        },
    })
}
```

### internal/validator/validator.go

```go
package validator

import (
    "regexp"

    "github.com/go-playground/validator/v10"
)

type CustomValidator struct {
    validator *validator.Validate
}

func New() *CustomValidator {
    v := validator.New()

    // カスタムバリデーション登録
    v.RegisterValidation("hexcolor", validateHexColor)

    return &CustomValidator{validator: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
    return cv.validator.Struct(i)
}

var hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

func validateHexColor(fl validator.FieldLevel) bool {
    return hexColorRegex.MatchString(fl.Field().String())
}
```

---

## テスト

### handler_test.go

```go
package handler_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"

    "todo-api/internal/handler"
)

func TestTodoHandler_Create(t *testing.T) {
    e := echo.New()

    // リクエスト作成
    body := `{"title":"Test Todo","priority":1}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(body))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // TODO: モックDB設定

    // テスト実行
    // h := handler.NewTodoHandler(mockDB)
    // err := h.Create(c)

    // アサーション
    assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestTodoHandler_Create_ValidationError(t *testing.T) {
    e := echo.New()

    // 空のタイトル
    body := `{"title":"","priority":1}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", strings.NewReader(body))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // テスト実行 & アサーション
    assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
```

---

## マイグレーション

GORMの自動マイグレーション（開発用）:

```go
db.AutoMigrate(
    &model.User{},
    &model.Todo{},
    &model.Category{},
    &model.Tag{},
    &model.TodoTag{},
    &model.Comment{},
    &model.TodoHistory{},
    &model.Note{},
    &model.NoteRevision{},
    &model.JwtDenylist{},
)
```

本番環境ではgolang-migrateを推奨:

```bash
# インストール
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# マイグレーション作成
migrate create -ext sql -dir db/migrations -seq create_users_table

# マイグレーション実行
migrate -path db/migrations -database "$DATABASE_URL" up
```

---

## Makefile

```makefile
.PHONY: build run test lint

build:
	go build -o bin/api cmd/api/main.go

run:
	go run cmd/api/main.go

dev:
	air

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1
```

---

## 本番環境の考慮事項

1. **環境変数**: `.env`は開発のみ、本番はシステム環境変数
2. **ログレベル**: 本番では構造化ログ（JSON形式）を推奨
3. **コネクションプール**: GORMの`SetMaxOpenConns`, `SetMaxIdleConns`を設定
4. **グレースフルシャットダウン**: シグナルハンドリングを実装
5. **ヘルスチェック**: `/health`エンドポイントを追加

### グレースフルシャットダウン例

```go
func main() {
    // ... 初期化コード

    // サーバー起動（別goroutine）
    go func() {
        if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
            e.Logger.Fatal("shutting down the server")
        }
    }()

    // シグナル待機
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // グレースフルシャットダウン
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := e.Shutdown(ctx); err != nil {
        e.Logger.Fatal(err)
    }
}
```
