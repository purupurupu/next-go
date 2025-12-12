# frozen_string_literal: true

require 'rails_helper'

RSpec.describe TodoHistory, type: :model do
  describe 'associations' do
    it { is_expected.to belong_to(:todo) }
    it { is_expected.to belong_to(:user) }
  end

  describe 'validations' do
    it { is_expected.to validate_presence_of(:field_name) }
    it { is_expected.to validate_presence_of(:action) }
  end

  describe 'enums' do
    it 'defines action enum with correct values' do
      expect(described_class.actions).to eq(
        'created' => 0,
        'updated' => 1,
        'deleted' => 2,
        'status_changed' => 3,
        'priority_changed' => 4
      )
    end
  end

  describe 'default scope' do
    let!(:old_history) { create(:todo_history, created_at: 2.days.ago) }
    let!(:new_history) { create(:todo_history, created_at: 1.day.ago) }

    it 'orders by created_at descending' do
      expect(described_class.all).to eq([new_history, old_history])
    end
  end

  describe 'scopes' do
    let(:user) { create(:user) }
    let(:todo) { create(:todo, user: user) }
    let!(:title_history) { create(:todo_history, todo: todo, user: user, field_name: 'title') }
    let!(:status_history) { create(:todo_history, todo: todo, user: user, field_name: 'status') }

    describe '.recent' do
      before do
        create_list(:todo_history, 15, todo: todo)
      end

      it 'returns limited number of records' do
        expect(described_class.recent(5).count).to eq(5)
      end

      it 'defaults to 10 records' do
        expect(described_class.recent.count).to eq(10)
      end
    end

    describe '.for_field' do
      it 'returns histories for specific field' do
        result = described_class.for_field('title')
        expect(result).to include(title_history)
        expect(result).not_to include(status_history)
      end
    end

    describe '.by_user' do
      let(:other_user) { create(:user) }
      let!(:other_history) { create(:todo_history, user: other_user) }

      it 'returns histories by specific user' do
        result = described_class.by_user(user)
        expect(result).to include(title_history, status_history)
        expect(result).not_to include(other_history)
      end
    end
  end

  describe '#human_readable_change' do
    context 'with title change' do
      let(:history) { build(:todo_history, field_name: 'title', old_value: 'Old', new_value: 'New') }

      it 'returns formatted message' do
        expect(history.human_readable_change).to eq('タイトルを「Old」から「New」に変更')
      end
    end

    context 'with status change' do
      let(:history) { build(:todo_history, :status_changed) }

      it 'returns translated status message' do
        expect(history.human_readable_change).to eq('ステータスを「未着手」から「進行中」に変更')
      end
    end

    context 'with priority change' do
      let(:history) { build(:todo_history, :priority_changed) }

      it 'returns translated priority message' do
        expect(history.human_readable_change).to eq('優先度を「低」から「高」に変更')
      end
    end

    context 'with completed change' do
      let(:history) { build(:todo_history, :completed) }

      it 'returns completion message' do
        expect(history.human_readable_change).to eq('タスクを完了にマーク')
      end
    end

    context 'with due_date change' do
      let(:history) { build(:todo_history, field_name: 'due_date', old_value: '2024-01-01', new_value: '2024-01-15') }

      it 'returns formatted date message' do
        expect(history.human_readable_change).to eq('期限日を「2024年01月01日」から「2024年01月15日」に変更')
      end
    end
  end

  describe '.grouped_by_timestamp' do
    let(:todo) { create(:todo) }
    let(:timestamp) { Time.current }
    let!(:history1) { create(:todo_history, todo: todo, created_at: timestamp) }
    let!(:history2) { create(:todo_history, todo: todo, created_at: timestamp) }
    let!(:history3) { create(:todo_history, todo: todo, created_at: timestamp + 1.minute) }

    it 'groups histories by timestamp' do
      grouped = described_class.grouped_by_timestamp
      expect(grouped.keys.count).to eq(2)
      expect(grouped[timestamp.to_i]).to contain_exactly(history1, history2)
      expect(grouped[(timestamp + 1.minute).to_i]).to eq([history3])
    end
  end
end
