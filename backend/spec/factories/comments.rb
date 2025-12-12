# frozen_string_literal: true

FactoryBot.define do
  factory :comment do
    # Use transient attributes for conditional associations
    transient do
      skip_user { false }
      skip_commentable { false }
      shared_user { nil }
    end

    # Lazy evaluation of user association
    user do
      if skip_user
        nil
      elsif shared_user
        shared_user
      else
        association(:user)
      end
    end

    # Simple content instead of Faker for performance
    sequence(:content) { |n| "Comment #{n} content" }
    deleted_at { nil }

    # 学習ポイント：ポリモーフィック関連のファクトリー
    # デフォルトではtodoに関連付ける
    for_todo # trait

    trait :for_todo do
      commentable do
        if skip_commentable
          nil
        elsif shared_user
          association(:todo, user: shared_user, skip_user: true)
        elsif user
          association(:todo, user: user, skip_user: true)
        else
          association(:todo)
        end
      end
    end

    trait :deleted do
      deleted_at { 1.hour.ago }
    end

    trait :long do
      content { 'This is a long comment. ' * 10 }
    end

    trait :recent do
      created_at { 1.minute.ago }
    end

    trait :old do
      created_at { 1.month.ago }
    end

    # For build_stubbed usage
    trait :stubbed do
      to_create { |instance| instance.id = instance.class.generate_id }
    end
  end

  # Helper method for generating IDs for stubbed instances
  def self.generate_id
    @generated_comment_id ||= 5000
    @generated_comment_id += 1
  end
end
