FactoryBot.define do
  factory :note do
    association :user
    sequence(:title) { |n| "Note #{n}" }
    body_md { "Body content #{SecureRandom.hex(4)}" }
    pinned { false }
    archived_at { nil }
    trashed_at { nil }
    last_edited_at { Time.current }
  end
end
