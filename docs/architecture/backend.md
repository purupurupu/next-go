# Backend Architecture

## Technology Stack

- **Framework**: Ruby on Rails 7.1.3+ (API-only mode)
- **Language**: Ruby 3.2.5
- **Database**: PostgreSQL 15
- **Web Server**: Puma
- **Authentication**: Devise + Devise-JWT
- **Background Jobs**: Sidekiq with Redis
- **Testing**: RSpec, FactoryBot, Faker

## Key Gems

### Core
- `rails ~> 7.1.3` - Web framework
- `pg ~> 1.1` - PostgreSQL adapter
- `puma >= 5.0` - Web server
- `rack-cors` - CORS handling
- `bootsnap` - Boot time optimization

### Authentication & Security
- `devise` - Authentication solution
- `devise-jwt` - JWT token authentication

### API & Serialization
- `active_model_serializers ~> 0.10.0` - JSON serialization

### Background Processing
- `sidekiq` - Background job processing
- `redis >= 4.0.1` - In-memory data store

### Development & Testing
- `rspec-rails` - Testing framework
- `factory_bot_rails` - Test data factories
- `faker` - Fake data generation
- `shoulda-matchers` - RSpec matchers
- `database_cleaner-active_record` - Test database cleaning

## Directory Structure

```
backend/
├── app/
│   ├── controllers/
│   │   ├── api/
│   │   │   └── v1/
│   │   │       ├── todos_controller.rb     # Todo CRUD endpoints + search
│   │   │       ├── categories_controller.rb # Category CRUD endpoints
│   │   │       ├── tags_controller.rb      # Tag CRUD endpoints
│   │   │       ├── comments_controller.rb  # Comment CRUD endpoints
│   │   │       └── todo_histories_controller.rb # History viewing
│   │   ├── users/
│   │   │   ├── sessions_controller.rb  # Login/logout
│   │   │   └── registrations_controller.rb # Signup
│   │   ├── concerns/
│   │   │   └── api_response_formatter.rb # Unified API response formatting
│   │   └── application_controller.rb   # Base controller with unified error handling
│   ├── models/
│   │   ├── user.rb                     # User model with Devise
│   │   ├── todo.rb                     # Todo model with associations & file attachments
│   │   ├── category.rb                 # Category model with counter_cache
│   │   ├── tag.rb                      # Tag model for flexible labeling
│   │   ├── todo_tag.rb                 # Junction table for todo-tag relationship
│   │   ├── comment.rb                  # Polymorphic comments with soft delete
│   │   ├── todo_history.rb             # Audit trail for todo changes
│   │   └── jwt_denylist.rb             # JWT revocation
│   ├── serializers/
│   │   ├── user_serializer.rb          # User JSON serialization
│   │   ├── todo_serializer.rb          # Todo JSON serialization
│   │   ├── category_serializer.rb      # Category JSON serialization
│   │   ├── tag_serializer.rb           # Tag JSON serialization
│   │   ├── comment_serializer.rb       # Comment JSON serialization
│   │   └── todo_history_serializer.rb  # History JSON serialization
│   └── services/                       # Business logic services
│       └── todo_search_service.rb      # Advanced search implementation
├── config/
│   ├── routes.rb                       # API routes
│   ├── initializers/
│   │   ├── cors.rb                     # CORS configuration
│   │   ├── devise.rb                   # Devise settings
│   │   └── sidekiq.rb                  # Background jobs
│   └── database.yml                    # Database configuration
├── db/
│   ├── migrate/                        # Database migrations
│   └── schema.rb                       # Current schema
└── spec/                               # Test suite
    ├── models/                         # Model tests
    ├── requests/                       # API tests
    └── factories/                      # Test factories
```

## Authentication Architecture

### JWT Token Flow
```
1. User Login (POST /auth/sign_in)
   ↓
2. Validate credentials with Devise
   ↓
3. Generate JWT token
   ↓
4. Return token in response body
   ↓
5. Client stores token
   ↓
6. Client sends token in Authorization header
   ↓
7. Rails validates token on each request
```

