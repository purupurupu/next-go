# Docker設定仕様書

## 現在のDocker構成（Rails）

### サービス構成

```
┌─────────────────────────────────────────────────────────────┐
│                      Docker Compose                          │
├─────────────┬─────────────┬─────────────┬─────────────────────┤
│  frontend   │   backend   │     db      │       redis        │
│  (Next.js)  │   (Rails)   │ (PostgreSQL)│      (Redis)       │
│   :3000     │    :3001    │   :5432     │       :6379        │
└─────────────┴─────────────┴─────────────┴─────────────────────┘
```

### 現在のcompose.yml

```yaml
services:
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    depends_on:
      - backend

  backend:
    build: ./backend
    ports:
      - "3001:3000"
    volumes:
      - ./backend:/app
    depends_on:
      - db
      - redis
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/todo_next
      - RAILS_ENV=development
      - RAILS_MASTER_KEY=${RAILS_MASTER_KEY}
      - SECRET_KEY_BASE=${SECRET_KEY_BASE}
      - REDIS_URL=redis://redis:6379/0
    command: bash -c "rm -f tmp/pids/server.pid && bundle exec rails s -p 3000 -b '0.0.0.0'"

  db:
    image: postgres:15
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### 現在のbackend/Dockerfile

```dockerfile
FROM ruby:3.4.5

RUN apt-get update -qq && apt-get install -y nodejs postgresql-client

WORKDIR /app

COPY Gemfile Gemfile.lock ./
RUN gem install bundler
RUN bundle install

COPY . .

ENV RAILS_ENV=development
ENV PATH="/app/bin:${PATH}"

ARG RAILS_MASTER_KEY
ENV RAILS_MASTER_KEY=${RAILS_MASTER_KEY}

ENTRYPOINT ["entrypoint.sh"]
EXPOSE 3000
CMD ["rails", "server", "-b", "0.0.0.0"]
```

---

## 環境変数

### 必須環境変数

| 変数名 | 説明 | 例 |
|--------|------|-----|
| DATABASE_URL | PostgreSQL接続URL | postgres://user:pass@db:5432/todo_next |
| JWT_SECRET | JWT署名用シークレット | ランダム文字列（256bit以上） |
| POSTGRES_DB | データベース名 | todo_next |
| POSTGRES_USER | DBユーザー名 | postgres |
| POSTGRES_PASSWORD | DBパスワード | password |

### .envファイル例

```env
# Database
POSTGRES_DB=todo_next
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
DATABASE_URL=postgres://postgres:password@db:5432/todo_next

# JWT
JWT_SECRET=your-secret-key-at-least-256-bits-long

# Redis (optional)
REDIS_URL=redis://redis:6379/0
```

---

## Rust向けDocker設定

### backend/Dockerfile (Rust)

```dockerfile
# ===== ビルドステージ =====
FROM rust:1.75-bookworm as builder

WORKDIR /app

# 依存関係のキャッシュ用
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release && rm -rf src

# ソースコードをコピーしてビルド
COPY . .
RUN touch src/main.rs && cargo build --release

# ===== 実行ステージ =====
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libpq5 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# ビルド済みバイナリをコピー
COPY --from=builder /app/target/release/todo-api /app/todo-api

ENV RUST_LOG=info

EXPOSE 3000

CMD ["./todo-api"]
```

### 開発用Dockerfile (Rust)

```dockerfile
FROM rust:1.75-bookworm

WORKDIR /app

# cargo-watch for hot reload
RUN cargo install cargo-watch

# PostgreSQLクライアント
RUN apt-get update && apt-get install -y \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

COPY . .

EXPOSE 3000

# ホットリロード付きで起動
CMD ["cargo", "watch", "-x", "run"]
```

### compose.yml (Rust)

```yaml
services:
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    depends_on:
      - backend

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.dev  # 開発用
    ports:
      - "3001:3000"
    volumes:
      - ./backend:/app
      - cargo_cache:/usr/local/cargo/registry
      - target_cache:/app/target
    depends_on:
      - db
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/todo_next
      - JWT_SECRET=${JWT_SECRET}
      - RUST_LOG=debug
      - RUST_BACKTRACE=1

  db:
    image: postgres:15
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

volumes:
  postgres_data:
  cargo_cache:
  target_cache:
```

---

## Go向けDocker設定

### backend/Dockerfile (Go)

```dockerfile
# ===== ビルドステージ =====
FROM golang:1.22-bookworm as builder

