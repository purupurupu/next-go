# Rust実装ガイド

## 概要

このドキュメントでは、RailsバックエンドをRustに移行する際の推奨技術スタックと実装パターンを説明します。

---

## 推奨技術スタック

| 領域 | 推奨ライブラリ | 代替案 |
|------|--------------|--------|
| Webフレームワーク | Axum | Actix-web |
| ORM | SeaORM | Diesel, SQLx |
| 非同期ランタイム | Tokio | async-std |
| シリアライゼーション | Serde | - |
| バリデーション | validator | garde |
| JWT | jsonwebtoken | - |
| パスワードハッシュ | bcrypt / argon2 | - |
| 環境変数 | dotenvy | config |
| ログ | tracing | log + env_logger |
| テスト | tokio::test | - |
| DB接続プール | sqlx / deadpool | bb8 |

---

## Cargo.toml

```toml
[package]
name = "todo-api"
version = "0.1.0"
edition = "2021"

[dependencies]
# Web framework
axum = { version = "0.7", features = ["macros"] }
tokio = { version = "1", features = ["full"] }
tower = "0.4"
tower-http = { version = "0.5", features = ["cors", "trace"] }

# Database
sea-orm = { version = "0.12", features = [
    "sqlx-postgres",
    "runtime-tokio-rustls",
    "macros"
] }
sqlx = { version = "0.7", features = ["runtime-tokio-rustls", "postgres"] }

# Serialization
serde = { version = "1", features = ["derive"] }
serde_json = "1"

# Validation
validator = { version = "0.16", features = ["derive"] }

# Authentication
jsonwebtoken = "9"
bcrypt = "0.15"

# Utilities
uuid = { version = "1", features = ["v4", "serde"] }
chrono = { version = "0.4", features = ["serde"] }
dotenvy = "0.15"
thiserror = "1"
anyhow = "1"

# Logging
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }

[dev-dependencies]
tokio-test = "0.4"
```

---

## ディレクトリ構成

```
src/
├── main.rs                 # エントリポイント
├── lib.rs                  # ライブラリルート
├── config.rs               # 設定管理
├── routes/
│   ├── mod.rs              # ルートモジュール
│   ├── auth.rs             # 認証ルート
│   ├── todos.rs            # Todo CRUD
│   ├── categories.rs       # Category CRUD
│   ├── tags.rs             # Tag CRUD
│   ├── comments.rs         # Comment CRUD
│   └── notes.rs            # Note CRUD
├── middleware/
│   ├── mod.rs
│   ├── auth.rs             # JWT認証ミドルウェア
│   ├── error_handler.rs    # エラーハンドリング
│   └── request_id.rs       # リクエストID
├── db/
│   ├── mod.rs
│   ├── connection.rs       # DB接続
│   └── entities/           # SeaORMエンティティ
│       ├── mod.rs
│       ├── user.rs
│       ├── todo.rs
│       ├── category.rs
│       ├── tag.rs
│       ├── todo_tag.rs
│       ├── comment.rs
│       ├── todo_history.rs
│       ├── note.rs
│       ├── note_revision.rs
│       └── jwt_denylist.rs
├── services/
│   ├── mod.rs
│   ├── auth_service.rs     # 認証サービス
│   └── todo_search.rs      # 検索サービス
├── validators/
│   ├── mod.rs
│   └── schemas.rs          # バリデーションスキーマ
├── errors/
│   ├── mod.rs
│   └── api_error.rs        # カスタムエラー
└── response/
    ├── mod.rs
    └── json.rs             # レスポンスヘルパー
```

---

## 実装パターン

### main.rs

```rust
use axum::{Router, Extension};
use std::sync::Arc;
use tower_http::cors::CorsLayer;
use tower_http::trace::TraceLayer;

mod config;
mod routes;
mod middleware;
mod db;
mod services;
mod errors;
mod response;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // 環境変数読み込み
    dotenvy::dotenv().ok();

    // ログ初期化
    tracing_subscriber::fmt::init();

    // 設定読み込み
    let config = config::Config::from_env()?;

    // DB接続
    let db = db::connection::establish(&config.database_url).await?;

    // ルーター構築
    let app = Router::new()
        .nest("/auth", routes::auth::router())
        .nest("/api/v1", routes::api_v1())
        .layer(middleware::auth::jwt_layer())
        .layer(CorsLayer::permissive())
        .layer(TraceLayer::new_for_http())
        .layer(Extension(Arc::new(db)))
        .layer(Extension(Arc::new(config)));

    // サーバー起動
    let listener = tokio::net::TcpListener::bind("0.0.0.0:3001").await?;
    tracing::info!("Server running on http://localhost:3001");
    axum::serve(listener, app).await?;

    Ok(())
}
```

