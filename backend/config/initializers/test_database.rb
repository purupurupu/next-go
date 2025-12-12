# Override DATABASE_URL for test environment to ensure test database is used
# In Docker environment, use 'db' as host; in CI environment, DATABASE_URL is already set correctly
if Rails.env.test? && ENV['DATABASE_URL'].blank?
  ENV['DATABASE_URL'] = 'postgres://postgres:password@db:5432/todo_app_test'
end