WORKDIR /app

# 依存関係のダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピーしてビルド
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-api ./cmd/api

# ===== 実行ステージ =====
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /todo-api /app/todo-api

EXPOSE 3000

CMD ["/app/todo-api"]
```

### 開発用Dockerfile (Go)

```dockerfile
FROM golang:1.22-bookworm

WORKDIR /app

# Air for hot reload
RUN go install github.com/cosmtrek/air@latest

# PostgreSQLクライアント
RUN apt-get update && apt-get install -y \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 3000

# ホットリロード付きで起動
CMD ["air"]
```

### .air.toml (Go ホットリロード設定)

```toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/api"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "node_modules"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```

### compose.yml (Go)

```yaml
services:
  frontend:
    build: ./frontend
    ports:
      - "3000:3000"
    volumes:
      - ./frontend:/app
      - /app/node_modules
    depends_on:
      - backend

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.dev  # 開発用
    ports:
      - "3001:3000"
    volumes:
      - ./backend:/app
      - go_cache:/go/pkg/mod
    depends_on:
      - db
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/todo_next
      - JWT_SECRET=${JWT_SECRET}
      - PORT=3000

  db:
    image: postgres:15
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

volumes:
  postgres_data:
  go_cache:
```

---

## 本番環境向けDocker設定

### マルチステージビルドの利点

| 項目 | 開発用 | 本番用 |
|------|--------|--------|
| イメージサイズ | 大（1GB+） | 小（10-50MB） |
| ビルドツール | 含む | 含まない |
| デバッグ | 可能 | 最小限 |
| セキュリティ | 標準 | 攻撃面が小さい |

### 本番用compose.yml

```yaml
services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile  # 本番用マルチステージ
    ports:
      - "3001:3000"
    depends_on:
      - db
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - JWT_SECRET=${JWT_SECRET}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  db:
    image: postgres:15
    volumes:
      - postgres_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    restart: unless-stopped

volumes:
  postgres_data:
```

---

## マイグレーション実行

### Rust (SeaORM CLI)

```yaml
# compose.ymlに追加
  migration:
    build:
      context: ./backend
      dockerfile: Dockerfile.dev
    command: sea-orm-cli migrate up
    depends_on:
      - db
    environment:
      - DATABASE_URL=postgres://postgres:password@db:5432/todo_next
```

### Go (golang-migrate)

```yaml
# compose.ymlに追加
  migration:
    image: migrate/migrate
    command: ["-path", "/migrations", "-database", "postgres://postgres:password@db:5432/todo_next?sslmode=disable", "up"]
    volumes:
      - ./backend/db/migrations:/migrations
    depends_on:
      - db
```

---

## よく使うコマンド

### 開発

```bash
# サービス起動
docker compose up -d

# ログ確認
docker compose logs -f backend

# コンテナに入る
docker compose exec backend sh

# 再ビルド
docker compose build --no-cache backend

# 停止
docker compose down
```

### データベース

```bash
# PostgreSQLに接続
docker compose exec db psql -U postgres -d todo_next

# マイグレーション実行
docker compose run --rm migration

# データベースリセット
docker compose down -v  # ボリューム削除
docker compose up -d
```

### 本番デプロイ

```bash
# 本番イメージビルド
docker build -t todo-api:latest ./backend

# イメージサイズ確認
docker images todo-api

# 本番起動
docker compose -f compose.prod.yml up -d
```

---

## トラブルシューティング

### よくある問題

| 問題 | 原因 | 解決策 |
|------|------|--------|
| DB接続エラー | dbサービス未起動 | `depends_on`確認、`docker compose up db`先に実行 |
| ポート競合 | 既存プロセス | `lsof -i :3001`で確認、プロセス終了 |
| ビルドキャッシュ問題 | 古いキャッシュ | `docker compose build --no-cache` |
| ボリュームマウントエラー | パーミッション | Dockerの共有設定確認 |

### ヘルスチェック追加

```yaml
backend:
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 40s
```

### ヘルスチェックエンドポイント実装

**Rust (Axum)**
```rust
async fn health() -> &'static str {
    "OK"
}

// ルーターに追加
app.route("/health", get(health))
```

**Go (Echo)**
```go
e.GET("/health", func(c echo.Context) error {
    return c.String(http.StatusOK, "OK")
})
```