### config.rs

```rust
use serde::Deserialize;

#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    pub database_url: String,
    pub jwt_secret: String,
    pub jwt_expiration_hours: i64,
}

impl Config {
    pub fn from_env() -> anyhow::Result<Self> {
        Ok(Self {
            database_url: std::env::var("DATABASE_URL")?,
            jwt_secret: std::env::var("JWT_SECRET")?,
            jwt_expiration_hours: std::env::var("JWT_EXPIRATION_HOURS")
                .unwrap_or_else(|_| "24".to_string())
                .parse()?,
        })
    }
}
```

### エラー定義 (errors/api_error.rs)

```rust
use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde::Serialize;
use serde_json::json;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum ApiError {
    #[error("Authentication failed")]
    AuthenticationFailed,

    #[error("Token expired")]
    TokenExpired,

    #[error("Not authorized")]
    AuthorizationFailed { resource: String, action: String },

    #[error("Resource not found")]
    NotFound { resource: String, id: i64 },

    #[error("Validation failed")]
    ValidationFailed { errors: serde_json::Value },

    #[error("Internal server error")]
    InternalError(#[from] anyhow::Error),
}

impl IntoResponse for ApiError {
    fn into_response(self) -> Response {
        let (status, code, message, details) = match &self {
            ApiError::AuthenticationFailed => (
                StatusCode::UNAUTHORIZED,
                "AUTHENTICATION_FAILED",
                self.to_string(),
                None,
            ),
            ApiError::TokenExpired => (
                StatusCode::UNAUTHORIZED,
                "TOKEN_EXPIRED",
                self.to_string(),
                None,
            ),
            ApiError::AuthorizationFailed { resource, action } => (
                StatusCode::FORBIDDEN,
                "AUTHORIZATION_FAILED",
                self.to_string(),
                Some(json!({ "resource": resource, "action": action })),
            ),
            ApiError::NotFound { resource, id } => (
                StatusCode::NOT_FOUND,
                "RESOURCE_NOT_FOUND",
                self.to_string(),
                Some(json!({ "resource": resource, "id": id })),
            ),
            ApiError::ValidationFailed { errors } => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "VALIDATION_FAILED",
                self.to_string(),
                Some(json!({ "validation_errors": errors })),
            ),
            ApiError::InternalError(_) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                "INTERNAL_ERROR",
                "An unexpected error occurred".to_string(),
                None,
            ),
        };

        let body = json!({
            "error": {
                "code": code,
                "message": message,
                "details": details,
                "request_id": uuid::Uuid::new_v4().to_string(),
                "timestamp": chrono::Utc::now().to_rfc3339()
            }
        });

        (status, Json(body)).into_response()
    }
}
```

### JWT認証ミドルウェア (middleware/auth.rs)

```rust
use axum::{
    async_trait,
    extract::FromRequestParts,
    http::{request::Parts, StatusCode},
    RequestPartsExt,
};
use axum_extra::{
    headers::{authorization::Bearer, Authorization},
    TypedHeader,
};
use jsonwebtoken::{decode, DecodingKey, Validation};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    pub sub: String,  // user_id
    pub jti: String,  // JWT ID
    pub exp: usize,   // 有効期限
    pub iat: usize,   // 発行日時
}

#[derive(Debug, Clone)]
pub struct CurrentUser {
    pub id: i64,
    pub email: String,
    pub name: String,
}

#[async_trait]
impl<S> FromRequestParts<S> for CurrentUser
where
    S: Send + Sync,
{
    type Rejection = (StatusCode, &'static str);

    async fn from_request_parts(parts: &mut Parts, _state: &S) -> Result<Self, Self::Rejection> {
        // Authorizationヘッダー取得
        let TypedHeader(Authorization(bearer)) = parts
            .extract::<TypedHeader<Authorization<Bearer>>>()
            .await
            .map_err(|_| (StatusCode::UNAUTHORIZED, "Missing authorization header"))?;

        // JWT検証
        let token_data = decode::<Claims>(
            bearer.token(),
            &DecodingKey::from_secret(std::env::var("JWT_SECRET").unwrap().as_bytes()),
            &Validation::default(),
        )
        .map_err(|_| (StatusCode::UNAUTHORIZED, "Invalid token"))?;

        // ユーザーID取得
        let user_id: i64 = token_data.claims.sub.parse()
            .map_err(|_| (StatusCode::UNAUTHORIZED, "Invalid user id in token"))?;

        // TODO: DBからユーザー取得 & jwt_denylistチェック

        Ok(CurrentUser {
            id: user_id,
            email: String::new(),  // DBから取得
            name: String::new(),   // DBから取得
        })
    }
}
```

