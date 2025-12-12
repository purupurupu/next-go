FactoryBot.define do
  factory :jwt_denylist do
    jti { SecureRandom.uuid }
    exp { 1.day.from_now }
  end
end
