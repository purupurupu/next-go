# CI/CD Documentation

## Overview

This project uses GitHub Actions for continuous integration and deployment. The CI pipeline ensures code quality, runs tests, and checks for linting errors before merging changes.

## Workflows

### Test Suite (`.github/workflows/test.yml`)

The main test workflow runs on:
- Push to `main` or `develop` branches
- Pull requests to `main` branch

#### Backend Tests
1. Sets up PostgreSQL service
2. Installs Ruby dependencies
3. Creates and migrates test database
4. Runs RSpec tests with coverage
5. Checks coverage threshold (90%)
6. Runs RuboCop for code quality

#### Frontend Tests
1. Installs Node.js and pnpm
2. Installs dependencies
3. Runs ESLint
4. Runs TypeScript type checking
5. Builds the frontend application


## Required Secrets

The following secrets must be configured in GitHub repository settings:

- `RAILS_MASTER_KEY`: Rails master key for encrypted credentials
- `SECRET_KEY_BASE`: (Optional) Secret key base for Rails

## Local CI Checks

Before pushing changes, run these commands locally:

### Backend
```bash
docker compose exec backend bundle exec rspec
docker compose exec backend bundle exec rubocop
```

### Frontend
```bash
docker compose exec frontend pnpm run lint
docker compose exec frontend pnpm run typecheck
```

## Coverage Requirements

- Backend: Minimum 90% line coverage
- Coverage reports are uploaded as artifacts for each test run

## Best Practices

1. Always ensure CI passes before merging
2. Keep test coverage above the threshold
3. Fix RuboCop violations promptly
4. Run local checks before pushing
5. Update CI configuration when adding new dependencies