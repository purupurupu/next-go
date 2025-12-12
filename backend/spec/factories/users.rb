FactoryBot.define do
  factory :user do
    sequence(:email) { |n| "user#{n}@example.com" }
    password { 'password123' }
    password_confirmation { 'password123' }
    sequence(:name) { |n| "Test User #{n}" }

    # Use transient attributes for flexible todo creation
    transient do
      todos_count { 3 }
      skip_todos { true }
    end

    trait :with_todos do
      skip_todos { false }

      after(:create) do |user, evaluator|
        create_list(:todo, evaluator.todos_count, user: user, skip_user: true)
      end
    end

    # Minimal factory for performance
    trait :minimal do
      name { nil }
    end
  end
end
