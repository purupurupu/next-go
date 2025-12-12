# frozen_string_literal: true

require 'rails_helper'

RSpec.describe TodoSearchService do
  let(:user) { create(:user) }
  let(:other_user) { create(:user) }
  let(:category) { create(:category, user: user) }
  let(:tag1) { create(:tag, user: user) }
  let(:tag2) { create(:tag, user: user) }

  let!(:todo1) do
    create(:todo,
           user: user,
           title: 'Buy groceries',
           description: 'Milk, eggs, bread',
           status: 'pending',
           priority: 'high',
           category: category,
           due_date: 1.day.from_now).tap { |t| t.tags << tag1 }
  end

  let!(:todo2) do
    create(:todo,
           user: user,
           title: 'Complete project',
           description: 'Finish the Rails API',
           status: 'in_progress',
           priority: 'medium',
           due_date: 3.days.from_now).tap { |t| t.tags << [tag1, tag2] }
  end

  let!(:todo3) do
    create(:todo,
           user: user,
           title: 'Read documentation',
           description: 'Study Ruby on Rails guides',
           status: 'completed',
           priority: 'low',
           due_date: nil) # 過去の日付はバリデーションエラーになるため
  end

  let!(:other_user_todo) do
    create(:todo, user: other_user, title: 'Other user task')
  end

  describe '#call' do
    subject(:search_results) { described_class.new(user, params).call }

    context 'without any filters' do
      let(:params) { {} }

      it 'returns all todos for the user' do
        expect(search_results).to contain_exactly(todo1, todo2, todo3)
      end

      it 'does not return other users todos' do
        expect(search_results).not_to include(other_user_todo)
      end

      it 'includes associations to prevent N+1' do
        expect(search_results.includes_values).to include(:category, :tags, :comments, :user)
      end
    end

    describe 'full-text search' do
      context 'searching by title' do
        let(:params) { { q: 'project' } }

        it 'returns todos matching the title' do
          expect(search_results).to contain_exactly(todo2)
        end
      end

      context 'searching by description' do
        let(:params) { { query: 'rails' } }

        it 'returns todos matching the description' do
          expect(search_results).to contain_exactly(todo2, todo3)
        end
      end

      context 'case-insensitive search' do
        let(:params) { { search: 'GROCERIES' } }

        it 'matches regardless of case' do
          expect(search_results).to contain_exactly(todo1)
        end
      end

      context 'partial matching' do
        let(:params) { { q: 'doc' } }

        it 'matches partial strings' do
          expect(search_results).to contain_exactly(todo3)
        end
      end
    end

    describe 'category filtering' do
      context 'filtering by single category' do
        let(:params) { { category_id: category.id } }

        it 'returns todos with the specified category' do
          expect(search_results).to contain_exactly(todo1)
        end
      end

      context 'filtering by null category' do
        let(:params) { { category_id: -1 } }

        it 'returns todos without a category' do
          expect(search_results).to contain_exactly(todo2, todo3)
        end
      end

      context 'filtering by multiple categories including null' do
        let(:params) { { category_id: [category.id, -1] } }

        it 'returns todos with specified category or no category' do
          expect(search_results).to contain_exactly(todo1, todo2, todo3)
        end
      end
    end

    describe 'status filtering' do
      context 'filtering by single status' do
        let(:params) { { status: 'pending' } }

        it 'returns todos with the specified status' do
          expect(search_results).to contain_exactly(todo1)
        end
      end

      context 'filtering by multiple statuses' do
        let(:params) { { status: %w[pending in_progress] } }

        it 'returns todos with any of the specified statuses' do
          expect(search_results).to contain_exactly(todo1, todo2)
        end
      end

      context 'with invalid status' do
        let(:params) { { status: 'invalid_status' } }

        it 'ignores invalid values' do
          expect(search_results).to contain_exactly(todo1, todo2, todo3)
        end
      end
    end

    describe 'priority filtering' do
      context 'filtering by single priority' do
        let(:params) { { priority: 'high' } }

        it 'returns todos with the specified priority' do
          expect(search_results).to contain_exactly(todo1)
        end
      end

      context 'filtering by multiple priorities' do
        let(:params) { { priority: %w[low medium] } }

        it 'returns todos with any of the specified priorities' do
          expect(search_results).to contain_exactly(todo2, todo3)
        end
      end
    end

    describe 'tag filtering' do
      context 'filtering by single tag (ANY mode)' do
        let(:params) { { tag_ids: [tag2.id] } }

        it 'returns todos with the specified tag' do
          expect(search_results).to contain_exactly(todo2)
        end
      end

      context 'filtering by multiple tags (ANY mode)' do
        let(:params) { { tag_ids: [tag1.id, tag2.id], tag_mode: 'any' } }

        it 'returns todos with any of the specified tags' do
          expect(search_results).to contain_exactly(todo1, todo2)
        end
      end

      context 'filtering by multiple tags (ALL mode)' do
        let(:params) { { tag_ids: [tag1.id, tag2.id], tag_mode: 'all' } }

        it 'returns only todos with all specified tags' do
          expect(search_results).to contain_exactly(todo2)
        end
      end

      context 'with invalid tag IDs' do
        let(:params) { { tag_ids: [999_999] } }

        it 'returns no results for non-existent tags' do
          expect(search_results).to be_empty
        end
      end

      context 'with tags from other users' do
        let(:other_user_tag) { create(:tag, user: other_user) }
        let(:params) { { tag_ids: [other_user_tag.id] } }

        it 'does not return results for other users tags' do
          expect(search_results).to be_empty
        end
      end
    end

    describe 'date range filtering' do
      context 'filtering by due_date_from' do
        let(:params) { { due_date_from: Time.zone.today.to_s } }

        it 'returns todos with due date on or after the specified date' do
          expect(search_results).to contain_exactly(todo1, todo2)
        end
      end

      context 'filtering by due_date_to' do
        let(:params) { { due_date_to: 2.days.from_now.to_date.to_s } }

        it 'returns todos with due date on or before the specified date' do
          expect(search_results).to contain_exactly(todo1) # todo3 has nil due_date
        end
      end

      context 'filtering by date range' do
        let(:params) do
          {
            due_date_from: Time.zone.today.to_s,
            due_date_to: 2.days.from_now.to_date.to_s
          }
        end

        it 'returns todos within the date range' do
          expect(search_results).to contain_exactly(todo1)
        end
      end

      context 'with invalid date format' do
        let(:params) { { due_date_from: 'invalid-date' } }

        it 'ignores the filter and returns all todos' do
          expect(search_results).to contain_exactly(todo1, todo2, todo3)
        end
      end
    end

    describe 'combined filters' do
      context 'search with status and priority filters' do
        let(:params) do
          {
            q: 'project',
            status: 'in_progress',
            priority: 'medium'
          }
        end

        it 'applies all filters together' do
          expect(search_results).to contain_exactly(todo2)
        end
      end

      context 'complex filter combination' do
        let(:params) do
          {
            q: 'e',
            status: %w[pending in_progress],
            tag_ids: [tag1.id],
            due_date_from: Time.zone.today.to_s
          }
        end

        it 'returns todos matching all criteria' do
          expect(search_results).to contain_exactly(todo1, todo2)
        end
      end
    end

    describe 'sorting' do
      context 'default sorting' do
        let(:params) { {} }

        it 'sorts by position' do
          expect(search_results.first).to eq(todo1)
        end
      end

      context 'sorting by created_at DESC' do
        let(:params) { { sort_by: 'created_at', sort_order: 'desc' } }

        it 'sorts by creation date descending' do
          expect(search_results.first).to eq(todo3)
        end
      end

      context 'sorting by due_date ASC' do
        let(:params) { { sort_by: 'due_date', sort_order: 'asc' } }

        it 'sorts by due date ascending with nulls last' do
          # todo1: 1.day.from_now, todo2: 3.days.from_now, todo3: nil
          # Expected order: todo1, todo2, todo3 (nulls last)
          results = search_results.to_a
          expect(results[0]).to eq(todo1)
          expect(results[1]).to eq(todo2)
          expect(results[2]).to eq(todo3)
        end
      end

      context 'sorting by title' do
        let(:params) { { sort_by: 'title', sort_order: 'asc' } }

        it 'sorts alphabetically by title' do
          expect(search_results.map(&:title)).to eq(['Buy groceries', 'Complete project', 'Read documentation'])
        end
      end

      context 'sorting by priority DESC' do
        let(:params) { { sort_by: 'priority', sort_order: 'desc' } }

        it 'sorts by priority with high first' do
          expect(search_results.first.priority).to eq('high')
          expect(search_results.last.priority).to eq('low')
        end
      end
    end
  end
end
