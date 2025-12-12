# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

日本語で回答してください。

## Architecture Overview

This is a full-stack Todo application with user authentication using:

- **Frontend**: Next.js 15.4.1 with TypeScript, React 19, and Tailwind CSS v4
- **Package Manager**: pnpm (NOT npm)
- **Backend**: Rails 7.1.3+ API-only application with Devise + JWT authentication
- **Database**: PostgreSQL 15
- **Infrastructure**: Docker Compose orchestrating three services

Services run on:

- Frontend: <http://localhost:3000>
- Backend API: <http://localhost:3001>
- PostgreSQL: localhost:5432

## 📚 Documentation

For detailed technical documentation, see the [docs directory](./docs/):

- [Architecture Details](./docs/architecture/) - System design and technical decisions
- [API Documentation](./docs/api/) - Complete API reference
- [Development Guides](./docs/guides/) - Setup, coding standards, and workflows
- [Feature Specifications](./docs/features/) - Detailed feature documentation

## Common Development Commands

### Docker Operations

```bash
# Start all services
docker compose up -d

# Database operations (run after services are up)
docker compose exec backend bundle exec rails db:create
docker compose exec backend bundle exec rails db:migrate
docker compose exec backend bundle exec rails db:seed

# Database seed options
docker compose exec backend bundle exec rails db:seed                    # Normal seed (preserves existing data)
docker compose exec backend bash -c "RESET_DB=true bundle exec rails db:seed"  # Reset and seed (clears all data first)

# Access Rails console
docker compose exec backend rails console

# View logs
docker compose logs -f backend
docker compose logs -f frontend

# IMPORTANT: Rebuild after package updates
# When you add new dependencies to package.json or Gemfile, you MUST rebuild the Docker image:
docker compose build frontend    # After updating frontend/package.json
docker compose build backend     # After updating backend/Gemfile
# Or rebuild without cache if you encounter dependency issues:
docker compose build --no-cache frontend
```

### Frontend Development

```bash
# All commands run inside the frontend container
docker compose exec frontend pnpm run dev        # Development server
docker compose exec frontend pnpm run build      # Production build
docker compose exec frontend pnpm run start      # Production server
docker compose exec frontend pnpm run lint       # ESLint
docker compose exec frontend pnpm run lint:fix   # ESLint with auto-fix
docker compose exec frontend pnpm run typecheck  # TypeScript type checking
docker compose exec frontend pnpm run typecheck:full  # Full TypeScript check
```

### Frontend Technical Details

**Stack**: Next.js 15.4.1, React 19.1.0, TypeScript 5, Tailwind CSS v4

**Key Dependencies**: Radix UI, date-fns, lucide-react, sonner

### Backend Development

```bash
# Run tests with RSpec (recommended for clean output)
docker compose exec backend env RAILS_ENV=test bundle exec rspec

# Run specific test file
docker compose exec backend env RAILS_ENV=test bundle exec rspec spec/models/todo_spec.rb

# Run tests with documentation format
docker compose exec backend env RAILS_ENV=test bundle exec rspec --format documentation

# Run tests with code coverage
docker compose exec backend env COVERAGE=true RAILS_ENV=test bundle exec rspec

# Run tests with coverage using rake task
docker compose exec backend bundle exec rake coverage

# Note: Always use RAILS_ENV=test when running tests to ensure test database is used

# Run authentication tests
docker compose exec backend env RAILS_ENV=test bundle exec rspec spec/requests/authentication_spec.rb


# Run RuboCop (code linter)
docker compose exec backend bundle exec rubocop                    # Check all files
docker compose exec backend bundle exec rubocop -a                 # Auto-correct safe offenses
docker compose exec backend bundle exec rubocop -A                 # Auto-correct all offenses (unsafe)
docker compose exec backend bundle exec rubocop path/to/file.rb    # Check specific file

# View coverage report (macOS)
open backend/coverage/index.html

# Quick coverage check
cat backend/coverage/.last_run.json

# Generate new resources
docker compose exec backend rails generate model ModelName
docker compose exec backend rails generate controller ControllerName

# Database operations
docker compose exec backend bundle exec rails db:drop
docker compose exec backend bundle exec rails db:create
docker compose exec backend bundle exec rails db:migrate
docker compose exec backend bundle exec rails db:seed
docker compose exec backend bundle exec rails db:reset  # drop + create + migrate + seed
```

### Backend Technical Details

**Stack**: Ruby 3.2.5, Rails 7.1.3+, PostgreSQL 15

**Key Gems**: Devise + JWT, Sidekiq, RSpec

**Core Models**:

