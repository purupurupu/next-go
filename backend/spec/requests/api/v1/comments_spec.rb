# frozen_string_literal: true

require 'rails_helper'

RSpec.describe 'Api::V1::Comments', type: :request do
  # 学習ポイント：認証を含むAPIテストのセットアップ
  let(:user) { create(:user) }
  let(:other_user) { create(:user) }
  let(:todo) { create(:todo, user: user) }
  let(:other_todo) { create(:todo, user: other_user) }

  # 学習ポイント：認証ヘルパーの使用
  # リクエストスペックではauth_headers_forを使って認証
  let(:auth_headers) { auth_headers_for(user) }
  let(:other_auth_headers) { auth_headers_for(other_user) }

  describe 'GET /api/v1/todos/:todo_id/comments' do
    let!(:old_comment) { create(:comment, commentable: todo, created_at: 2.days.ago) }

    context 'with valid todo_id' do
      before do
        create_list(:comment, 3, commentable: todo)
      end

      it 'returns comments in chronological order' do
        get "/api/v1/todos/#{todo.id}/comments", headers: auth_headers

        expect(response).to have_http_status(:ok)
        json = response.parsed_body
        expect(json['data']).to be_an(Array)
        expect(json['data'].length).to eq(4)
        # 古いコメントが最初に来ることを確認
        expect(json['data'].first['id']).to eq(old_comment.id)
      end

      it 'includes user information' do
        get "/api/v1/todos/#{todo.id}/comments", headers: auth_headers

        json = response.parsed_body
        expect(json['data']).to be_an(Array)
        expect(json['data'].first['user']).to be_present
        expect(json['data'].first['user']['id']).to be_present
        expect(json['data'].first['user']['name']).to be_present
      end

      it 'includes editable flag' do
        get "/api/v1/todos/#{todo.id}/comments", headers: auth_headers

        json = response.parsed_body
        expect(json['data']).to be_an(Array)
        expect(json['data'].first).to have_key('editable')
      end
    end

    context "with other user's todo" do
      it 'returns 404' do
        get "/api/v1/todos/#{other_todo.id}/comments", headers: auth_headers

        expect(response).to have_http_status(:not_found)
      end
    end
  end

  describe 'POST /api/v1/todos/:todo_id/comments' do
    let(:valid_params) { { comment: { content: 'This is a test comment' } } }
    let(:invalid_params) { { comment: { content: '' } } }

    context 'with valid params' do
      it 'creates a new comment' do
        expect do
          post "/api/v1/todos/#{todo.id}/comments", params: valid_params.to_json, headers: auth_headers
        end.to change { todo.comments.count }.by(1)

        expect(response).to have_http_status(:created)
        json = response.parsed_body
        expect(json['data']).to be_present
        expect(json['data']['content']).to eq('This is a test comment')
        expect(json['data']['user']['id']).to eq(user.id)
      end
    end

    context 'with invalid params' do
      it 'returns unprocessable entity' do
        post "/api/v1/todos/#{todo.id}/comments", params: invalid_params.to_json, headers: auth_headers

        expect(response).to have_http_status(:unprocessable_content)
        json = response.parsed_body
        expect(json['error']).to be_present
        expect(json['error']['code']).to eq('VALIDATION_FAILED')
      end
    end

    context "with other user's todo" do
      it 'returns 404' do
        post "/api/v1/todos/#{other_todo.id}/comments", params: valid_params.to_json, headers: auth_headers

        expect(response).to have_http_status(:not_found)
      end
    end
  end

  describe 'PATCH /api/v1/todos/:todo_id/comments/:id' do
    let(:comment) { create(:comment, commentable: todo, user: user) }
    let(:other_comment) { create(:comment, commentable: todo, user: other_user) }
    let(:update_params) { { comment: { content: 'Updated comment' } } }

    context 'when user owns the comment' do
      context 'within editable time' do
        it 'updates the comment' do
          patch "/api/v1/todos/#{todo.id}/comments/#{comment.id}", params: update_params.to_json, headers: auth_headers

          expect(response).to have_http_status(:ok)
          json = response.parsed_body
          expect(json['data']).to be_present
          expect(json['data']['content']).to eq('Updated comment')
        end
      end

      context 'after editable time' do
        before { comment.update!(created_at: 20.minutes.ago) }

        it 'returns unprocessable entity' do
          patch "/api/v1/todos/#{todo.id}/comments/#{comment.id}", params: update_params.to_json, headers: auth_headers

          expect(response).to have_http_status(:unprocessable_content)
          json = response.parsed_body
          expect(json['error']['message']).to include('編集可能時間')
        end
      end
    end

    context 'when user does not own the comment' do
      it 'returns forbidden' do
        patch "/api/v1/todos/#{todo.id}/comments/#{other_comment.id}", params: update_params.to_json, headers: auth_headers

        expect(response).to have_http_status(:forbidden)
        json = response.parsed_body
        expect(json['error']['message']).to include('編集権限')
      end
    end
  end

  describe 'DELETE /api/v1/todos/:todo_id/comments/:id' do
    let!(:comment) { create(:comment, commentable: todo, user: user) }
    let!(:other_comment) { create(:comment, commentable: todo, user: other_user) }

    context 'when user owns the comment' do
      it 'soft deletes the comment' do
        expect do
          delete "/api/v1/todos/#{todo.id}/comments/#{comment.id}", headers: auth_headers
        end.not_to(change { Comment.unscoped.count })

        expect(response).to have_http_status(:no_content)
        expect(comment.reload.deleted?).to be true
      end
    end

    context 'when user does not own the comment' do
      it 'returns forbidden' do
        delete "/api/v1/todos/#{todo.id}/comments/#{other_comment.id}", headers: auth_headers

        expect(response).to have_http_status(:forbidden)
        json = response.parsed_body
        expect(json['error']['message']).to include('削除権限')
      end
    end
  end

  describe 'authentication requirements' do
    # 学習ポイント：認証なしでのアクセステスト
    # ヘッダーを送らないことで未認証状態をテスト

    it 'requires authentication for all endpoints' do
      # 学習ポイント：Devise JWTの挙動により、認証なしのアクセスは401(unauthorized)を返す
      get "/api/v1/todos/#{todo.id}/comments"
      expect(response).to have_http_status(:unauthorized)

      post "/api/v1/todos/#{todo.id}/comments", params: { comment: { content: 'Test' } }.to_json
      expect(response).to have_http_status(:unauthorized)

      patch "/api/v1/todos/#{todo.id}/comments/1", params: { comment: { content: 'Test' } }.to_json
      expect(response).to have_http_status(:unauthorized)

      delete "/api/v1/todos/#{todo.id}/comments/1"
      expect(response).to have_http_status(:unauthorized)
    end
  end
end