### Token Management
- **Generation**: Warden::JWTAuth::UserEncoder
- **Storage**: Client-side (localStorage)
- **Validation**: Each API request
- **Revocation**: JWT denylist table
- **Expiration**: Configurable (default: 24 hours)

## API Design

### RESTful Endpoints
```ruby
# config/routes.rb
namespace :api do
  namespace :v1 do
    resources :todos do
      collection do
        patch 'update_order'  # Bulk position update
        get 'search'          # Advanced search with filtering
      end
      member do
        patch 'tags'          # Update todo tags
        delete 'files/:file_id', to: 'todos#destroy_file' # Delete file attachment
      end
      resources :comments, only: [:index, :create, :update, :destroy]
      resources :histories, controller: 'todo_histories', only: [:index]
    end
    resources :categories
    resources :tags
  end
end
```

### Controller Pattern
```ruby
class Api::V1::TodosController < ApplicationController
  before_action :authenticate_user!
  before_action :set_todo, only: [:show, :update, :destroy, :tags, :destroy_file]

  def index
    @todos = current_user.todos
                        .includes(:category, :tags, files_attachments: :blob)
                        .order(:position)
    render json: @todos
  end

  def search
    result = TodoSearchService.new(current_user, search_params).call
    render json: result
  end

  private

  def set_todo
    @todo = current_user.todos.find(params[:id])
  end

  def todo_params
    params.require(:todo).permit(:title, :completed, :due_date, :priority, 
                                  :status, :description, :category_id, 
                                  tag_ids: [], files: [])
  end

  def search_params
    params.permit(:q, :category_id, :sort_by, :sort_order, :page, :per_page,
                  :due_date_from, :due_date_to, :tag_mode,
                  status: [], priority: [], tag_ids: [])
  end
end
```

## Model Architecture

### User Model
```ruby
class User < ApplicationRecord
  devise :database_authenticatable, :registerable,
         :jwt_authenticatable, jwt_revocation_strategy: JwtDenylist
  
  has_many :todos, dependent: :destroy
  
  validates :email, presence: true, uniqueness: true
  validates :name, presence: true
end
```

### Todo Model
```ruby
class Todo < ApplicationRecord
  belongs_to :user
  belongs_to :category, optional: true, counter_cache: true
  has_many :todo_tags, dependent: :destroy
  has_many :tags, through: :todo_tags
  has_many :comments, as: :commentable, dependent: :destroy
  has_many :todo_histories, dependent: :destroy
  has_many_attached :files
  
  enum priority: { low: 0, medium: 1, high: 2 }
  enum status: { pending: 0, in_progress: 1, completed: 2 }
  
  validates :title, presence: true
  validates :due_date, comparison: { greater_than: Date.today }, 
            allow_nil: true, on: :create
  validates :priority, inclusion: { in: priorities.keys }
  validates :status, inclusion: { in: statuses.keys }
  
  before_create :set_position
  after_create :track_creation
  after_update :track_update
  
  scope :ordered, -> { order(:position) }
  scope :active, -> { where(completed: false) }
  scope :completed, -> { where(completed: true) }
end
```

## Error Handling (UNIFIED)

### Unified Error Handling in ApplicationController
```ruby
class ApplicationController < ActionController::API
  rescue_from ActionController::ParameterMissing, with: :handle_parameter_missing
  rescue_from ActiveRecord::RecordNotFound, with: :handle_not_found
  rescue_from ActiveRecord::RecordInvalid, with: :handle_unprocessable_entity
  rescue_from ActiveRecord::RecordNotUnique, with: :handle_unprocessable_entity
  
  private
  
  def handle_parameter_missing(exception)
    error_response(message: "Parameter missing: #{exception.param}", status: :bad_request)
  end
  
  def handle_not_found(exception)
    error_response(message: 'Record not found', status: :not_found)
  end
  
  def handle_unprocessable_entity(exception)
    if exception.respond_to?(:record) && exception.record&.errors&.any?
      render json: { errors: exception.record.errors }, status: :unprocessable_entity
    else
      error_response(message: exception.message, status: :unprocessable_entity)
    end
  end
end
```