- **User**: Authentication with email/password
- **Todo**: User's tasks with title, completion status, position, priority (low/medium/high), status (pending/in_progress/completed), optional description, due date, category association, tag associations, file attachments, comments, and history tracking
- **Category**: User-scoped organization categories with name and color for grouping todos
- **Tag**: User-scoped flexible tags with name and color for labeling todos (many-to-many relationship)
- **TodoTag**: Junction table for the many-to-many relationship between todos and tags
- **Comment**: Polymorphic comments on todos with soft delete and edit time limits
- **TodoHistory**: Automatic audit trail for all todo changes
- **JwtDenylist**: Revoked tokens for secure logout

See [Database Architecture](./docs/architecture/database.md) for detailed schema.

**API Endpoints**:

- Authentication: `/auth/*` (login, register, logout)
- Todos: `/api/v1/todos/*` (CRUD + bulk reorder + tag assignment + file attachments)
- Todo Search: `/api/v1/todos/search` (advanced search and filtering)
- Categories: `/api/v1/categories/*` (CRUD operations)
- Tags: `/api/v1/tags/*` (CRUD operations)
- Comments: `/api/v1/todos/:todo_id/comments/*` (CRUD with soft delete)
- History: `/api/v1/todos/:todo_id/histories` (view todo change history)

See [API Documentation](./docs/api/) for details.

## API Structure

The Rails backend provides:

- **Authentication API** at `/auth/*` with user registration, login, and logout
- **Todo API** at `/api/v1/todos` with user-scoped CRUD operations
- **Category API** at `/api/v1/categories` with user-scoped CRUD operations
- **Tag API** at `/api/v1/tags` with user-scoped CRUD operations
- **Comment API** at `/api/v1/todos/:todo_id/comments` for todo discussions
- **History API** at `/api/v1/todos/:todo_id/histories` for audit trail
- **JWT Authentication** for API access with token-based authentication
- Standard CRUD endpoints (GET, POST, PUT, DELETE)
- Bulk update endpoint: `PATCH /api/v1/todos/update_order` for drag-and-drop reordering
- Todo model attributes: `title`, `completed`, `position`, `priority`, `status`, `description`, `due_date`, `category_id`, `tag_ids`, `user_id`, `files` (attachments)
- Category model attributes: `name`, `color`, `user_id`, `todos_count` (counter cache), `created_at`, `updated_at`
- Tag model attributes: `name`, `color`, `user_id`, `created_at`, `updated_at`
- Comment model attributes: `content`, `user_id`, `commentable_type`, `commentable_id`, `deleted_at`, `created_at`, `updated_at`
- TodoHistory model attributes: `todo_id`, `user_id`, `action`, `changes` (JSONB), `created_at`
- User model attributes: `email`, `name`, `created_at`

Frontend should make API calls to:

- `http://localhost:3001/api/v1/todos` - Basic todo operations
- `http://localhost:3001/api/v1/todos/search` - Search and filtering
- `http://localhost:3001/api/v1/categories` - Category management
- `http://localhost:3001/api/v1/tags` - Tag management
- `http://localhost:3001/api/v1/todos/:id/comments` - Comment management
- `http://localhost:3001/api/v1/todos/:id/histories` - View history
- `http://localhost:3001/auth/*` - Authentication

## Key Implementation Details

1. **Authentication System**:

   - Devise + Devise-JWT for user authentication
   - JWT tokens returned in Authorization header and stored in localStorage on frontend
   - Token-based API authentication with Bearer tokens
   - JWT denylist for secure logout
   - User-scoped todos (each user sees only their own todos)

2. **CORS Configuration**: Backend is configured to accept requests from `localhost:3000` (frontend) with credentials support and Authorization header exposure

3. **Database Migrations**: Always run migrations after pulling new changes

4. **Hot Reloading**: Both frontend and backend support hot reloading in development

5. **TypeScript Path Aliases**: Use `@/*` for imports from the `src` directory in frontend

6. **Docker Compose**: Three services (frontend, backend, db) with proper dependency management

7. **Environment Variables**: Database credentials and Rails secrets managed via environment variables

8. **API Error Handling**: Consistent error responses with proper HTTP status codes

9. **State Management**: React hooks with optimistic updates for better UX

10. **Testing**: RSpec for backend testing with dedicated test service in Docker Compose

## Current State

The project has successfully transitioned from Nuxt.js to Next.js and now includes full user authentication. The backend API is fully functional with complete CRUD operations, validation, and drag-and-drop reordering support. The frontend is a Next.js application with a complete todo feature implementation including:

**Authentication Features**:

- User registration and login
- JWT token-based authentication
- Protected routes with authentication guards
- Persistent authentication state
- Secure logout with token invalidation

**Todo Features**:

