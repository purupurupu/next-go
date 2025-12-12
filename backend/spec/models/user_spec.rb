require 'rails_helper'

RSpec.describe User, type: :model do
  # 学習ポイント：モデルテストのベストプラクティス

  describe 'associations' do
    it { is_expected.to have_many(:todos).dependent(:destroy) }
    it { is_expected.to have_many(:categories).dependent(:destroy) }
  end

  describe 'validations' do
    subject { build(:user) }

    it { is_expected.to validate_presence_of(:name) }
    it { is_expected.to validate_length_of(:name).is_at_least(2).is_at_most(50) }
    it { is_expected.to validate_presence_of(:email) }
    it { is_expected.to validate_uniqueness_of(:email).case_insensitive }
    it { is_expected.to validate_presence_of(:password) }
  end

  describe 'devise modules' do
    it 'includes database_authenticatable module' do
      expect(described_class.devise_modules).to include(:database_authenticatable)
    end

    it 'includes registerable module' do
      expect(described_class.devise_modules).to include(:registerable)
    end

    it 'includes recoverable module' do
      expect(described_class.devise_modules).to include(:recoverable)
    end

    it 'includes rememberable module' do
      expect(described_class.devise_modules).to include(:rememberable)
    end

    it 'includes validatable module' do
      expect(described_class.devise_modules).to include(:validatable)
    end
  end

  describe 'factory' do
    it 'creates a valid user' do
      user = build(:user)
      expect(user).to be_valid
    end
  end

  describe 'email case insensitivity' do
    it 'does not allow duplicate emails with different cases' do
      create(:user, email: 'test@example.com')
      duplicate_user = build(:user, email: 'TEST@EXAMPLE.COM')
      expect(duplicate_user).not_to be_valid
      expect(duplicate_user.errors[:email]).to include('has already been taken')
    end
  end
end