### SeaORMエンティティ例 (db/entities/todo.rs)

```rust
use sea_orm::entity::prelude::*;
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, PartialEq, Eq, DeriveEntityModel, Serialize, Deserialize)]
#[sea_orm(table_name = "todos")]
pub struct Model {
    #[sea_orm(primary_key)]
    pub id: i64,
    pub user_id: i64,
    pub category_id: Option<i64>,
    pub title: String,
    pub description: Option<String>,
    pub completed: bool,
    pub position: Option<i32>,
    pub priority: i32,
    pub status: i32,
    pub due_date: Option<Date>,
    pub created_at: DateTimeUtc,
    pub updated_at: DateTimeUtc,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, EnumIter)]
pub enum Priority {
    Low = 0,
    Medium = 1,
    High = 2,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, EnumIter)]
pub enum Status {
    Pending = 0,
    InProgress = 1,
    Completed = 2,
}

#[derive(Copy, Clone, Debug, EnumIter, DeriveRelation)]
pub enum Relation {
    #[sea_orm(
        belongs_to = "super::user::Entity",
        from = "Column::UserId",
        to = "super::user::Column::Id"
    )]
    User,
    #[sea_orm(
        belongs_to = "super::category::Entity",
        from = "Column::CategoryId",
        to = "super::category::Column::Id"
    )]
    Category,
    #[sea_orm(has_many = "super::todo_tag::Entity")]
    TodoTags,
    #[sea_orm(has_many = "super::comment::Entity")]
    Comments,
    #[sea_orm(has_many = "super::todo_history::Entity")]
    TodoHistories,
}

impl ActiveModelBehavior for ActiveModel {}
```

### バリデーション (validators/schemas.rs)

```rust
use serde::Deserialize;
use validator::Validate;

#[derive(Debug, Deserialize, Validate)]
pub struct CreateTodoRequest {
    #[validate(length(min = 1, message = "can't be blank"))]
    pub title: String,

    pub description: Option<String>,

    #[validate(range(min = 0, max = 2))]
    pub priority: Option<i32>,

    #[validate(range(min = 0, max = 2))]
    pub status: Option<i32>,

    pub due_date: Option<String>,
    pub category_id: Option<i64>,
    pub tag_ids: Option<Vec<i64>>,
}

#[derive(Debug, Deserialize, Validate)]
pub struct CreateUserRequest {
    #[validate(email(message = "is invalid"))]
    pub email: String,

    #[validate(length(min = 6, message = "is too short (minimum is 6 characters)"))]
    pub password: String,

    pub password_confirmation: String,

    #[validate(length(min = 2, max = 50, message = "must be between 2 and 50 characters"))]
    pub name: String,
}

#[derive(Debug, Deserialize, Validate)]
pub struct CreateCategoryRequest {
    #[validate(length(min = 1, max = 50))]
    pub name: String,

    #[validate(regex(path = "HEX_COLOR_REGEX", message = "must be a valid hex color"))]
    pub color: Option<String>,
}

lazy_static::lazy_static! {
    static ref HEX_COLOR_REGEX: regex::Regex =
        regex::Regex::new(r"^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$").unwrap();
}
```

### Todoルートハンドラー (routes/todos.rs)