- User-scoped todos (each user sees only their own)
- Todo creation with due dates
- Todo editing and deletion
- Status completion toggle
- Drag-and-drop reordering (API ready, UI pending)
- Basic filtering (all, active, completed)
- Category assignment (one-to-many)
- Tag assignment (many-to-many)
- File attachments with Active Storage
- Comments with soft delete and 15-minute edit window
- Complete audit history tracking
- Optimistic updates for better UX
- Error handling and validation
- Responsive design with Tailwind CSS
- shadcn/ui components for consistent UI

**Search and Filtering Features**:

- Full-text search in title and description
- Advanced filtering by category, status, priority, tags
- Date range filtering for due dates
- Multi-criteria sorting (7+ fields)
- Pagination with customizable page size
- Search result highlighting
- Empty state with helpful suggestions
- Real-time search with debouncing

**Additional Features**:

- **Comments**: Add discussions to todos with soft delete and edit time limits
- **File Attachments**: Upload multiple files per todo (10MB limit per file)
- **History Tracking**: Complete audit trail of all todo changes
- **Counter Cache**: Efficient todo counting for categories
- **Performance**: Optimized queries with proper indexing

## Development Guidelines

1. **Package Manager**: Always use pnpm, NOT npm
2. **API Calls**: Use the provided API clients, not direct fetch
3. **Authentication**: Check auth state before protected features
4. **Commits**: Small, frequent commits with clear messages
5. **Code Style**: Follow existing patterns in the codebase
6. **Docker Dependencies**: After adding new packages to package.json or Gemfile, always rebuild the Docker image

### Git Commit Best Practices

**Commit Size and Scope**:

- Keep commits small and focused on a single logical change
- Each commit should represent a complete, working state
- Avoid mixing unrelated changes in a single commit

**Recommended Commit Granularity**:

1. **Model/Migration Changes**: Create model with migration and validations
2. **API Endpoints**: Controller actions with routing changes
3. **Frontend Components**: Component with its types and styles
4. **Integration Changes**: Updates that connect different parts
5. **Test Additions**: Tests for specific features
6. **Documentation Updates**: Separate from code changes

**Example Commit Sequence for a Feature**:

```
1. feat(backend): Add Category model with validations
2. feat(backend): Add category association to Todo model
3. feat(backend): Add Category API controller
4. feat(backend): Add serializers for Category and Todo
5. feat(frontend): Add Category types and interfaces
6. feat(frontend): Add Category API client
7. feat(frontend): Add Category management components
8. feat(frontend): Update Todo components to support categories
9. test(backend): Add Category model and controller tests
10. docs: Update API documentation for categories
```

**Commit Message Format**:

- Use conventional commits: `type(scope): description`
- Types: feat, fix, docs, style, refactor, test, chore
- Keep the first line under 50 characters
- Add detailed description if needed

**Before Creating Pull Requests**:

- Run frontend checks: `pnpm run lint`, `pnpm run typecheck`
- Run backend tests: `docker compose exec backend env RAILS_ENV=test bundle exec rspec`
- Run backend linter: `docker compose exec backend bundle exec rubocop`
- Update documentation if APIs or architecture changed
- Ensure all checks pass before pushing

See [Development Guide](./docs/guides/development.md) for detailed guidelines.

## Troubleshooting

### Node modules issues in Docker

If you encounter errors like "Module not found" after adding new packages:

1. **Symptoms:**

   - `Module not found: Can't resolve '@some-package'`
   - pnpm store location mismatch errors
   - Turbopack panic errors

2. **Solution:**

   ```bash
   # Stop and rebuild the container
   docker compose down
   docker compose build --no-cache frontend
   docker compose up -d
   ```

3. **Prevention:**
   - Always rebuild Docker images after updating package.json
   - Use `docker compose build frontend` after adding new npm packages
   - Use `docker compose build backend` after adding new gems

## Frontend Architecture

### Directory Structure

```
frontend/src/
├── app/               # Next.js App Router pages (routing files)
│   ├── auth/         # Authentication pages
│   ├── layout.tsx    # Root layout with auth provider
│   └── page.tsx      # Main todo page
├── components/        # 横断的（ドメインに依存しない）なUIコンポーネント
│   ├── auth/         # Authentication components
│   ├── layouts/      # Layout components
│   ├── ui/           # shadcn/ui components
│   ├── navigation.tsx        # Navigation component
│   └── protected-route.tsx   # Authentication guard
├── contexts/          # React contexts
│   └── auth-context.tsx     # Authentication context
├── features/          # 特定のドメイン・機能に関係するコンポーネント
│   └── todo/         # Todo feature
│       ├── components/   # Todo-specific components
│       ├── hooks/        # Todo-specific hooks
│       ├── lib/          # Todo-specific API client
│       └── types/        # Todo-specific types
├── hooks/             # ドメインに依存しない、横断的なhooks
├── lib/               # ライブラリの処理や標準処理を共通化したコード
│   ├── api-client.ts     # Base HTTP client
│   ├── auth-client.ts    # Authentication API client
│   ├── constants.ts      # API endpoints and constants
│   └── utils.ts          # Utility functions
├── styles/            # スタイリング（css）に関するファイル
└── types/             # 横断的な型定義
```

