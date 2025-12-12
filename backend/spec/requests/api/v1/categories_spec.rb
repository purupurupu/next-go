require 'rails_helper'

RSpec.describe '/api/v1/categories', type: :request do
  include AuthenticationHelpers

  let(:user) { create(:user) }
  let(:other_user) { create(:user, email: 'other@example.com') }
  let(:headers) { auth_headers_for(user) }

  describe 'GET /api/v1/categories' do
    let!(:user_categories) { create_list(:category, 3, user: user) }

    before do
      create(:category, user: other_user) # Other user's category
    end

    it 'returns user categories in alphabetical order' do
      get '/api/v1/categories', headers: headers

      expect(response).to have_http_status(:ok)
      json_response = response.parsed_body

      expect(json_response['data'].size).to eq(3)
      expect(json_response['data'].pluck('name')).to eq(user_categories.map(&:name).sort)
    end

    it 'includes todo_count in response' do
      category = user_categories.first
      create_list(:todo, 2, user: user, category: category)

      get '/api/v1/categories', headers: headers

      expect(response).to have_http_status(:ok)
      json_response = response.parsed_body

      category_response = json_response['data'].find { |c| c['id'] == category.id }
      expect(category_response['todo_count']).to eq(2)
    end

    it 'includes proper serialized data with timestamps' do
      get '/api/v1/categories', headers: headers

      json_response = response.parsed_body
      expect(json_response['data'].first).to include(
        'id' => be_a(Integer),
        'name' => be_a(String),
        'color' => be_a(String),
        'todo_count' => be_a(Integer),
        'created_at' => be_a(String),
        'updated_at' => be_a(String)
      )
    end

    context 'with response headers' do
      before { get '/api/v1/categories', headers: headers }

      it_behaves_like 'standard success response'
    end

    context 'without authentication' do
      it 'returns unauthorized' do
        get '/api/v1/categories'
        expect(response).to have_http_status(:unauthorized)
      end
    end
  end

  describe 'GET /api/v1/categories/:id' do
    let(:category) { create(:category, user: user) }

    context 'with valid category' do
      before { get "/api/v1/categories/#{category.id}", headers: headers }

      it 'returns the category' do
        expect(response).to have_http_status(:ok)
        json_response = response.parsed_body

        expect(json_response['data']['id']).to eq(category.id)
        expect(json_response['data']['name']).to eq(category.name)
        expect(json_response['data']['color']).to eq(category.color)
      end

      it_behaves_like 'standard success response'
    end

    context 'when category does not exist' do
      before { get '/api/v1/categories/999', headers: headers }

      it 'returns not found' do
        expect(response).to have_http_status(:not_found)
      end

      it_behaves_like 'standard error response', 'ERROR'
    end

    context 'when category belongs to other user' do
      let(:other_category) { create(:category, user: other_user) }

      before { get "/api/v1/categories/#{other_category.id}", headers: headers }

      it 'returns not found' do
        expect(response).to have_http_status(:not_found)

        json_response = response.parsed_body
        expect(json_response['error']['message']).to eq('Category not found')
      end
    end
  end

  describe 'POST /api/v1/categories' do
    let(:valid_params) do
      {
        category: {
          name: 'New Category',
          color: '#FF5733'
        }
      }
    end

    context 'with valid params' do
      it 'creates a new category' do
        expect do
          post '/api/v1/categories', params: valid_params, headers: headers, as: :json
        end.to change(Category, :count).by(1)

        expect(response).to have_http_status(:created)
        json_response = response.parsed_body

        expect(json_response['data']['name']).to eq('New Category')
        expect(json_response['data']['color']).to eq('#FF5733')
      end

      it 'assigns category to current user' do
        post '/api/v1/categories', params: valid_params, headers: headers, as: :json

        new_category = Category.last
        expect(new_category.user).to eq(user)
      end

      it 'uses default color when not provided' do
        post '/api/v1/categories', params: { category: { name: 'No Color' } }, headers: headers, as: :json

        expect(response).to have_http_status(:created)
        new_category = Category.last
        expect(new_category.color).to be_present
        expect(new_category.color).to match(/^#[0-9A-F]{6}$/i)
      end
    end

    context 'with invalid params' do
      let(:invalid_params) do
        {
          category: {
            name: '',
            color: 'invalid-color'
          }
        }
      end

      it 'returns validation errors' do
        post '/api/v1/categories', params: invalid_params, headers: headers, as: :json

        expect(response).to have_http_status(:unprocessable_content)
        json_response = response.parsed_body

        expect(json_response['error']).to be_present
        expect(json_response['error']['code']).to eq('VALIDATION_FAILED')
        expect(json_response['error']['details']['validation_errors']['name']).to include("can't be blank")
        expect(json_response['error']['details']['validation_errors']['color']).to include('must be a valid hex color')
      end

      it 'returns error when name is duplicate for same user' do
        create(:category, user: user, name: 'Existing')

        expect do
          post '/api/v1/categories', params: { category: { name: 'Existing' } }, headers: headers, as: :json
        end.not_to change(Category, :count)

        expect(response).to have_http_status(:unprocessable_content)
      end

      it 'allows duplicate names for different users' do
        create(:category, user: other_user, name: 'Shared Name')

        expect do
          post '/api/v1/categories', params: { category: { name: 'Shared Name' } }, headers: headers, as: :json
        end.to change(Category, :count).by(1)

        expect(response).to have_http_status(:created)
      end
    end
  end

  describe 'PATCH /api/v1/categories/:id' do
    let(:category) { create(:category, user: user, name: 'Old Name', color: '#000000') }

    let(:update_params) do
      {
        category: {
          name: 'Updated Name',
          color: '#FFFFFF'
        }
      }
    end

    context 'with valid params' do
      it 'updates the category' do
        patch "/api/v1/categories/#{category.id}", params: update_params, headers: headers, as: :json

        expect(response).to have_http_status(:ok)
        json_response = response.parsed_body

        expect(json_response['data']['name']).to eq('Updated Name')
        expect(json_response['data']['color']).to eq('#FFFFFF')

        category.reload
        expect(category.name).to eq('Updated Name')
        expect(category.color).to eq('#FFFFFF')
      end

      it 'allows partial updates' do
        patch "/api/v1/categories/#{category.id}", params: { category: { name: 'Only Name Updated' } }, headers: headers, as: :json

        expect(response).to have_http_status(:ok)

        category.reload
        expect(category.name).to eq('Only Name Updated')
        expect(category.color).to eq('#000000') # Original color unchanged
      end
    end

    context 'with invalid params' do
      it 'returns error when updating with duplicate name' do
        create(:category, user: user, name: 'Taken Name')

        patch "/api/v1/categories/#{category.id}", params: { category: { name: 'Taken Name' } }, headers: headers, as: :json

        expect(response).to have_http_status(:unprocessable_content)
      end
    end

    context 'when category belongs to other user' do
      let(:other_category) { create(:category, user: other_user) }

      it 'returns not found' do
        patch "/api/v1/categories/#{other_category.id}", params: update_params, headers: headers, as: :json

        expect(response).to have_http_status(:not_found)
      end
    end
  end

  describe 'DELETE /api/v1/categories/:id' do
    let!(:category) { create(:category, user: user) }
    let!(:todo_with_category) { create(:todo, user: user, category: category) }

    it 'deletes the category and nullifies todo categories' do
      expect do
        delete "/api/v1/categories/#{category.id}", headers: headers
      end.to change(Category, :count).by(-1)

      expect(response).to have_http_status(:ok)
      json_response = response.parsed_body
      expect(json_response['status']['message']).to eq('Category deleted successfully')

      todo_with_category.reload
      expect(todo_with_category.category).to be_nil
    end
  end

  describe 'authentication' do
    it 'requires authentication for all endpoints' do
      get '/api/v1/categories'
      expect(response.status).to be_in([401, 403])

      post '/api/v1/categories', params: { category: { name: 'Test', color: '#000000' } }, as: :json
      expect(response.status).to be_in([401, 403])

      patch '/api/v1/categories/1', params: { category: { name: 'Test' } }, as: :json
      expect(response.status).to be_in([401, 403])

      delete '/api/v1/categories/1'
      expect(response.status).to be_in([401, 403])
    end
  end
end
