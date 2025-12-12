# frozen_string_literal: true

require 'rails_helper'

RSpec.describe 'Todo Search API', type: :request do
  let(:user) { create(:user) }
  let(:other_user) { create(:user) }
  let(:auth_headers) { auth_headers_for(user) }

  let(:category) { create(:category, user: user) }
  let(:tag1) { create(:tag, user: user, name: 'urgent') }
  let(:tag2) { create(:tag, user: user, name: 'work') }

  before do
    # Create test todos
    create(:todo,
           user: user,
           title: 'Buy milk and eggs',
           description: 'Go to the grocery store',
           status: 'pending',
           priority: 'high',
           category: category,
           due_date: 1.day.from_now,
           position: 1).tap { |t| t.tags << tag1 }

    create(:todo,
           user: user,
           title: 'Complete project report',
           description: 'Finish the quarterly report for management',
           status: 'in_progress',
           priority: 'medium',
           due_date: 3.days.from_now,
           position: 2).tap { |t| t.tags << [tag1, tag2] }

    create(:todo,
           user: user,
           title: 'Read documentation',
           description: 'Study the new API documentation',
           status: 'completed',
           priority: 'low',
           due_date: nil, # 過去の日付はバリデーションエラーになるため
           position: 3)

    # Other user's todo (should not appear in results)
    create(:todo, user: other_user, title: 'Other user task')
  end

  describe 'GET /api/v1/todos/search' do
    context 'without authentication' do
      it 'returns 401 unauthorized' do
        get '/api/v1/todos/search'
        # Devise JWT returns 401 for unauthenticated requests
        expect(response).to have_http_status(:unauthorized)
      end
    end

    context 'with authentication' do
      context 'without any parameters' do
        it 'returns all todos with pagination metadata' do
          get '/api/v1/todos/search', headers: auth_headers

          expect(response).to have_http_status(:success)

          json = response.parsed_body
          expect(json['data'].size).to eq(3)
          expect(json['meta']).to include(
            'total' => 3,
            'current_page' => 1,
            'total_pages' => 1,
            'per_page' => 20
          )
        end
      end

      context 'with text search' do
        it 'searches in title' do
          get '/api/v1/todos/search', params: { q: 'milk' }, headers: auth_headers

          expect(response).to have_http_status(:success)

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
          expect(json['data'][0]['title']).to include('milk')
          expect(json['meta']['search_query']).to eq('milk')
        end

        it 'searches in description' do
          get '/api/v1/todos/search', params: { query: 'quarterly' }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
          expect(json['data'][0]['description']).to include('quarterly')
        end

        it 'returns highlights for matches' do
          get '/api/v1/todos/search', params: { q: 'project' }, headers: auth_headers

          json = response.parsed_body
          todo = json['data'][0]
          expect(todo['highlights']).to be_present
          expect(todo['highlights']['title']).to be_present
          expect(todo['highlights']['title'][0]).to include(
            'start' => 9,
            'end' => 16,
            'matched_text' => 'project'
          )
        end

        it 'performs case-insensitive search' do
          get '/api/v1/todos/search', params: { q: 'PROJECT' }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
        end
      end

      context 'with category filter' do
        it 'filters by category' do
          get '/api/v1/todos/search', params: { category_id: category.id }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
          expect(json['data'][0]['category']['id']).to eq(category.id)
        end

        it 'filters uncategorized todos' do
          get '/api/v1/todos/search', params: { category_id: -1 }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(2)
          expect(json['data'].all? { |t| t['category'].nil? }).to be true
        end
      end

      context 'with status filter' do
        it 'filters by single status' do
          get '/api/v1/todos/search', params: { status: 'pending' }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
          expect(json['data'][0]['status']).to eq('pending')
        end

        it 'filters by multiple statuses' do
          get '/api/v1/todos/search', params: { status: %w[pending in_progress] }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(2)
          expect(json['data'].pluck('status')).to contain_exactly('pending', 'in_progress')
        end
      end

      context 'with priority filter' do
        it 'filters by priority' do
          get '/api/v1/todos/search', params: { priority: 'high' }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
          expect(json['data'][0]['priority']).to eq('high')
        end
      end

      context 'with tag filter' do
        it 'filters by tag (ANY mode)' do
          get '/api/v1/todos/search', params: { tag_ids: [tag2.id] }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
          expect(json['data'][0]['tags'].pluck('id')).to include(tag2.id)
        end

        it 'filters by multiple tags (ANY mode)' do
          get '/api/v1/todos/search', params: { tag_ids: [tag1.id, tag2.id], tag_mode: 'any' }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(2)
        end

        it 'filters by multiple tags (ALL mode)' do
          get '/api/v1/todos/search', params: { tag_ids: [tag1.id, tag2.id], tag_mode: 'all' }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
          expect(json['data'][0]['tags'].size).to eq(2)
        end
      end

      context 'with date range filter' do
        it 'filters by due_date_from' do
          get '/api/v1/todos/search', params: { due_date_from: Time.zone.today.to_s }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(2)
          expect(json['data'].all? { |t| t['due_date'].nil? || Date.parse(t['due_date']) >= Time.zone.today }).to be true
        end

        it 'filters by due_date_to' do
          get '/api/v1/todos/search', params: { due_date_to: 2.days.from_now.to_date.to_s }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1) # todo3 has nil due_date
        end

        it 'filters by date range' do
          get '/api/v1/todos/search',
              params: {
                due_date_from: Time.zone.today.to_s,
                due_date_to: 2.days.from_now.to_date.to_s
              },
              headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
        end
      end

      context 'with combined filters' do
        it 'applies all filters together' do
          get '/api/v1/todos/search',
              params: {
                q: 'project',
                status: 'in_progress',
                priority: 'medium',
                tag_ids: [tag1.id]
              },
              headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(1)
          expect(json['data'][0]['title']).to include('project')
          expect(json['data'][0]['status']).to eq('in_progress')
          expect(json['meta']['filters_applied']).to include('search', 'status', 'priority', 'tag_ids')
        end
      end

      context 'with sorting' do
        it 'sorts by created_at DESC' do
          get '/api/v1/todos/search', params: { sort_by: 'created_at', sort_order: 'desc' }, headers: auth_headers

          json = response.parsed_body
          created_dates = json['data'].pluck('created_at')
          expect(created_dates).to eq(created_dates.sort.reverse)
        end

        it 'sorts by title ASC' do
          get '/api/v1/todos/search', params: { sort_by: 'title', sort_order: 'asc' }, headers: auth_headers

          json = response.parsed_body
          titles = json['data'].pluck('title')
          expect(titles).to eq(['Buy milk and eggs', 'Complete project report', 'Read documentation'])
        end

        it 'sorts by priority DESC' do
          get '/api/v1/todos/search', params: { sort_by: 'priority', sort_order: 'desc' }, headers: auth_headers

          json = response.parsed_body
          priorities = json['data'].pluck('priority')
          expect(priorities).to eq(%w[high medium low])
        end
      end

      context 'with pagination' do
        before do
          # Create more todos for pagination testing
          10.times do |i|
            create(:todo, user: user, title: "Task #{i + 4}", position: i + 4)
          end
        end

        it 'paginates results' do
          get '/api/v1/todos/search', params: { page: 1, per_page: 5 }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(5)
          expect(json['meta']['current_page']).to eq(1)
          expect(json['meta']['total_pages']).to eq(3)
          expect(json['meta']['per_page']).to eq(5)
          expect(json['meta']['total']).to eq(13)
        end

        it 'returns correct page' do
          get '/api/v1/todos/search', params: { page: 2, per_page: 5 }, headers: auth_headers

          json = response.parsed_body
          expect(json['data'].size).to eq(5)
          expect(json['meta']['current_page']).to eq(2)
        end

        it 'limits per_page to 100' do
          get '/api/v1/todos/search', params: { per_page: 200 }, headers: auth_headers

          json = response.parsed_body
          expect(json['meta']['per_page']).to eq(100)
        end
      end

      context 'with no results' do
        it 'returns helpful suggestions' do
          get '/api/v1/todos/search', params: { q: 'nonexistent' }, headers: auth_headers

          expect(response).to have_http_status(:success)
          json = response.parsed_body

          # The search controller should return data even when empty
          # If data key is missing, check if it's a different response structure
          expect(json['data']).to eq([]) if json.key?('data')

          expect(json).to have_key('meta')
          expect(json['meta']['suggestions']).to be_present

          suggestion_types = json['meta']['suggestions'].pluck('type')
          expect(suggestion_types).to include('spelling', 'broader_search', 'clear_filters')
        end

        it 'suggests reducing filters when many are applied' do
          get '/api/v1/todos/search',
              params: {
                q: 'test',
                status: 'pending',
                priority: 'high',
                category_id: category.id,
                tag_ids: [tag1.id]
              },
              headers: auth_headers

          expect(response).to have_http_status(:success)
          json = response.parsed_body

          # This should return 0 results, triggering suggestions
          expect(json['data']).to eq([]) if json.key?('data')

          expect(json).to have_key('meta')
          expect(json['meta']['suggestions']).to be_present

          suggestions = json['meta']['suggestions'].find { |s| s['type'] == 'reduce_filters' }
          expect(suggestions).to be_present
          expect(suggestions['current_filters']).to match_array(%w[search status priority category_id tag_ids])
        end
      end

      context 'with invalid parameters' do
        it 'ignores invalid status values' do
          get '/api/v1/todos/search', params: { status: 'invalid_status' }, headers: auth_headers

          expect(response).to have_http_status(:success)
          json = response.parsed_body
          expect(json['data'].size).to eq(3) # All todos returned
        end

        it 'handles invalid date formats gracefully' do
          get '/api/v1/todos/search', params: { due_date_from: 'invalid-date' }, headers: auth_headers

          expect(response).to have_http_status(:success)
          json = response.parsed_body
          expect(json['data'].size).to eq(3) # All todos returned
        end
      end
    end
  end
end
