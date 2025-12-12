# frozen_string_literal: true

require 'rails_helper'

RSpec.describe 'Api::V1::TodoHistories', type: :request do
  let(:user) { create(:user) }
  let(:other_user) { create(:user) }
  let(:todo) { create(:todo, user: user) }
  let(:other_todo) { create(:todo, user: other_user) }

  let(:auth_headers) { auth_headers_for(user) }

  describe 'GET /api/v1/todos/:todo_id/histories' do
    context 'with valid todo_id' do
      before do
        create(:todo_history, :created, todo: todo, user: user, created_at: 3.days.ago)
        create(:todo_history, :status_changed, todo: todo, user: user, created_at: 2.days.ago)
        create(:todo_history, :priority_changed, todo: todo, user: user, created_at: 1.day.ago)
      end

      it 'returns todo histories in reverse chronological order' do
        get "/api/v1/todos/#{todo.id}/histories", headers: auth_headers

        expect(response).to have_http_status(:ok)
        json = response.parsed_body
        expect(json.length).to eq(3)
        # 新しい履歴が最初に来ることを確認
        expect(json.first['action']).to eq('priority_changed')
      end

      it 'includes user information' do
        get "/api/v1/todos/#{todo.id}/histories", headers: auth_headers

        json = response.parsed_body
        expect(json.first['user']).to be_present
        expect(json.first['user']['id']).to eq(user.id)
      end

      it 'includes human readable change description' do
        get "/api/v1/todos/#{todo.id}/histories", headers: auth_headers

        json = response.parsed_body
        expect(json.first['human_readable_change']).to be_present
      end
    end

    context "with other user's todo" do
      it 'returns 404' do
        get "/api/v1/todos/#{other_todo.id}/histories", headers: auth_headers

        expect(response).to have_http_status(:not_found)
      end
    end

    context 'without authentication' do
      it 'returns 401' do
        get "/api/v1/todos/#{todo.id}/histories"

        expect(response).to have_http_status(:unauthorized)
      end
    end
  end

  describe 'Todo history tracking' do
    it 'records history when todo is created' do
      todo_params = { todo: { title: 'New Todo' } }

      expect do
        # 学習ポイント：current_userを設定するためのテスト
        post '/api/v1/todos', params: todo_params.to_json, headers: auth_headers
      end.to change(TodoHistory, :count).by(1)

      history = TodoHistory.last
      expect(history.action).to eq('created')
      expect(history.field_name).to eq('created')
      expect(history.new_value).to eq('New Todo')
    end

    it 'records history when todo is updated' do
      patch_params = { todo: { title: 'Updated Title' } }

      expect do
        patch "/api/v1/todos/#{todo.id}", params: patch_params.to_json, headers: auth_headers
      end.to change { todo.todo_histories.count }.by(1)

      history = todo.todo_histories.last
      expect(history.action).to eq('updated')
      expect(history.field_name).to eq('title')
      expect(history.old_value).to eq(todo.title)
      expect(history.new_value).to eq('Updated Title')
    end

    it 'records status change with special action' do
      patch_params = { todo: { status: 'in_progress' } }

      patch "/api/v1/todos/#{todo.id}", params: patch_params.to_json, headers: auth_headers

      history = todo.todo_histories.last
      expect(history.action).to eq('status_changed')
      expect(history.field_name).to eq('status')
    end

    it 'records priority change with special action' do
      patch_params = { todo: { priority: 'high' } }

      patch "/api/v1/todos/#{todo.id}", params: patch_params.to_json, headers: auth_headers

      history = todo.todo_histories.last
      expect(history.action).to eq('priority_changed')
      expect(history.field_name).to eq('priority')
    end
  end
end
