# FactoryBot Best Practices

This document outlines best practices for writing efficient FactoryBot factories in this Rails application.

## Performance Optimization Guidelines

### 1. Use Transient Attributes for Conditional Associations

**Bad:**
```ruby
factory :todo do
  association :user
end
```

**Good:**
```ruby
factory :todo do
  transient do
    skip_user { false }
  end
  
  user { skip_user ? nil : association(:user) }
end
```

This allows tests to skip creating unnecessary associated records:
```ruby
# Creates a todo without creating a user
create(:todo, skip_user: true, user: existing_user)
```

### 2. Share Common Data Across Tests

Use shared contexts to avoid creating duplicate records:

```ruby
RSpec.describe TodosController do
  include_context 'with user'
  
  let(:todo) { create(:todo, user: user, skip_user: true) }
  let(:category) { create(:category, user: user, skip_user: true) }
end
```

### 3. Use `build_stubbed` When Database Persistence Isn't Required

**Bad:**
```ruby
let(:todo) { create(:todo) } # Hits the database
```

**Good:**
```ruby
let(:todo) { build_stubbed(:todo) } # No database hit
```

### 4. Replace Faker with Simple Sequences

**Bad:**
```ruby
title { Faker::Lorem.sentence(word_count: 3) }
description { Faker::Lorem.paragraph(sentence_count: 2) }
```

**Good:**
```ruby
sequence(:title) { |n| "Todo #{n}" }
description { "Simple description" }
```

### 5. Make Optional Fields Nil by Default

**Bad:**
```ruby
factory :todo do
  title { "Todo" }
  description { "Long description..." }
  due_date { 7.days.from_now }
end
```

**Good:**
```ruby
factory :todo do
  title { "Todo" }
  description { nil }
  due_date { nil }
  
  trait :with_details do
    description { "Description" }
    due_date { 7.days.from_now }
  end
end
```

### 6. Use Traits for Variations

Instead of creating multiple factories, use traits:

```ruby
factory :todo do
  # Base attributes
  
  trait :completed do
    completed { true }
    status { :completed }
  end
  
  trait :high_priority do
    priority { :high }
  end
end

# Usage
create(:todo, :completed, :high_priority)
```

### 7. Leverage `let_it_be` for Suite-Wide Data

For data that doesn't change between tests:

```ruby
RSpec.describe "Todos API" do
  let_it_be(:user) { create(:user) }
  
  # This user is created once for all tests in this describe block
end
```

## Factory Structure

Each factory in this application follows this pattern:

1. **Transient attributes** for controlling associations
2. **Lazy evaluation** of associations
3. **Simple sequences** instead of Faker
4. **Minimal defaults** with optional fields as nil
5. **Traits** for common variations
6. **Stubbed trait** for build_stubbed support

## Profiling Factory Performance

Run tests with factory profiling enabled:

```bash
PROFILE_FACTORIES=true bundle exec rspec
```

This will show:
- Which factories are used most frequently
- Average time per factory
- Total time spent in each factory
- Warnings for slow factories (>50ms average)

## Common Patterns

### Creating Related Records with Shared User

```ruby
let(:user) { create(:user) }
let(:todo) { create(:todo, user: user, skip_user: true) }
let(:category) { create(:category, user: user, skip_user: true) }
let(:tag) { create(:tag, user: user, skip_user: true) }
```

### Creating Many-to-Many Relationships

```ruby
let(:user) { create(:user) }
let(:todo) { create(:todo, user: user, skip_user: true) }
let(:tag) { create(:tag, user: user, skip_user: true) }
let(:todo_tag) { create(:todo_tag, todo: todo, tag: tag, skip_associations: true) }
```

### Testing Without Database Hits

```ruby
# For unit tests that don't need persistence
let(:todo) { build_stubbed(:todo, :completed) }

# For testing validations
let(:invalid_todo) { build(:todo, title: nil) }
```

## Anti-Patterns to Avoid

1. **Creating unnecessary associations** - Always question if you need the associated record
2. **Using Faker everywhere** - Simple sequences are faster
3. **Setting all attributes** - Only set what's needed for the test
4. **Not using shared data** - Reuse users and other common records
5. **Always using create** - Use build or build_stubbed when possible
6. **Complex after(:create) callbacks** - Keep factories simple

## Maintenance

- Review factory performance monthly using the profiler
- Update this document as new patterns emerge
- Remove unused factories and traits
- Keep factories focused on the minimum required data