```rust
use axum::{
    extract::{Extension, Path, Query},
    routing::{get, post, patch, delete},
    Json, Router,
};
use sea_orm::DatabaseConnection;
use std::sync::Arc;

use crate::errors::ApiError;
use crate::middleware::auth::CurrentUser;
use crate::validators::schemas::CreateTodoRequest;

pub fn router() -> Router {
    Router::new()
        .route("/", get(list_todos).post(create_todo))
        .route("/search", get(search_todos))
        .route("/:id", get(get_todo).patch(update_todo).delete(delete_todo))
        .route("/update_order", patch(update_order))
        .route("/:id/tags", patch(update_tags))
        .route("/:id/files/:file_id", delete(delete_file))
}

async fn list_todos(
    current_user: CurrentUser,
    Extension(db): Extension<Arc<DatabaseConnection>>,
) -> Result<Json<serde_json::Value>, ApiError> {
    use sea_orm::{EntityTrait, QueryFilter, ColumnTrait, QueryOrder};
    use crate::db::entities::todo;

    let todos = todo::Entity::find()
        .filter(todo::Column::UserId.eq(current_user.id))
        .order_by_asc(todo::Column::Position)
        .all(db.as_ref())
        .await
        .map_err(|e| ApiError::InternalError(e.into()))?;

    Ok(Json(serde_json::json!({
        "status": { "code": 200, "message": "Success" },
        "data": todos
    })))
}

async fn create_todo(
    current_user: CurrentUser,
    Extension(db): Extension<Arc<DatabaseConnection>>,
    Json(payload): Json<CreateTodoRequest>,
) -> Result<Json<serde_json::Value>, ApiError> {
    use validator::Validate;

    // バリデーション
    payload.validate().map_err(|e| ApiError::ValidationFailed {
        errors: serde_json::to_value(e).unwrap()
    })?;

    // TODO: Todo作成ロジック

    Ok(Json(serde_json::json!({
        "status": { "code": 201, "message": "Todo created successfully" },
        "data": {}
    })))
}

async fn get_todo(
    current_user: CurrentUser,
    Path(id): Path<i64>,
    Extension(db): Extension<Arc<DatabaseConnection>>,
) -> Result<Json<serde_json::Value>, ApiError> {
    use sea_orm::{EntityTrait, QueryFilter, ColumnTrait};
    use crate::db::entities::todo;

    let todo = todo::Entity::find()
        .filter(todo::Column::Id.eq(id))
        .filter(todo::Column::UserId.eq(current_user.id))
        .one(db.as_ref())
        .await
        .map_err(|e| ApiError::InternalError(e.into()))?
        .ok_or(ApiError::NotFound {
            resource: "Todo".to_string(),
            id,
        })?;

    Ok(Json(serde_json::json!({
        "status": { "code": 200, "message": "Success" },
        "data": todo
    })))
}

// ... 他のハンドラー
```

### レスポンスヘルパー (response/json.rs)

```rust
use axum::Json;
use serde::Serialize;

#[derive(Serialize)]
pub struct ApiResponse<T: Serialize> {
    pub status: Status,
    pub data: T,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub meta: Option<Meta>,
}

#[derive(Serialize)]
pub struct Status {
    pub code: u16,
    pub message: String,
}

#[derive(Serialize)]
pub struct Meta {
    pub total: i64,
    pub current_page: i64,
    pub total_pages: i64,
    pub per_page: i64,
}

pub fn success<T: Serialize>(data: T) -> Json<ApiResponse<T>> {
    Json(ApiResponse {
        status: Status {
            code: 200,
            message: "Success".to_string(),
        },
        data,
        meta: None,
    })
}

pub fn created<T: Serialize>(data: T, message: &str) -> Json<ApiResponse<T>> {
    Json(ApiResponse {
        status: Status {
            code: 201,
            message: message.to_string(),
        },
        data,
        meta: None,
    })
}

pub fn paginated<T: Serialize>(
    data: T,
    total: i64,
    page: i64,
    per_page: i64,
) -> Json<ApiResponse<T>> {
    let total_pages = (total as f64 / per_page as f64).ceil() as i64;
    Json(ApiResponse {
        status: Status {
            code: 200,
            message: "Success".to_string(),
        },
        data,
        meta: Some(Meta {
            total,
            current_page: page,
            total_pages,
            per_page,
        }),
    })
}
```

---

## テスト

### ユニットテスト例

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_create_todo_validation() {
        let payload = CreateTodoRequest {
            title: "".to_string(),  // 空のタイトル
            description: None,
            priority: Some(5),      // 無効な優先度
            status: None,
            due_date: None,
            category_id: None,
            tag_ids: None,
        };

        let result = payload.validate();
        assert!(result.is_err());
    }

    #[test]
    fn test_hex_color_validation() {
        assert!(HEX_COLOR_REGEX.is_match("#FF5733"));
        assert!(HEX_COLOR_REGEX.is_match("#fff"));
        assert!(!HEX_COLOR_REGEX.is_match("FF5733"));
        assert!(!HEX_COLOR_REGEX.is_match("#FFFF"));
    }
}
```

---

## マイグレーション

SeaORM CLIを使用してマイグレーションを管理：

```bash
# SeaORM CLI インストール
cargo install sea-orm-cli

# マイグレーション作成
sea-orm-cli migrate generate create_users_table

# マイグレーション実行
sea-orm-cli migrate up

# エンティティ生成
sea-orm-cli generate entity -o src/db/entities
```

---

## 本番環境の考慮事項

1. **環境変数**: `.env`ファイルは開発のみ、本番はシステム環境変数を使用
2. **ログレベル**: 本番では`RUST_LOG=info`を推奨
3. **コネクションプール**: `max_connections`を適切に設定
4. **エラーハンドリング**: 本番では詳細なエラー情報を隠す
5. **HTTPS**: TLSは必須（リバースプロキシでも可）
