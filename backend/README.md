# Rails Todo API Backend

This is the Rails API backend for the Todo application.

## Ruby Version

* Ruby 3.2.5
* Rails 7.1.3+

## System Dependencies

* PostgreSQL 15
* Redis (for Sidekiq background jobs)
* Docker and Docker Compose

## Getting Started

See the main project README and CLAUDE.md for detailed setup instructions.

## Running Tests

```bash
# Run all tests
docker compose exec backend env RAILS_ENV=test bundle exec rspec

# Run specific test file
docker compose exec backend env RAILS_ENV=test bundle exec rspec spec/models/todo_spec.rb

# Run tests with documentation format
docker compose exec backend env RAILS_ENV=test bundle exec rspec --format documentation
```

## Code Coverage

This project uses SimpleCov for code coverage reporting.

```bash
# Run tests with coverage
docker compose exec backend env COVERAGE=true RAILS_ENV=test bundle exec rspec

# Or use the rake task
docker compose exec backend bundle exec rake coverage

# View coverage report
open backend/coverage/index.html  # macOS
xdg-open backend/coverage/index.html  # Linux
```

See [docs/testing/coverage.md](docs/testing/coverage.md) for detailed coverage documentation.

## Services

* **Sidekiq**: Background job processing
* **Redis**: Caching and job queue backend
* **PostgreSQL**: Primary database

## API Documentation

See the [docs/api/](../docs/api/) directory for API endpoint documentation.
