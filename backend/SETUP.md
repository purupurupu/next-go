# Rich Todo App - Backend Setup

This document describes the enhanced project structure and dependencies setup for the Rich Todo App.

## Dependencies Added

### Core Gems

- **devise**: User authentication system
- **jwt**: JSON Web Token implementation for API authentication
- **active_model_serializers**: JSON API serialization
- **sidekiq**: Background job processing
- **redis**: Caching and job queue storage
- **image_processing**: File attachment processing

### Testing Gems

- **rspec-rails**: Testing framework
- **factory_bot_rails**: Test data factories
- **faker**: Fake data generation
- **shoulda-matchers**: Additional RSpec matchers
- **database_cleaner-active_record**: Database cleanup for tests

## Configuration

### Database

- PostgreSQL configured for development and test environments
- Active Storage configured for file attachments

### Background Jobs

- Sidekiq configured with Redis for background processing
- Active Job adapter set to Sidekiq

### Authentication

- Devise configured for API-only mode
- JWT service ready for implementation

### Testing

- RSpec configured with Factory Bot
- Shoulda Matchers integrated
- Database Cleaner configured for test isolation
- Authentication helpers prepared

## Docker Commands

### 基本的な操作

```bash
# サービス起動
docker compose up -d

# サービス停止
docker compose down

# ログ確認
docker compose logs -f backend
```

### テスト実行

```bash
# 全テスト実行
docker compose --profile test up backend-test

# 特定のテストファイル実行
docker compose exec backend bundle exec rspec spec/models/todo_spec.rb

# テストを対話的に実行
docker compose exec backend bash
# コンテナ内で: bundle exec rspec
```

### 開発コマンド

```bash
# Rails console
docker compose exec backend rails console

# マイグレーション実行
docker compose exec backend rails db:migrate

# テストDB準備
docker compose run --rm backend-test rails db:test:prepare

# 任意のRailsコマンド実行
docker compose exec backend rails [command]

# バックエンドコンテナのシェルにアクセス
docker compose exec backend bash

# 新しいマイグレーション作成
docker compose exec backend rails generate migration [name]

# 新しいモデル作成
docker compose exec backend rails generate model [name]
```

### よく使うコマンド例

```bash
# 開発環境でテスト実行
docker compose exec backend bundle exec rspec

# 特定のテストのみ実行
docker compose exec backend bundle exec rspec spec/models/todo_spec.rb

# テストを詳細表示で実行
docker compose exec backend bundle exec rspec --format documentation

# Railsサーバー再起動
docker compose restart backend

# データベースリセット
docker compose exec backend rails db:drop db:create db:migrate
```

## Next Steps

The project structure is now ready for implementing:

1. User authentication system with Devise and JWT
2. Enhanced Todo model with categories, tags, and priorities
3. File attachment system with Active Storage
4. Background job processing with Sidekiq
5. Comprehensive test suite with RSpec and Factory Bot
