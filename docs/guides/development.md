# Development Guide

## Development Workflow

### 1. Environment Setup
See [Getting Started Guide](./getting-started.md) for initial setup.

### 2. Development Process
1. Start Docker containers: `docker compose up -d`
2. Frontend runs on http://localhost:3000
3. Backend API runs on http://localhost:3001
4. Both support hot reloading

### 3. Making Changes
1. Create feature branch from main
2. Make changes with frequent commits
3. Run tests before pushing
4. Create pull request with clear description

## Coding Standards

### General Principles
1. **Consistency**: Follow existing patterns in the codebase
2. **Clarity**: Write self-documenting code
3. **Simplicity**: Prefer simple solutions over clever ones
4. **Testing**: Write tests for new features

### Frontend Standards

#### TypeScript
```typescript
// Use explicit types for function parameters and return values
function calculateTotal(items: TodoItem[]): number {
  return items.length;
}

// Use interfaces for component props
interface TodoItemProps {
  todo: Todo;
  onUpdate: (id: number, data: UpdateTodoData) => void;
  onDelete: (id: number) => void;
}

// Prefer const assertions
const TODO_FILTERS = {
  ALL: "all",
  ACTIVE: "active",
  COMPLETED: "completed",
} as const;
```

#### React Components
```tsx
// Functional components with TypeScript
export function TodoItem({ todo, onUpdate, onDelete }: TodoItemProps) {
  // Component logic
}

// Use destructuring for props
// Use meaningful variable names
// Keep components focused and small
```

#### File Organization
```typescript
// 1. Imports (external libraries first, then internal)
import React, { useState } from "react";
import { format } from "date-fns";
import { TodoApiClient } from "@/features/todo/lib/api-client";
import { Button } from "@/components/ui/button";

// 2. Types/Interfaces
interface Props {
  // ...
}

// 3. Component
export function Component() {
  // ...
}

// 4. Helper functions (if needed)
function helperFunction() {
  // ...
}
```

### Backend Standards

#### Ruby Style
Follow standard Ruby style guide with these specifics:
```ruby
# Use 2 spaces for indentation
# Use snake_case for methods and variables
# Use CamelCase for classes and modules

class TodosController < ApplicationController
  before_action :authenticate_user!
  before_action :set_todo, only: [:show, :update, :destroy]

  def index
    @todos = current_user.todos.ordered
    render json: @todos
  end

  private

  def set_todo
    @todo = current_user.todos.find(params[:id])
  end

  def todo_params
    params.require(:todo).permit(:title, :completed, :due_date)
  end
end
```

#### Rails Conventions
1. **Skinny Controllers**: Keep business logic in models or services
2. **RESTful Routes**: Follow REST conventions
3. **Strong Parameters**: Always use parameter filtering
4. **Scopes**: Use scopes for common queries

## Naming Conventions

### Frontend

#### Files and Directories
- **Components**: PascalCase (`TodoItem.tsx`)
- **Hooks**: camelCase with `use` prefix (`useTodos.ts`)
- **Utilities**: kebab-case (`api-client.ts`)
- **Types**: kebab-case (`todo.ts`)
- **Constants**: kebab-case (`constants.ts`)

#### Code
- **Components**: PascalCase (`TodoList`)
- **Functions**: camelCase (`handleSubmit`)
- **Variables**: camelCase (`isLoading`)
- **Constants**: UPPER_SNAKE_CASE (`API_BASE_URL`)
- **Types/Interfaces**: PascalCase (`TodoItemProps`)

### Backend

#### Files
- **Models**: singular, snake_case (`user.rb`, `todo.rb`)
- **Controllers**: plural, snake_case (`todos_controller.rb`)
- **Migrations**: timestamp_description (`20240101000000_create_todos.rb`)

#### Code
- **Classes**: CamelCase (`TodosController`)
- **Methods**: snake_case (`update_position`)
- **Variables**: snake_case (`current_user`)
- **Constants**: UPPER_SNAKE_CASE (`DEFAULT_POSITION`)

## Git Workflow

### Branch Naming
- `feature/description` - New features
- `fix/description` - Bug fixes
- `refactor/description` - Code refactoring
- `docs/description` - Documentation updates

### Commit Messages
Follow conventional commits:
```
feat: add drag and drop to todo list
fix: resolve todo deletion error
refactor: extract todo api client
docs: update API documentation
test: add todo controller specs
```

### Pull Request Process
1. Create feature branch
2. Make changes with clear commits
3. Run quality checks locally:
   - Frontend: `pnpm run lint`, `pnpm run typecheck`
   - Backend: `bundle exec rspec`, `bundle exec rubocop`
4. Push branch and create PR
5. Ensure CI checks pass (tests, RuboCop)
6. Request code review
7. Merge after approval

## Testing Guidelines

### Frontend Testing
```typescript
// Component tests
describe("TodoItem", () => {
  it("displays todo title", () => {
    // Test implementation
  });

  it("handles completion toggle", () => {
    // Test implementation
  });
});

// Hook tests
describe("useTodos", () => {
  it("fetches todos on mount", () => {
    // Test implementation
  });
});
```

### Backend Testing
```ruby
# Model specs
RSpec.describe Todo, type: :model do
  it { should belong_to(:user) }
  it { should validate_presence_of(:title) }
end

# Request specs
RSpec.describe "Todos API", type: :request do
  describe "GET /api/todos" do
    it "returns user todos" do
      get api_todos_path, headers: auth_headers
      expect(response).to have_http_status(:ok)
    end
  end
end
```

## Debugging Tips

### Frontend Debugging
1. Use React Developer Tools
2. Check Network tab for API calls
3. Use `console.log` strategically
4. Check for TypeScript errors

### Backend Debugging
1. Use `rails console` for interactive debugging
2. Check `log/development.log`
3. Use `byebug` for breakpoints
4. Check SQL queries in logs

### Docker Debugging
```bash
# View container logs
docker compose logs -f frontend
docker compose logs -f backend

# Access container shell
docker compose exec frontend sh
docker compose exec backend bash

# Check container status
docker compose ps
```

## Performance Best Practices

### Frontend
1. Use React.memo for expensive components
2. Implement pagination for large lists
3. Optimize bundle size
4. Use proper image formats

### Backend
1. Use database indexes properly
2. Implement eager loading (includes)
3. Cache expensive operations
4. Use background jobs for heavy tasks

## Security Best Practices

### Frontend
1. Sanitize user input
2. Use HTTPS in production
3. Store tokens securely
4. Implement proper CORS

### Backend
1. Use strong parameters
2. Validate all input
3. Use parameterized queries
4. Keep dependencies updated

## Environment Variables

### Frontend
Create `.env.local`:
```
NEXT_PUBLIC_API_URL=http://localhost:3001
```

### Backend
Create `.env`:
```
DATABASE_URL=postgres://user:pass@localhost/dbname
RAILS_MASTER_KEY=your_key_here
```