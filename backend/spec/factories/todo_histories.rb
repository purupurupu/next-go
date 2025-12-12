# frozen_string_literal: true

FactoryBot.define do
  factory :todo_history do
    # Use transient attributes for conditional associations
    transient do
      skip_todo { false }
      skip_user { false }
      shared_user { nil }
      shared_todo { nil }
    end

    # Lazy evaluation of associations
    todo do
      if skip_todo
        nil
      elsif shared_todo
        shared_todo
      elsif shared_user
        association(:todo, user: shared_user, skip_user: true)
      else
        association(:todo)
      end
    end

    user do
      if skip_user
        nil
      elsif shared_user
        shared_user
      elsif todo&.user
        todo.user
      else
        association(:user)
      end
    end

    field_name { 'title' }
    old_value { 'Old Title' }
    new_value { 'New Title' }
    action { 'updated' }

    trait :created do
      field_name { 'created' }
      old_value { nil }
      new_value { 'New Todo' }
      action { 'created' }
    end

    trait :status_changed do
      field_name { 'status' }
      old_value { 'pending' }
      new_value { 'in_progress' }
      action { 'status_changed' }
    end

    trait :priority_changed do
      field_name { 'priority' }
      old_value { 'low' }
      new_value { 'high' }
      action { 'priority_changed' }
    end

    trait :completed do
      field_name { 'completed' }
      old_value { 'false' }
      new_value { 'true' }
      action { 'updated' }
    end

    # For build_stubbed usage
    trait :stubbed do
      to_create { |instance| instance.id = instance.class.generate_id }
    end
  end

  # Helper method for generating IDs for stubbed instances
  def self.generate_id
    @generated_history_id ||= 6000
    @generated_history_id += 1
  end
end
