FactoryBot.define do
  factory :category do
    # Use transient attributes for conditional association
    transient do
      skip_user { false }
    end

    # Lazy evaluation of user association
    user { skip_user ? nil : association(:user) }

    sequence(:name) { |n| "Category #{n}" }
    color { '#3B82F6' }

    trait :work do
      name { 'Work' }
      color { '#EF4444' }
    end

    trait :personal do
      name { 'Personal' }
      color { '#10B981' }
    end

    trait :urgent do
      name { 'Urgent' }
      color { '#F59E0B' }
    end

    # For build_stubbed usage
    trait :stubbed do
      to_create { |instance| instance.id = instance.class.generate_id }
    end
  end

  # Helper method for generating IDs for stubbed instances
  def self.generate_id
    @generated_category_id ||= 2000
    @generated_category_id += 1
  end
end
