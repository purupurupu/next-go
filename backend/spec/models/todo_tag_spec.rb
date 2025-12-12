require 'rails_helper'

RSpec.describe TodoTag, type: :model do
  describe 'associations' do
    it { is_expected.to belong_to(:todo) }
    it { is_expected.to belong_to(:tag) }
  end

  describe 'validations' do
    subject { build(:todo_tag, todo: todo, tag: tag) }

    let(:user) { create(:user) }
    let(:todo) { create(:todo, user: user) }
    let(:tag) { create(:tag, user: user) }

    it { is_expected.to validate_uniqueness_of(:todo_id).scoped_to(:tag_id) }

    it 'prevents duplicate todo-tag combinations' do
      create(:todo_tag, todo: todo, tag: tag)
      duplicate = build(:todo_tag, todo: todo, tag: tag)
      expect(duplicate).not_to be_valid
      expect(duplicate.errors[:todo_id]).to include('has already been taken')
    end

    it 'allows same tag on different todos' do
      another_todo = create(:todo, user: user)
      create(:todo_tag, todo: todo, tag: tag)
      todo_tag_for_another_todo = build(:todo_tag, todo: another_todo, tag: tag)
      expect(todo_tag_for_another_todo).to be_valid
    end

    it 'allows same todo with different tags' do
      another_tag = create(:tag, user: user)
      create(:todo_tag, todo: todo, tag: tag)
      todo_tag_with_another_tag = build(:todo_tag, todo: todo, tag: another_tag)
      expect(todo_tag_with_another_tag).to be_valid
    end
  end

  describe 'database constraints' do
    let(:user) { create(:user) }
    let(:todo) { create(:todo, user: user) }
    let(:tag) { create(:tag, user: user) }

    it 'requires a todo' do
      todo_tag = build(:todo_tag, todo: nil, tag: tag)
      expect(todo_tag).not_to be_valid
    end

    it 'requires a tag' do
      todo_tag = build(:todo_tag, todo: todo, tag: nil)
      expect(todo_tag).not_to be_valid
    end
  end
end
