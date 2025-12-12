require 'rails_helper'

RSpec.describe Tag, type: :model do
  describe 'associations' do
    it { is_expected.to belong_to(:user) }
    it { is_expected.to have_many(:todo_tags).dependent(:destroy) }
    it { is_expected.to have_many(:todos).through(:todo_tags) }
  end

  describe 'validations' do
    subject { build(:tag, user: user) }

    let(:user) { create(:user) }

    it { is_expected.to validate_presence_of(:name) }
    it { is_expected.to validate_length_of(:name).is_at_most(30) }

    context 'name uniqueness' do
      it 'validates uniqueness of name scoped to user (case-insensitive)' do
        create(:tag, name: 'Work', user: user)
        duplicate_tag = build(:tag, name: 'WORK', user: user)
        expect(duplicate_tag).not_to be_valid
        expect(duplicate_tag.errors[:name]).to include('has already been taken')
      end

      it 'allows same name for different users' do
        another_user = create(:user)
        create(:tag, name: 'Work', user: user)
        tag_for_another_user = build(:tag, name: 'Work', user: another_user)
        expect(tag_for_another_user).to be_valid
      end
    end

    context 'color validation' do
      it 'accepts valid hex colors' do
        valid_colors = ['#FF0000', '#00FF00', '#0000FF', '#FFF', '#000', '#abc123', '#ABC']
        valid_colors.each do |color|
          tag = build(:tag, user: user, color: color)
          expect(tag).to be_valid, "Expected #{color} to be valid"
        end
      end

      it 'rejects invalid hex colors' do
        invalid_colors = ['FF0000', '#GGGGGG', '#12345', 'red', '#1234567', '#12']
        invalid_colors.each do |color|
          tag = build(:tag, user: user, color: color)
          expect(tag).not_to be_valid
          expect(tag.errors[:color]).to include('must be a valid hex color')
        end
      end

      it 'allows blank color' do
        tag = build(:tag, user: user, color: '')
        expect(tag).to be_valid
      end

      it 'allows nil color' do
        tag = build(:tag, user: user, color: nil)
        expect(tag).to be_valid
      end
    end
  end

  describe 'callbacks' do
    let(:user) { create(:user) }

    describe '#normalize_name' do
      it 'strips whitespace and downcases name' do
        tag = create(:tag, name: '  Work Project  ', user: user)
        expect(tag.name).to eq('work project')
      end

      it 'handles nil name gracefully' do
        tag = build(:tag, name: nil, user: user)
        tag.valid?
        expect(tag.name).to be_nil
      end
    end

    describe '#normalize_color' do
      it 'upcases color' do
        tag = create(:tag, name: 'test', color: '#ff0000', user: user)
        expect(tag.color).to eq('#FF0000')
      end

      it 'handles nil color gracefully' do
        tag = create(:tag, name: 'test', color: nil, user: user)
        expect(tag.color).to be_nil
      end
    end
  end

  describe 'scopes' do
    describe '.ordered' do
      let(:user) { create(:user) }

      it 'returns tags ordered by name ascending' do
        tag_c = create(:tag, name: 'Charlie', user: user)
        tag_a = create(:tag, name: 'Alpha', user: user)
        tag_b = create(:tag, name: 'Bravo', user: user)

        expect(described_class.ordered).to eq([tag_a, tag_b, tag_c])
      end
    end
  end

  describe 'dependent destroy' do
    let(:user) { create(:user) }
    let(:tag) { create(:tag, user: user) }
    let(:todo) { create(:todo, user: user) }

    it 'destroys associated todo_tags when tag is destroyed' do
      create(:todo_tag, todo: todo, tag: tag)
      expect { tag.destroy }.to change(TodoTag, :count).by(-1)
    end

    it 'does not destroy associated todos when tag is destroyed' do
      create(:todo_tag, todo: todo, tag: tag)
      expect { tag.destroy }.not_to(change(Todo, :count))
    end
  end
end
