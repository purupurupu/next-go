require 'rails_helper'

RSpec.describe 'Api::V1::Tags', type: :request do
  let(:user) { create(:user) }
  let(:headers) { auth_headers_for(user) }
  let(:valid_attributes) { { name: 'Work', color: '#FF0000' } }
  let(:invalid_attributes) { { name: '', color: 'invalid' } }

  describe 'GET /api/v1/tags' do
    context 'when authenticated' do
      before do
        create_list(:tag, 3, user: user)
        create(:tag, user: create(:user)) # Another user's tag
      end

      it 'returns all tags for the current user' do
        get '/api/v1/tags', headers: headers
        expect(response).to have_http_status(:ok)
        expect(response.parsed_body['data'].size).to eq(3)
      end

      it 'returns tags ordered by name' do
        create(:tag, name: 'Zebra', user: user)
        create(:tag, name: 'Alpha', user: user)
        create(:tag, name: 'Beta', user: user)

        get '/api/v1/tags', headers: headers
        names = response.parsed_body['data'].pluck('name')
        expect(names).to eq(names.sort)
      end

      it "does not return other users' tags" do
        get '/api/v1/tags', headers: headers
        returned_tags = response.parsed_body['data']
        # Since user_id is not in the serializer, we need to verify by checking the count
        expect(returned_tags.size).to eq(3) # Only the user's 3 tags
      end
    end

    context 'when not authenticated' do
      it 'returns unauthorized status' do
        get '/api/v1/tags'
        expect(response).to have_http_status(:unauthorized)
      end
    end
  end

  describe 'GET /api/v1/tags/:id' do
    let(:tag) { create(:tag, user: user) }

    context 'when authenticated' do
      it 'returns the tag' do
        get "/api/v1/tags/#{tag.id}", headers: headers
        expect(response).to have_http_status(:ok)
        expect(response.parsed_body['data']['id']).to eq(tag.id)
        expect(response.parsed_body['data']['name']).to eq(tag.name)
        expect(response.parsed_body['data']['color']).to eq(tag.color)
      end

      it 'returns 404 for non-existent tag' do
        get '/api/v1/tags/999999', headers: headers
        expect(response).to have_http_status(:not_found)
        expect(response.parsed_body['error']['message']).to eq('Tag not found')
      end

      it "returns 404 for another user's tag" do
        another_tag = create(:tag, user: create(:user))
        get "/api/v1/tags/#{another_tag.id}", headers: headers
        expect(response).to have_http_status(:not_found)
      end
    end

    context 'when not authenticated' do
      it 'returns unauthorized status' do
        get "/api/v1/tags/#{tag.id}"
        expect(response).to have_http_status(:unauthorized)
      end
    end
  end

  describe 'POST /api/v1/tags' do
    context 'when authenticated' do
      context 'with valid parameters' do
        it 'creates a new tag' do
          expect do
            post '/api/v1/tags', params: { tag: valid_attributes }, headers: headers, as: :json
          end.to change(Tag, :count).by(1)
        end

        it 'returns the created tag' do
          post '/api/v1/tags', params: { tag: valid_attributes }, headers: headers, as: :json
          expect(response).to have_http_status(:created)
          expect(response.parsed_body['data']['name']).to eq('work') # normalized
          expect(response.parsed_body['data']['color']).to eq('#FF0000')
        end

        it 'normalizes the tag name' do
          post '/api/v1/tags', params: { tag: { name: '  WORK  ' } }, headers: headers, as: :json
          expect(response.parsed_body['data']['name']).to eq('work')
        end

        it 'normalizes the color' do
          post '/api/v1/tags', params: { tag: { name: 'test', color: '#ff0000' } }, headers: headers, as: :json
          expect(response.parsed_body['data']['color']).to eq('#FF0000')
        end
      end

      context 'with invalid parameters' do
        it 'does not create a new tag' do
          expect do
            post '/api/v1/tags', params: { tag: invalid_attributes }, headers: headers, as: :json
          end.not_to change(Tag, :count)
        end

        it 'returns unprocessable entity status' do
          post '/api/v1/tags', params: { tag: invalid_attributes }, headers: headers, as: :json
          expect(response).to have_http_status(:unprocessable_content)
        end

        it 'returns validation errors' do
          post '/api/v1/tags', params: { tag: { name: '' } }, headers: headers, as: :json
          json = response.parsed_body
          expect(json['error']['code']).to eq('VALIDATION_FAILED')
          expect(json['error']['details']['validation_errors']['name']).to include("can't be blank")
        end

        it 'returns error for duplicate name' do
          create(:tag, name: 'work', user: user)
          post '/api/v1/tags', params: { tag: { name: 'WORK' } }, headers: headers, as: :json
          expect(response).to have_http_status(:unprocessable_content)
          json = response.parsed_body
          expect(json['error']['code']).to eq('VALIDATION_FAILED')
          expect(json['error']['details']['validation_errors']['name']).to include('has already been taken')
        end

        it 'returns error for invalid color format' do
          post '/api/v1/tags', params: { tag: { name: 'test', color: 'red' } }, headers: headers, as: :json
          json = response.parsed_body
          expect(json['error']['code']).to eq('VALIDATION_FAILED')
          expect(json['error']['details']['validation_errors']['color']).to include('must be a valid hex color')
        end
      end
    end

    context 'when not authenticated' do
      it 'returns unauthorized status' do
        post '/api/v1/tags', params: { tag: valid_attributes }
        expect(response).to have_http_status(:unauthorized)
      end
    end
  end

  describe 'PATCH /api/v1/tags/:id' do
    let(:tag) { create(:tag, name: 'old-name', color: '#000000', user: user) }
    let(:new_attributes) { { name: 'new-name', color: '#FFFFFF' } }

    context 'when authenticated' do
      context 'with valid parameters' do
        it 'updates the tag' do
          patch "/api/v1/tags/#{tag.id}", params: { tag: new_attributes }, headers: headers, as: :json
          tag.reload
          expect(tag.name).to eq('new-name')
          expect(tag.color).to eq('#FFFFFF')
        end

        it 'returns the updated tag' do
          patch "/api/v1/tags/#{tag.id}", params: { tag: new_attributes }, headers: headers, as: :json
          expect(response).to have_http_status(:ok)
          expect(response.parsed_body['data']['name']).to eq('new-name')
          expect(response.parsed_body['data']['color']).to eq('#FFFFFF')
        end
      end

      context 'with invalid parameters' do
        it 'returns unprocessable entity status' do
          patch "/api/v1/tags/#{tag.id}", params: { tag: invalid_attributes }, headers: headers, as: :json
          expect(response).to have_http_status(:unprocessable_content)
        end

        it 'returns validation errors' do
          patch "/api/v1/tags/#{tag.id}", params: { tag: { name: '' } }, headers: headers, as: :json
          json = response.parsed_body
          expect(json['error']['code']).to eq('VALIDATION_FAILED')
          expect(json['error']['details']['validation_errors']['name']).to include("can't be blank")
        end
      end

      it 'returns 404 for non-existent tag' do
        patch '/api/v1/tags/999999', params: { tag: new_attributes }, headers: headers, as: :json
        expect(response).to have_http_status(:not_found)
      end

      it "returns 404 for another user's tag" do
        another_tag = create(:tag, user: create(:user))
        patch "/api/v1/tags/#{another_tag.id}", params: { tag: new_attributes }, headers: headers, as: :json
        expect(response).to have_http_status(:not_found)
      end
    end

    context 'when not authenticated' do
      it 'returns unauthorized status' do
        patch "/api/v1/tags/#{tag.id}", params: { tag: new_attributes }
        expect(response).to have_http_status(:unauthorized)
      end
    end
  end

  describe 'DELETE /api/v1/tags/:id' do
    let!(:tag) { create(:tag, user: user) }

    context 'when authenticated' do
      it 'destroys the tag' do
        expect do
          delete "/api/v1/tags/#{tag.id}", headers: headers
        end.to change(Tag, :count).by(-1)
      end

      it 'returns ok status with success message' do
        delete "/api/v1/tags/#{tag.id}", headers: headers
        expect(response).to have_http_status(:ok)
        json_response = response.parsed_body
        expect(json_response['status']['message']).to eq('Tag deleted successfully')
      end

      it 'destroys associated todo_tags' do
        todo = create(:todo, user: user)
        create(:todo_tag, todo: todo, tag: tag)

        expect do
          delete "/api/v1/tags/#{tag.id}", headers: headers
        end.to change(TodoTag, :count).by(-1)
      end

      it 'returns 404 for non-existent tag' do
        delete '/api/v1/tags/999999', headers: headers
        expect(response).to have_http_status(:not_found)
      end

      it "returns 404 for another user's tag" do
        another_tag = create(:tag, user: create(:user))
        delete "/api/v1/tags/#{another_tag.id}", headers: headers
        expect(response).to have_http_status(:not_found)
      end
    end

    context 'when not authenticated' do
      it 'returns unauthorized status' do
        delete "/api/v1/tags/#{tag.id}"
        expect(response).to have_http_status(:unauthorized)
      end
    end
  end
end
