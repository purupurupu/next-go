# frozen_string_literal: true

require 'rails_helper'

RSpec.describe 'Error Handling', type: :request do
  let(:user) { create(:user) }
  let(:auth_headers) { auth_headers_for(user) }

  describe 'Request ID tracking' do
    it 'includes request ID in response headers' do
      get '/api/v1/todos', headers: auth_headers

      expect(response.headers['X-Request-Id']).to be_present
    end

    it 'includes request ID in error responses' do
      get '/api/v1/todos/999999', headers: auth_headers

      json = response.parsed_body
      expect(json['error']['request_id']).to be_present
      expect(json['error']['request_id']).to eq(response.headers['X-Request-Id'])
    end
  end

  describe 'Standardized error responses' do
    context 'with not found errors' do
      it 'returns proper error format for missing resources' do
        get '/api/v1/todos/999999', headers: auth_headers

        expect(response).to have_http_status(:not_found)

        json = response.parsed_body
        expect(json['error']['code']).to eq('ERROR')
        expect(json['error']['message']).to include('not found')
        expect(json['error']['request_id']).to be_present
      end
    end

    context 'with validation errors' do
      it 'returns proper error format for validation failures' do
        post '/api/v1/todos',
             params: { todo: { title: '' } },
             headers: auth_headers,
             as: :json

        expect(response).to have_http_status(:unprocessable_content)

        json = response.parsed_body
        # Expect standardized error format
        expect(json).to have_key('error')
        expect(json['error']).to include(
          'code' => 'VALIDATION_FAILED',
          'message' => 'Validation failed. Please check your input.'
        )
        expect(json['error']['details']['validation_errors']['title']).to include("can't be blank")
      end
    end

    context 'with authentication errors' do
      it 'returns proper error format for missing authentication' do
        get '/api/v1/todos'

        expect(response).to have_http_status(:unauthorized)
      end

      it 'returns proper error format for invalid credentials' do
        post '/auth/sign_in',
             params: { user: { email: 'wrong@example.com', password: 'wrong' } },
             as: :json

        expect(response).to have_http_status(:unauthorized)

        json = response.parsed_body
        expect(json['error']).to eq('Invalid Email or password.')
      end
    end

    context 'with parameter errors' do
      it 'returns proper error format for missing parameters' do
        post '/api/v1/todos',
             params: {},
             headers: auth_headers,
             as: :json

        expect(response).to have_http_status(:bad_request)

        json = response.parsed_body
        expect(json['error']['code']).to eq('ERROR')
        expect(json['error']['message']).to eq('Required parameter missing: todo')
      end
    end
  end

  describe 'Custom error classes' do
    it 'handles business logic errors properly' do
      # Create a todo with maximum tags (assuming there's a limit)
      todo = create(:todo, user: user)

      # Try to add invalid tags
      patch "/api/v1/todos/#{todo.id}/tags",
            params: { tag_ids: [999_999] },
            headers: auth_headers,
            as: :json

      expect(response).to have_http_status(:unprocessable_content)

      json = response.parsed_body
      expect(json['error']['code']).to eq('ERROR')
      expect(json['error']['message']).to eq('Invalid tag IDs')
      expect(json['error']['details']['invalid_tags']).to eq([999_999])
    end
  end
end
