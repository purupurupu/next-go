# Go → Hono 移行ガイド

Go Echo バックエンドを Hono (TypeScript) に移行するためのガイドです。

## 技術スタック比較

| 領域 | Go (現行) | Hono (移行先) |
|------|----------|---------------|
| Runtime | Go 1.25 | Bun / Node.js / CF Workers |
| Framework | Echo v4 | Hono |
| ORM | GORM | Drizzle / Prisma |
| Validation | go-playground/validator | Zod |
| JWT | golang-jwt/jwt v5 | hono/jwt |
| Logging | zerolog | pino / console |

---

## プロジェクト構成

### Go (現行)

```
backend/
├── cmd/api/main.go
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── model/
│   └── middleware/
└── pkg/
```

### Hono (推奨)

```
backend-hono/
├── src/
│   ├── index.ts           # Entry point
│   ├── routes/
│   │   ├── auth.ts
│   │   ├── todos.ts
│   │   ├── notes.ts
│   │   └── index.ts
│   ├── services/
│   ├── repositories/
│   ├── models/            # Drizzle schema
│   ├── middleware/
│   │   └── auth.ts
│   ├── validators/        # Zod schemas
│   └── lib/
│       ├── db.ts
│       └── errors.ts
├── drizzle/
│   └── migrations/
├── package.json
└── tsconfig.json
```

---

## ルーティング比較

### Go Echo

```go
// cmd/api/main.go
e := echo.New()

auth := e.Group("/auth")
auth.POST("/sign_up", authHandler.SignUp)
auth.POST("/sign_in", authHandler.SignIn)

api := e.Group("/api/v1", authMiddleware.JWTAuth(cfg))
api.GET("/todos", todoHandler.List)
api.POST("/todos", todoHandler.Create)
api.GET("/todos/:id", todoHandler.Show)
```

### Hono

```typescript
// src/index.ts
import { Hono } from 'hono'
import { authRoutes } from './routes/auth'
import { todoRoutes } from './routes/todos'
import { jwtMiddleware } from './middleware/auth'

const app = new Hono()

// Auth routes (public)
app.route('/auth', authRoutes)

// API routes (protected)
const api = new Hono()
api.use('*', jwtMiddleware)
api.route('/todos', todoRoutes)

app.route('/api/v1', api)

export default app
```

```typescript
// src/routes/todos.ts
import { Hono } from 'hono'
import { zValidator } from '@hono/zod-validator'
import { createTodoSchema, updateTodoSchema } from '../validators/todo'

export const todoRoutes = new Hono()

todoRoutes.get('/', async (c) => {
  const todos = await todoService.list(c.get('userId'))
  return c.json({ data: todos, meta: { ... } })
})

todoRoutes.post('/', zValidator('json', createTodoSchema), async (c) => {
  const data = c.req.valid('json')
  const todo = await todoService.create(c.get('userId'), data)
  return c.json(todo, 201)
})

todoRoutes.get('/:id', async (c) => {
  const id = parseInt(c.req.param('id'))
  const todo = await todoService.findById(id, c.get('userId'))
  if (!todo) {
    return c.json({ error: { code: 'NOT_FOUND' } }, 404)
  }
  return c.json(todo)
})
```

---

## ハンドラ移行パターン

### Go Handler

```go
func (h *TodoHandler) Create(c echo.Context) error {
    currentUser, err := GetCurrentUserOrFail(c)
    if err != nil {
        return err
    }

    var req CreateTodoRequest
    if err := BindAndValidate(c, &req); err != nil {
        return err
    }

    todo, err := h.todoService.Create(service.CreateTodoInput{
        UserID: currentUser.ID,
        Title:  req.Title,
    })
    if err != nil {
        return err
    }

    return response.Created(c, todoToResponse(todo))
}
```

### Hono Handler

```typescript
// src/routes/todos.ts
import { zValidator } from '@hono/zod-validator'
import { z } from 'zod'

const createTodoSchema = z.object({
  title: z.string().min(1).max(255),
  description: z.string().optional(),
  category_id: z.number().optional(),
  priority: z.number().min(0).max(2).default(1),
  status: z.number().min(0).max(2).default(0),
})

todoRoutes.post('/', zValidator('json', createTodoSchema), async (c) => {
  const userId = c.get('userId')
  const input = c.req.valid('json')

  const todo = await todoService.create({
    userId,
    title: input.title,
    description: input.description,
    categoryId: input.category_id,
    priority: input.priority,
    status: input.status,
  })

  return c.json(todo, 201)
})
```

---

## ミドルウェア移行

### Go JWT Middleware

```go
func JWTAuth(cfg *config.Config, userRepo *repository.UserRepository) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            token := c.Request().Header.Get("Authorization")
            token = strings.TrimPrefix(token, "Bearer ")

            claims, err := validateToken(token, cfg.JWTSecret)
            if err != nil {
                return errors.Unauthorized("Invalid token")
            }

            user, err := userRepo.FindByID(claims.UserID)
            if err != nil {
                return errors.Unauthorized("User not found")
            }

            c.Set("currentUser", user)
            return next(c)
        }
    }
}
```

### Hono JWT Middleware

