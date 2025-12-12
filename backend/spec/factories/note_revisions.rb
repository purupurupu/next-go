FactoryBot.define do
  factory :note_revision do
    association :note
    association :user
    sequence(:title) { |n| "Revision #{n}" }
    body_md { "Revision body #{SecureRandom.hex(3)}" }
  end
end
