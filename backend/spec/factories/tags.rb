FactoryBot.define do
  factory :tag do
    # Use transient attributes for conditional association
    transient do
      skip_user { false }
    end

    # Lazy evaluation of user association
    user { skip_user ? nil : association(:user) }

    sequence(:name) { |n| "tag-#{n}" }
    color { '#FF0000' }

    trait :with_todos do
      transient do
        todos_count { 3 }
        shared_user { nil }
      end

      after(:create) do |tag, evaluator|
        # Use shared user if provided to avoid creating multiple users
        user_for_todos = evaluator.shared_user || tag.user
        evaluator.todos_count.times do
          todo = create(:todo, user: user_for_todos, skip_user: true)
          create(:todo_tag, tag: tag, todo: todo, skip_associations: true)
        end
      end
    end

    trait :without_color do
      color { nil }
    end

    # For build_stubbed usage
    trait :stubbed do
      to_create { |instance| instance.id = instance.class.generate_id }
    end
  end

  # Helper method for generating IDs for stubbed instances
  def self.generate_id
    @generated_tag_id ||= 3000
    @generated_tag_id += 1
  end
end