### ApiResponseFormatter Concern
```ruby
module ApiResponseFormatter
  extend ActiveSupport::Concern
  
  private
  
  def success_response(message:, data: nil, status: :ok)
    response_body = { message: message }
    response_body[:data] = data if data
    render json: response_body, status: status
  end
  
  def error_response(message:, status: :unprocessable_entity)
    render json: { error: message }, status: status
  end
end
```

### Standard Error Response
```json
{
  "error": "Record not found"
}
```

### Success Response with Data
```json
{
  "message": "Todo created successfully",
  "data": {
    "id": 1,
    "title": "New Todo",
    "completed": false
  }
}
```

### Validation Error Response
```json
{
  "errors": {
    "title": ["can't be blank"],
    "due_date": ["must be in the future"]
  }
}
```

## Background Jobs

### Sidekiq Configuration
- Redis-backed job queue
- Configured for async operations
- Future use: email notifications, data cleanup

## Security Considerations

1. **Authentication**: All API endpoints require JWT token
2. **Authorization**: Users can only access their own todos
3. **CORS**: Restricted to frontend origin
4. **Parameter Filtering**: Strong parameters in controllers
5. **SQL Injection**: ActiveRecord parameterized queries

## Performance Optimizations (ENHANCED)

1. **Database Indexes**
   - User email (unique)
   - Todo position, user_id, category_id, priority, status (for filtering)
   - Category user_id with unique constraint on (user_id, name)
   - Tag user_id with unique constraint on (user_id, name)
   - TodoTag todo_id, tag_id with unique constraint
   - Comment (commentable_type, commentable_id) for polymorphic queries
   - TodoHistory todo_id, user_id, action for audit queries
   - JWT jti (denylist lookup)
   - Full-text search indexes on todo title and description

2. **Query Optimization**
   - **Counter Cache**: `todos_count` on categories eliminates N+1 queries
   - **Bulk Updates**: `Todo.update_order` for efficient position updates
   - **Search Service**: Advanced filtering with optimized queries
   - Eager loading associations with `includes(:category, :tags, files_attachments: :blob)`
   - Scoped queries for filtering
   - Ordered by position for consistency

3. **N+1 Query Prevention**
   - Counter cache implementation:
   ```ruby
   # Category model
   has_many :todos, counter_cache: true
   
   # Migration adds todos_count column
   add_column :categories, :todos_count, :integer, default: 0, null: false
   ```
   - Eager loading in search:
   ```ruby
   todos.includes(:category, :tags, :user)
   ```

4. **Bulk Operations**
   ```ruby
   # TodosController#update_order
   def update_order
     Todo.transaction do
       params[:todos].each_with_index do |todo_data, index|
         current_user.todos.find(todo_data[:id]).update!(position: index)
       end
     end
   end
   ```

5. **Caching Strategy**
   - Redis for future caching needs
   - HTTP caching headers
   - Counter cache for aggregate queries
   - Active Storage blob caching for file attachments

6. **Search Performance**
   - TodoSearchService with optimized queries
   - Database indexes for common filter combinations
   - Pagination to limit result set size
   - Search result highlighting for better UX

## Testing Strategy

### RSpec Test Suite
```ruby
# Model specs
describe Todo do
  it { should belong_to(:user) }
  it { should validate_presence_of(:title) }
end

# Request specs
describe "POST /api/todos" do
  it "creates a new todo" do
    post api_todos_path, params: { todo: attributes }
    expect(response).to have_http_status(:created)
  end
end
```

### Factory Bot
```ruby
FactoryBot.define do
  factory :todo do
    title { Faker::Lorem.sentence }
    completed { false }
    due_date { 1.week.from_now }
    user
  end
end
```