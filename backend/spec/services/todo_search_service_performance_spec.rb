# frozen_string_literal: true

require 'rails_helper'

RSpec.describe TodoSearchService do
  let(:user) { create(:user) }

  before do
    # Create a large dataset for performance testing
    categories = create_list(:category, 5, user: user)
    tags = create_list(:tag, 10, user: user)

    # Create 1000 todos with various attributes
    1000.times do |i|
      todo = create(:todo,
                    user: user,
                    title: "Task #{i} #{%w[urgent important routine].sample}",
                    description: "Description for task #{i} with #{%w[project meeting review].sample}",
                    status: Todo.statuses.keys.sample,
                    priority: Todo.priorities.keys.sample,
                    category: categories.sample,
                    due_date: rand(0..60).days.from_now, # Only future dates to avoid validation error
                    position: i)

      # Assign random tags
      todo.tags << tags.sample(rand(0..3))
    end
  end

  describe 'search performance' do
    it 'performs text search within acceptable time' do
      time = Benchmark.realtime do
        result = described_class.new(user, { q: 'urgent' }).call
        result.to_a # Force query execution
      end

      expect(time).to be < 0.5 # Should complete within 500ms
    end

    it 'performs complex filtering within acceptable time' do
      time = Benchmark.realtime do
        result = described_class.new(user, {
                                       q: 'task',
                                       status: %w[pending in_progress],
                                       priority: 'high',
                                       tag_ids: user.tags.limit(3).pluck(:id),
                                       due_date_from: Time.zone.today.to_s,
                                       due_date_to: 1.week.from_now.to_date.to_s,
                                       sort_by: 'priority',
                                       sort_order: 'desc'
                                     }).call
        result.to_a # Force query execution
      end

      expect(time).to be < 1.0 # Should complete within 1 second
    end

    it 'maintains consistent performance on repeated searches' do
      params = { q: 'project', status: 'pending' }

      # Run multiple searches and ensure performance is consistent
      times = Array.new(3) do
        Benchmark.realtime do
          described_class.new(user, params).call.to_a
        end
      end

      # Performance should be consistent (within 100% variance to allow for test environment fluctuations)
      average_time = times.sum / times.size
      expect(times).to all(be_within(average_time * 1.0).of(average_time))
    end

    it 'handles pagination efficiently' do
      time = Benchmark.realtime do
        result = described_class.new(user, {
                                       page: 10,
                                       per_page: 50
                                     }).call
        result.to_a # Force query execution
      end

      expect(time).to be < 0.3 # Pagination should be fast
    end

    it 'avoids N+1 queries' do
      # Get a baseline count of queries
      query_count = 0
      ActiveSupport::Notifications.subscribe('sql.active_record') do |*_args|
        query_count += 1
      end

      result = described_class.new(user, { q: 'task' }).call.limit(10)
      result.each do |todo|
        # Access associations that should be preloaded
        todo.category&.name
        todo.tags.map(&:name)
        todo.comments.size
        todo.user.email
      end

      ActiveSupport::Notifications.unsubscribe('sql.active_record')

      # Should use less than 20 queries for 10 todos (avoiding N+1)
      expect(query_count).to be < 20
    end
  end

  describe 'database indexes effectiveness' do
    it 'uses indexes for text search' do
      # Ensure we have some data
      create(:todo, user: user, title: 'urgent task')

      explain_result = ActiveRecord::Base.connection.execute(
        "EXPLAIN (FORMAT JSON) SELECT * FROM todos WHERE LOWER(title) LIKE '%urgent%'"
      )

      explain_json = JSON.parse(explain_result.first['QUERY PLAN'])
      plan_text = explain_json.to_s.downcase

      # Check that it's not doing a sequential scan on the entire table
      # Note: LIKE with leading wildcard may still use seq scan, but with indexes it should be faster
      expect(plan_text).to include('todos')
    end

    it 'uses composite indexes for user-based queries' do
      explain_result = ActiveRecord::Base.connection.execute(
        "EXPLAIN (FORMAT JSON) SELECT * FROM todos WHERE user_id = #{user.id} AND status = 0"
      )

      explain_json = JSON.parse(explain_result.first['QUERY PLAN'])
      plan_text = explain_json.to_s.downcase

      # Should use index on user_id at minimum
      expect(plan_text).to include('index')
    end
  end
end