```typescript
// src/middleware/auth.ts
import { jwt } from 'hono/jwt'
import type { Context, Next } from 'hono'

export const jwtMiddleware = jwt({
  secret: process.env.JWT_SECRET!,
})

export const authMiddleware = async (c: Context, next: Next) => {
  const payload = c.get('jwtPayload')

  if (!payload?.sub) {
    return c.json({ error: { code: 'UNAUTHORIZED' } }, 401)
  }

  const user = await userRepository.findById(parseInt(payload.sub))
  if (!user) {
    return c.json({ error: { code: 'UNAUTHORIZED' } }, 401)
  }

  c.set('userId', user.id)
  c.set('currentUser', user)

  await next()
}
```

---

## バリデーション移行

### Go Validator

```go
type CreateTodoRequest struct {
    Title       string  `json:"title" validate:"required,max=255"`
    Description *string `json:"description" validate:"omitempty,max=5000"`
    Priority    *int    `json:"priority" validate:"omitempty,min=0,max=2"`
}
```

### Zod Schema

```typescript
// src/validators/todo.ts
import { z } from 'zod'

export const createTodoSchema = z.object({
  title: z.string().min(1).max(255),
  description: z.string().max(5000).optional(),
  priority: z.number().min(0).max(2).optional(),
  status: z.number().min(0).max(2).optional(),
  category_id: z.number().positive().optional(),
  due_date: z.string().datetime().optional(),
})

export type CreateTodoInput = z.infer<typeof createTodoSchema>
```

---

## データベースアクセス

### GORM (Go)

```go
func (r *TodoRepository) FindByID(id, userID int64) (*model.Todo, error) {
    var todo model.Todo
    err := r.db.
        Preload("Category").
        Preload("Tags").
        Where("id = ? AND user_id = ?", id, userID).
        First(&todo).Error
    if err != nil {
        return nil, err
    }
    return &todo, nil
}
```

### Drizzle (Hono)

```typescript
// src/repositories/todo.ts
import { db } from '../lib/db'
import { todos, categories, tags, todoTags } from '../models/schema'
import { eq, and } from 'drizzle-orm'

export const todoRepository = {
  async findById(id: number, userId: number) {
    const result = await db.query.todos.findFirst({
      where: and(eq(todos.id, id), eq(todos.userId, userId)),
      with: {
        category: true,
        tags: {
          with: { tag: true }
        }
      }
    })
    return result
  },

  async create(data: NewTodo) {
    const [todo] = await db.insert(todos).values(data).returning()
    return todo
  }
}
```

---

## エラーハンドリング

### Go

```go
// internal/errors/errors.go
func NotFound(resource string, id int64) error {
    return echo.NewHTTPError(http.StatusNotFound, map[string]interface{}{
        "error": map[string]interface{}{
            "code":    "RESOURCE_NOT_FOUND",
            "message": fmt.Sprintf("%s with ID '%d' not found", resource, id),
        },
    })
}
```

### Hono

```typescript
// src/lib/errors.ts
import { HTTPException } from 'hono/http-exception'

export class NotFoundError extends HTTPException {
  constructor(resource: string, id: number) {
    super(404, {
      message: JSON.stringify({
        error: {
          code: 'RESOURCE_NOT_FOUND',
          message: `${resource} with ID '${id}' not found`,
        }
      })
    })
  }
}

// Error handler
app.onError((err, c) => {
  if (err instanceof HTTPException) {
    return c.json(JSON.parse(err.message), err.status)
  }
  console.error(err)
  return c.json({ error: { code: 'INTERNAL_ERROR' } }, 500)
})
```

---

## 移行チェックリスト

### Phase 1: 基盤

- [ ] プロジェクトセットアップ (Bun/pnpm)
- [ ] Drizzle スキーマ定義
- [ ] データベースマイグレーション
- [ ] JWT 認証ミドルウェア
- [ ] エラーハンドリング共通化

### Phase 2: 認証

- [ ] POST /auth/sign_up
- [ ] POST /auth/sign_in
- [ ] DELETE /auth/sign_out

### Phase 3: Todo CRUD

- [ ] GET /api/v1/todos
- [ ] POST /api/v1/todos
- [ ] GET /api/v1/todos/:id
- [ ] PATCH /api/v1/todos/:id
- [ ] DELETE /api/v1/todos/:id
- [ ] GET /api/v1/todos/search
- [ ] PATCH /api/v1/todos/update_order

### Phase 4: 関連リソース

- [ ] Categories CRUD
- [ ] Tags CRUD
- [ ] Comments CRUD
- [ ] Todo History

### Phase 5: ファイル・ノート

- [ ] File upload (S3)
- [ ] Notes CRUD
- [ ] Note revisions

---

## 依存関係 (package.json)

```json
{
  "dependencies": {
    "hono": "^4.0.0",
    "@hono/zod-validator": "^0.2.0",
    "drizzle-orm": "^0.29.0",
    "postgres": "^3.4.0",
    "zod": "^3.22.0",
    "bcrypt": "^5.1.0",
    "jose": "^5.2.0"
  },
  "devDependencies": {
    "drizzle-kit": "^0.20.0",
    "typescript": "^5.3.0",
    "@types/bcrypt": "^5.0.0"
  }
}
```

---

## 参考リンク

- [Hono Documentation](https://hono.dev/)
- [Drizzle ORM](https://orm.drizzle.team/)
- [Zod Documentation](https://zod.dev/)
- [Hono JWT Middleware](https://hono.dev/middleware/builtin/jwt)
