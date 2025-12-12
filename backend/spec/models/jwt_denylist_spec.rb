require 'rails_helper'

RSpec.describe JwtDenylist, type: :model do
  describe 'database table' do
    it 'has required columns' do
      expect(described_class.column_names).to include('jti', 'exp')
    end
  end

  describe 'devise-jwt integration' do
    it 'includes RevocationStrategies::Denylist' do
      expect(described_class.included_modules).to include(Devise::JWT::RevocationStrategies::Denylist)
    end
  end

  describe 'factory' do
    it 'creates a valid jwt_denylist' do
      jwt_denylist = create(:jwt_denylist)
      expect(jwt_denylist).to be_valid
    end
  end
end
