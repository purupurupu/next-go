require 'rails_helper'

RSpec.describe Category, type: :model do
  let(:user) { create(:user) }

  describe 'associations' do
    it { is_expected.to belong_to(:user) }
    it { is_expected.to have_many(:todos).dependent(:nullify) }
  end

  describe 'validations' do
    subject { build(:category, user: user) }

    it { is_expected.to validate_presence_of(:name) }
    it { is_expected.to validate_length_of(:name).is_at_most(50) }
    it { is_expected.to validate_presence_of(:color) }

    it 'validates color format' do
      category = build(:category, user: user, color: 'invalid')
      expect(category).not_to be_valid
      expect(category.errors[:color]).to include('must be a valid hex color')
    end

    it 'validates name uniqueness per user' do
      create(:category, user: user, name: 'Work')
      duplicate = build(:category, user: user, name: 'work')
      expect(duplicate).not_to be_valid
      expect(duplicate.errors[:name]).to include('has already been taken')
    end

    it 'allows same name for different users' do
      another_user = create(:user, email: 'another@example.com')
      create(:category, user: user, name: 'Work')
      duplicate = build(:category, user: another_user, name: 'Work')
      expect(duplicate).to be_valid
    end
  end

  describe 'color normalization' do
    it 'normalizes color to uppercase' do
      category = create(:category, user: user, color: '#ff0000')
      expect(category.color).to eq('#FF0000')
    end
  end

  describe 'valid colors' do
    it 'accepts 6-digit hex colors' do
      category = build(:category, user: user, color: '#FF0000')
      expect(category).to be_valid
    end

    it 'accepts 3-digit hex colors' do
      category = build(:category, user: user, color: '#F00')
      expect(category).to be_valid
    end
  end
end
