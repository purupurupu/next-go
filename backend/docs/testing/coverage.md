# Code Coverage with SimpleCov

This Rails application uses SimpleCov to track code coverage for the test suite.

## Installation

SimpleCov is already configured in the Gemfile. To install dependencies:

```bash
docker compose exec backend bundle install
```

## Running Tests with Coverage

### Method 1: Using Environment Variable

```bash
# Run all tests with coverage
docker compose exec backend env COVERAGE=true RAILS_ENV=test bundle exec rspec

# Run specific test file with coverage
docker compose exec backend env COVERAGE=true RAILS_ENV=test bundle exec rspec spec/models/todo_spec.rb
```

### Method 2: Using Rake Task

```bash
# Run all tests with coverage
docker compose exec backend bundle exec rake coverage

# Or use the full task name
docker compose exec backend bundle exec rake coverage:rspec

# Run tests and automatically open HTML report (macOS only)
docker compose exec backend bundle exec rake coverage:rspec_open

# Clean coverage reports
docker compose exec backend bundle exec rake coverage:clean
```

## Viewing Coverage Reports

After running tests with coverage, reports are generated in the `coverage/` directory:

1. **HTML Report**: `coverage/index.html` - Detailed line-by-line coverage
2. **JSON Report**: `coverage/coverage.json` - For CI integration
3. **Summary**: `coverage/.last_run.json` - Quick coverage percentage

### Opening HTML Report

```bash
# macOS
open backend/coverage/index.html

# Linux
xdg-open backend/coverage/index.html

# Windows
start backend/coverage/index.html
```

### Quick Coverage Check

```bash
# View coverage percentage
cat backend/coverage/.last_run.json

# Display only the percentage
docker compose exec backend ruby -rjson -e "puts JSON.parse(File.read('coverage/.last_run.json'))['result']['line'].to_s + '%'"
```

### Understanding the HTML Report

- **Green lines**: Covered by tests
- **Red lines**: Not covered by tests  
- **Yellow lines**: Partially covered (branches)
- Click on any file name to see detailed line-by-line coverage

## Configuration

SimpleCov is configured in `spec/spec_helper.rb` with:

- **Current Coverage**: 91.73% ✅
- **Minimum Coverage**: 90% (enforced in CI/CD)
- **Excluded Directories**: `/spec/`, `/config/`, `/vendor/`, `/db/`, `/lib/tasks/`
- **Coverage Groups**: Models, Controllers, Serializers, Services, Errors, Middleware, etc.
- **Output Formats**: HTML and JSON

### Current Coverage Status

The application has achieved 91.73% code coverage. Well-tested areas include:
- Error handling middleware (93.44%)
- All error classes (100%)
- Models (90%+)
- Services (95%+)

Areas for potential improvement:
- ApplicationController (77.59%) - JWT expiration handling
- API Response Formatter (66.67%) - Error response edge cases
- Sessions Controller (81.25%) - Logout error scenarios

## CI/CD Integration

### GitHub Actions

The project includes a comprehensive test workflow (`.github/workflows/test.yml`) that:

1. Runs all backend tests with coverage
2. Checks if coverage meets the 90% threshold
3. Uploads coverage reports as artifacts
4. Comments on PRs with coverage results
5. Fails the build if coverage drops below 90%

For CI environments, SimpleCov will automatically run when the `CI` environment variable is set:

```bash
CI=true bundle exec rspec
```

### Viewing CI Coverage Results

- **Pull Requests**: Coverage percentage in PR comments
- **Actions Tab**: Download coverage artifacts
- **Job Summary**: Coverage shown in workflow summary

## Coverage Best Practices

1. **Aim for High Coverage**: Target 90%+ coverage for critical code
2. **Focus on Business Logic**: Prioritize coverage for models, services, and controllers
3. **Don't Chase 100%**: Some code (like Rails configurations) doesn't need coverage
4. **Review Uncovered Code**: Use coverage reports to identify untested edge cases
5. **Combine with Other Metrics**: Coverage alone doesn't guarantee quality tests

## Troubleshooting

### Coverage Not Generated

Ensure you're setting the `COVERAGE` environment variable:
```bash
COVERAGE=true bundle exec rspec
```

### Coverage Report Not Found

Check that tests are actually running and the `coverage/` directory exists:
```bash
ls -la coverage/
```

### Low Coverage Failures

SimpleCov will fail the test suite if coverage drops below 90%. To see detailed coverage:
1. Open `coverage/index.html`
2. Click on files with low coverage
3. Red lines indicate uncovered code

## Ignoring Files

To exclude files from coverage, update `spec/spec_helper.rb`:

```ruby
SimpleCov.start 'rails' do
  add_filter '/path/to/exclude/'
  # ... other configuration
end
```