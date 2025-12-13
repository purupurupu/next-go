# Repository Guidelines

## Project Structure & Module Organization
- `backend/`: Go 1.25 API using Echo. Entry at `cmd/api/main.go`; domain packages in `internal/{config,handler,middleware,model,repository,service,validator}`; shared DB helpers in `pkg/database`; migrations placeholder in `db/migrations/`. Tests live alongside code as `*_test.go` (e.g., `internal/handler/auth_test.go`).
- `frontend/`: Next.js 16 app under `src/` (`src/app` for routes/components, `src/lib` for shared helpers); static assets in `public/`.
- `compose.yml` starts the full stack (frontend:3000, backend:3001, Postgres:5432, Redis:6379); additional docs in `docs/`.

## Build, Test, and Development Commands
- Full stack dev: `docker compose up --build` (rebuilds frontend/backend images and runs dev servers).
- Backend tests: `docker compose exec backend go test ./...` (add `-cover` for coverage).
- Backend build: `docker compose exec backend go build -o bin/api cmd/api/main.go`.
- Frontend dev only: `cd frontend && pnpm dev`.
- Frontend checks: `pnpm lint`, `pnpm typecheck`; production build `pnpm build`, start `pnpm start`.

## Coding Style & Naming Conventions
- Go: format with `gofmt` (and `go fmt ./...` before committing); keep handlers/repositories small and inject dependencies; keep package names lower_snake; tests mirror packages with `_test.go`.
- TypeScript/React: ESM with named exports; pages/routes may use default export only when Next.js requires it; components use PascalCase files/folders under `src/app/**`; shared utilities stay in `src/lib`; prefer explicit types.
- Keep comments minimal and purposeful; avoid committing generated artifacts.

## Testing Guidelines
- Backend: `go test ./...` expects Postgres reachable at `postgres://postgres:password@localhost:5432/todo_next` (docker compose provides this). Seed/migrate via auto-migrate in dev; clean up test data in helpers when needed.
- Frontend: no harness yet—if adding tests, use Vitest + Testing Library co-located in `__tests__` near the component.

## Commit & Pull Request Guidelines
- Commits: light Conventional Commits style (e.g., `feat(auth): …`, `fix: …`, `chore: …`, `style: …`); keep scopes meaningful.
- PRs: include a short summary, test commands/results, linked issues, and UI screenshots for visible changes; call out database migrations or breaking API changes explicitly.

## Security & Configuration Tips
- Do not commit secrets; configure `JWT_SECRET`, DB credentials, etc., via environment or `.env` (ignored by git). Copy `.env.example` to get started.
- For local Docker runs, ensure Postgres/Redis ports are free to avoid container start failures; set `PORT`/`ENV` as needed in `compose.yml`.
