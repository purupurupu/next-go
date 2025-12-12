require 'rails_helper'

RSpec.describe Todo, type: :model do
  describe 'associations' do
    it { is_expected.to belong_to(:user) }
    it { is_expected.to belong_to(:category).optional }
  end

  describe 'validations' do
    it 'is valid with valid attributes' do
      todo = build(:todo)
      expect(todo).to be_valid
    end

    it 'is not valid without a title' do
      todo = build(:todo, title: nil)
      expect(todo).not_to be_valid
    end

    it 'is not valid with a past due date' do
      todo = build(:todo, due_date: 1.day.ago)
      expect(todo).not_to be_valid
    end

    it 'is not valid without a user' do
      todo = build(:todo, user: nil)
      expect(todo).not_to be_valid
    end
  end

  describe 'enums' do
    describe 'priority' do
      it 'defines priority values correctly' do
        expect(described_class.priorities).to eq({ 'low' => 0, 'medium' => 1, 'high' => 2 })
      end

      it 'sets default priority to medium' do
        todo = described_class.new(title: 'Test', user: create(:user))
        todo.save!
        expect(todo.priority).to eq('medium')
      end

      it 'allows valid priority values' do
        %w[low medium high].each do |priority|
          todo = build(:todo, priority: priority)
          expect(todo).to be_valid
        end
      end
    end

    describe 'status' do
      it 'defines status values correctly' do
        expect(described_class.statuses).to eq({ 'pending' => 0, 'in_progress' => 1, 'completed' => 2 })
      end

      it 'sets default status to pending' do
        todo = described_class.new(title: 'Test', user: create(:user))
        todo.save!
        expect(todo.status).to eq('pending')
      end

      it 'allows valid status values' do
        %w[pending in_progress completed].each do |status|
          todo = build(:todo, status: status)
          expect(todo).to be_valid
        end
      end
    end
  end

  describe 'description' do
    it 'allows nil description' do
      todo = build(:todo, description: nil)
      expect(todo).to be_valid
    end

    it 'allows empty string description' do
      todo = build(:todo, description: '')
      expect(todo).to be_valid
    end

    it 'allows text description' do
      todo = build(:todo, description: 'This is a detailed description of the todo item.')
      expect(todo).to be_valid
      expect(todo.description).to eq('This is a detailed description of the todo item.')
    end
  end

  describe 'factory' do
    it 'creates a valid todo' do
      todo = create(:todo)
      expect(todo).to be_persisted
      expect(todo.title).to be_present
      expect(todo.user).to be_present
      expect(todo.priority).to be_present
      expect(todo.status).to be_present
    end
  end
end
