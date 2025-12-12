# Repository Guidelines

## Project Structure & Module Organization
- `backend`: Rails 7 API with domain logic under `app/`, background jobs/Sidekiq config in `config/`, and API docs in `docs/`. Specs live in `spec/`, shared support in `spec/support/`.
- `frontend`: Next.js 15 app in `src/` (UI under `src/app`, shared helpers in `src/lib`). Static assets live in `public/`.
- `docs`: Additional reference docs; root `compose.yml` starts the full stack (frontend:3000, backend:3001, Postgres:5432, Redis:6379).

## Build, Test, and Development Commands
- Start full stack: `docker compose up --build` (rebuilds frontend/backend images and runs Next dev + Rails server).
- Backend DB prep: `docker compose exec backend bundle exec rails db:prepare` (setup or migrate).
- Backend tests: `docker compose exec backend env RAILS_ENV=test bundle exec rspec` (SimpleCov via `COVERAGE=true`).
- Backend lint: `docker compose exec backend bundle exec rubocop`.
- Frontend dev only: `cd frontend && pnpm dev` (runs Turbopack at http://localhost:3000).
- Frontend lint/typecheck: `pnpm lint`, `pnpm typecheck`.
- Frontend production build: `pnpm build`; start with `pnpm start`.

## Coding Style & Naming Conventions
- Ruby: 2-space indent, follow RuboCop defaults; RSpec files as `*_spec.rb`. Prefer service objects and POROs under `app/` over fat controllers.
- TypeScript/React: ESM with named exports; components in `src/app/**` use PascalCase folders/files. Keep shared utilities in `src/lib`.
- Favor explicit types; avoid default exports unless a page or Next route requires it.

## Testing Guidelines
- Backend uses RSpec; group examples with descriptive `context`/`it` blocks. Place shared helpers in `spec/support` and require via `.rspec`.
- Aim for meaningful coverage; run `COVERAGE=true` locally to check reports at `backend/coverage/index.html`.
- Frontend has no test harness yet—add Vitest/Testing Library co-located in `__tests__` near components when introducing tests.

## Commit & Pull Request Guidelines
- Commit messages follow a light Conventional Commits style observed in history (e.g., `feat(mcp): …`, `fix: …`, `chore: …`, `style: …`). Keep scopes meaningful.
- PRs should include: brief summary, test commands/results, linked issues, and UI screenshots for visible changes. Call out database migrations or breaking API changes explicitly.

## Security & Configuration
- Keep secrets out of VCS; provide `RAILS_MASTER_KEY`, database credentials, and `SECRET_KEY_BASE` via environment or `.env` files ignored by git.
- For local Docker runs, ensure Postgres/Redis ports are free to avoid container start failures.***
