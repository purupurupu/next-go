FactoryBot.define do
  factory :todo do
    # Use transient attributes for conditional association
    transient do
      skip_user { false }
    end

    # Lazy evaluation of user association
    user { skip_user ? nil : association(:user) }

    # Simple sequence instead of Faker for performance
    sequence(:title) { |n| "Todo #{n}" }
    completed { false }
    priority { :medium }
    status { :pending }

    # Optional fields are nil by default for performance
    due_date { nil }
    description { nil }
    category { nil }

    # Minimal factory for maximum performance
    trait :minimal do
      # Only required fields
    end

    # Full factory with all fields populated
    trait :full do
      due_date { 7.days.from_now }
      description { 'Todo description' }
      association :category, strategy: :build
    end

    trait :completed do
      completed { true }
      status { :completed }
    end

    trait :overdue do
      due_date { 5.days.ago }
    end

    trait :with_due_date do
      due_date { 7.days.from_now }
    end

    trait :high_priority do
      priority { :high }
    end

    trait :low_priority do
      priority { :low }
    end

    trait :in_progress do
      status { :in_progress }
    end

    trait :with_description do
      description { 'This is a todo description' }
    end

    trait :with_category do
      transient do
        category_user { user }
      end

      category { association :category, user: category_user }
    end

    # For build_stubbed usage
    trait :stubbed do
      to_create { |instance| instance.id = instance.class.generate_id }
    end
  end

  # Helper method for generating IDs for stubbed instances
  def self.generate_id
    @generated_id ||= 1000
    @generated_id += 1
  end
end