### Naming Conventions

1. **Files and Directories**:

   - Components: PascalCase (e.g., `TodoItem.tsx`, `TodoForm.tsx`)
   - Hooks: camelCase with `use` prefix (e.g., `useTodos.ts`)
   - Utilities: kebab-case (e.g., `api-client.ts`, `constants.ts`)
   - Types: kebab-case (e.g., `todo.ts`)

2. **Code Naming**:
   - React components: PascalCase
   - Functions/variables: camelCase
   - Constants: UPPER_SNAKE_CASE
   - Interfaces: PascalCase with descriptive suffix (e.g., `TodoItemProps`, `CreateTodoData`)

### Key Files

**Authentication**:

- `lib/auth-client.ts` - Authentication API client with login/register/logout
- `contexts/auth-context.tsx` - Authentication context provider
- `components/auth/login-form.tsx` - Login form component
- `components/auth/register-form.tsx` - Registration form component
- `components/protected-route.tsx` - Route protection component
- `app/auth/page.tsx` - Authentication page

**API & Core**:

- `lib/api-client.ts` - Base HttpClient with common HTTP methods (GET, POST, PUT, PATCH, DELETE)
- `lib/constants.ts` - API endpoints and configuration constants
- `lib/utils.ts` - Utility functions (date formatting, validation, etc.)

**Todo Feature**:

- `features/todo/lib/api-client.ts` - Todo-specific API client extending HttpClient
- `features/todo/types/todo.ts` - Todo-related TypeScript interfaces and types
- `features/todo/hooks/useTodos.ts` - Todo state management with optimistic updates
- `features/todo/components/TodoList.tsx` - Main todo list component with drag-and-drop
- `features/todo/components/TodoItem.tsx` - Individual todo item component
- `features/todo/components/TodoForm.tsx` - Todo creation and editing form
- `features/todo/components/TodoFilters.tsx` - Filter controls (all, active, completed)

**Category Feature**:

- `features/category/lib/api-client.ts` - Category-specific API client extending HttpClient
- `features/category/types/category.ts` - Category-related TypeScript interfaces and types
- `features/category/hooks/useCategories.ts` - Category state management
- `features/category/components/CategoryManager.tsx` - Main category management component
- `features/category/components/CategoryForm.tsx` - Category creation and editing form
- `features/category/components/CategorySelector.tsx` - Category selection dropdown

**UI Components**:

- `components/ui/` - Shared UI components (shadcn/ui based)
- `components/navigation.tsx` - Navigation with authentication state

### Architecture Principles

1. **Feature-based organization**: Domain-specific code lives in `features/[domain]/`

2. **Cross-cutting concerns**: Shared utilities, types, and components live in root-level directories

3. **Separation of concerns**: Each feature has its own components, hooks, types, and utilities

4. **Reusability**: Common UI components and hooks are shared across features

5. **Authentication Architecture**:

   - JWT token-based authentication
   - Auth context for global authentication state
   - Protected routes with authentication guards
   - Persistent authentication with localStorage
   - Separate auth client for authentication API calls

6. **API Client Pattern**:

   - Base `HttpClient` provides common HTTP methods (GET, POST, PUT, PATCH, DELETE)
   - Separate `AuthClient` for authentication operations
   - Feature-specific API clients extend `HttpClient` and implement domain-specific methods
   - Hooks use feature API clients for data fetching and state management
   - Automatic JWT token injection for authenticated requests

7. **Error Handling**: Consistent error handling with `ApiError` class and proper user feedback

8. **Optimistic Updates**: UI updates immediately with rollback on API failure

9. **Type Safety**: Full TypeScript coverage with proper interfaces and type definitions

10. **Component Composition**: Small, focused components that compose well together

11. **State Management**: Local state with React hooks, avoiding external state management libraries

## Docker Environment Variables

The application uses environment variables for configuration:

**Backend Services**:

- `DATABASE_URL` - Full database connection string
- `RAILS_ENV` - Rails environment (development/production/test)
- `RAILS_MASTER_KEY` - Rails master key for encrypted credentials
- `SECRET_KEY_BASE` - Secret key for Rails session encryption

**Database (compose.yml)**:

- `POSTGRES_DB` - Database name
- `POSTGRES_USER` - Database user
- `POSTGRES_PASSWORD` - Database password

**Testing**:

- Test database: `todo_app_test`
- Development database: `todo_next`
- Test database is automatically selected when `RAILS_ENV=test` is specified
- The test_database.rb initializer ensures the correct database is used in test environment

**Note**: Create a `.env` file in the root directory with these variables for Docker Compose.

