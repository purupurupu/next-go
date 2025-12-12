FactoryBot.define do
  factory :todo_tag do
    # Use transient attributes for conditional associations
    transient do
      skip_associations { false }
      shared_user { nil }
    end

    # Lazy evaluation of associations
    todo do
      if skip_associations
        nil
      elsif shared_user
        association(:todo, user: shared_user, skip_user: true)
      else
        association(:todo)
      end
    end

    tag do
      if skip_associations
        nil
      elsif shared_user
        association(:tag, user: shared_user, skip_user: true)
      elsif todo&.user
        association(:tag, user: todo.user, skip_user: true)
      else
        association(:tag)
      end
    end

    # For build_stubbed usage
    trait :stubbed do
      to_create { |instance| instance.id = instance.class.generate_id }
    end
  end

  # Helper method for generating IDs for stubbed instances
  def self.generate_id
    @generated_todo_tag_id ||= 4000
    @generated_todo_tag_id += 1
  end
end
