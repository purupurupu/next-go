# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

日本語で回答してください。
コミットメッセージとPRも日本語で作成してください。

## Architecture Overview

Full-stack Todo application with:

- **Frontend**: Next.js 16 with TypeScript, React 19.2, Tailwind CSS v4
- **Package Manager**: pnpm (NOT npm)
- **Backend**: Go 1.25 with Echo framework, GORM, JWT authentication
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Infrastructure**: Docker Compose

Services:

- Frontend: http://localhost:3000
- Backend API: http://localhost:3001
- PostgreSQL: localhost:5432
- Redis: localhost:6379

## Common Development Commands

### Docker Operations

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f backend
docker compose logs -f frontend

# Rebuild after dependency updates
docker compose build frontend    # After package.json changes
docker compose build backend     # After go.mod changes
docker compose build --no-cache frontend  # Force rebuild
```

### Frontend Development

```bash
docker compose exec frontend pnpm run dev        # Development server
docker compose exec frontend pnpm run build      # Production build
docker compose exec frontend pnpm run lint       # ESLint
docker compose exec frontend pnpm run lint:fix   # ESLint with auto-fix
docker compose exec frontend pnpm run typecheck  # TypeScript check
```

### Backend Development

```bash
# Start test database
docker compose up -d db_test

# Run all tests
docker compose exec backend go test -v ./...

# Run specific package tests
docker compose exec backend go test -v ./internal/handler/...

# Run single test function
docker compose exec backend go test -v ./internal/handler/... -run TestTodoCreate

# Run tests matching pattern (e.g., all Comment tests)
docker compose exec backend go test -v ./internal/handler/... -run "TestComment"

# Run with coverage
docker compose exec backend go test -cover ./...
docker compose exec backend go test -coverprofile=coverage.out ./...

# Lint (requires golangci-lint)
docker compose exec backend golangci-lint run

# Build
docker compose exec backend go build -o bin/api cmd/api/main.go
```

**Test Utilities** (`internal/testutil/`):
- `SetupTestFixture(t)` - 全テスト依存関係を初期化
- `f.CreateUser(email)` - テストユーザー作成（JWT token返却）
- `f.CreateTodo(userID, title)` - テストTodo作成
- `f.CreateComment(userID, todoID, content)` - テストコメント作成
- `f.CallAuth(token, method, path, body, handler)` - 認証付きハンドラ呼び出し

### Database

Database auto-migrates on startup in development mode via GORM AutoMigrate.

## Backend Architecture (Go)

```
backend/
├── cmd/api/main.go           # Entry point, DI, routing
├── internal/
│   ├── config/               # Environment config (envconfig)
│   ├── handler/              # HTTP handlers (Echo)
│   ├── middleware/           # JWT auth middleware
│   ├── model/                # GORM models
│   ├── repository/           # Data access layer (interfaces.go にインターフェース定義)
│   ├── service/              # Business logic (TodoService: 履歴記録含む)
│   ├── testutil/             # テストヘルパー (fixture, helpers)
│   ├── validator/            # Request validation (go-playground/validator)
│   └── errors/               # API error handling (EditTimeExpired など)
└── pkg/
    ├── database/             # DB connection
    ├── response/             # Standardized API responses
    └── util/                 # Time formatting utilities
```

**Key Dependencies**:
- Echo v4 (web framework)
- GORM (ORM)
- golang-jwt/jwt v5 (authentication)
- go-playground/validator v10 (validation)
- zerolog (structured logging)
- Air (hot reload in development)

**Current Implementation Status**:
- Authentication (sign_up, sign_in, sign_out) - Complete
- Todo CRUD (with position ordering) - Complete
- Categories CRUD (with todo_count counter cache) - Complete
- Tags CRUD - Complete
- TodoService (business logic layer) - Complete
- Todo Search/Filter (GET /api/v1/todos/search) - Complete
- Comments (15分編集制限、ソフトデリート) - Complete
- TodoHistory (自動履歴記録、日本語メッセージ) - Complete

## Frontend Architecture

```
frontend/src/
├── app/                  # Next.js App Router pages
├── components/           # Shared UI components (shadcn/ui)
├── contexts/             # React contexts (auth-context)
├── features/             # Feature-based modules
│   ├── todo/            # Todo feature (components, hooks, types, api)
│   ├── category/        # Category management
│   ├── tag/             # Tag management
│   ├── comment/         # Comments
│   ├── history/         # Audit history
│   └── notes/           # Notes feature
├── hooks/                # Shared hooks
├── lib/                  # API clients, utilities
│   ├── api-client.ts    # HttpClient + ApiClient（/api/v1プレフィックス付き）
│   ├── auth-client.ts   # Auth API client
│   └── constants.ts     # API base URL
└── types/                # Shared type definitions
```

**Key Patterns**:
- Feature-based organization: `features/[domain]/` contains domain-specific code
- API Client pattern: `ApiClient`（/api/v1プレフィックス付き）を各featureで継承
- Auth API: `httpClient`を使用（/api/v1プレフィックスなし）
- TypeScript path aliases: `@/*` maps to `src/`
- Optimistic updates with rollback on failure

## API Endpoints

**Authentication** (public):
- `POST /auth/sign_up` - Register
- `POST /auth/sign_in` - Login
- `DELETE /auth/sign_out` - Logout (requires auth)

**API v1** (requires Bearer token):
- `/api/v1/todos` - Todo CRUD
- `/api/v1/todos/search` - Todo search with filters/sort/pagination
- `/api/v1/categories` - Category CRUD
- `/api/v1/tags` - Tag CRUD
- `/api/v1/todos/:todo_id/comments` - Comments (CRUD、15分編集制限)
- `/api/v1/todos/:todo_id/histories` - Audit trail (読み取り専用、ページネーション)

## Environment Variables

**Backend** (compose.yml):
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT signing key
- `PORT` - Server port (default: 3000, mapped to 3001)
- `ENV` - Environment (development/production)
- `REDIS_URL` - Redis connection string

**Database**:
- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`

## Development Guidelines

1. **Package Manager**: Always use pnpm for frontend
2. **API Calls**: Use provided API clients, not direct fetch
3. **Docker**: Rebuild images after dependency changes
4. **Before PR**: Run `pnpm run lint`, `pnpm run typecheck`, `go test ./...`

## Git Conventions

**コミットメッセージ** (Semantic Commits、日本語):
```
feat(backend): ユーザー認証機能を追加
fix(frontend): ログインフォームのバリデーションエラーを修正
docs: APIドキュメントを更新
refactor(backend): 認証ミドルウェアをリファクタリング
test(backend): Todoハンドラのテストを追加
chore: 依存関係を更新
```

**PRタイトル・本文**: 日本語で記述

## Documentation

See `docs/` for detailed documentation:
- `docs/migration/go-implementation-guide.md` - Go backend patterns
- `docs/api/` - API reference
- `docs/architecture/` - System design

## Troubleshooting

### Module not found errors in Docker
```bash
docker compose down
docker compose build --no-cache frontend
docker compose up -d
```

### Go dependency issues
```bash
docker compose exec backend go mod tidy
docker compose build backend
```