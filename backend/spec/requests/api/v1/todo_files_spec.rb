require 'rails_helper'

RSpec.describe 'Todo Files API', type: :request do
  let(:user) { create(:user) }
  let(:other_user) { create(:user) }
  let(:auth_headers) { auth_headers_for(user) }
  let(:todo) { create(:todo, user: user) }

  describe 'POST /api/v1/todos with files' do
    let(:valid_attributes) do
      {
        todo: {
          title: 'Test Todo with Files',
          completed: false,
          priority: 'medium',
          status: 'pending'
        }
      }
    end

    context 'when creating todo with files' do
      let(:text_file) { fixture_file_upload('test_file.txt', 'text/plain') }
      let(:image_file) { fixture_file_upload('test_image.png', 'image/png') }

      it 'creates a todo with a single file' do
        post api_v1_todos_path,
             params: valid_attributes.merge(todo: valid_attributes[:todo].merge(files: [text_file])),
             headers: auth_headers

        expect(response).to have_http_status(:created)

        json_response = response.parsed_body
        todo_data = json_response['data']
        expect(todo_data['files']).to be_an(Array)
        expect(todo_data['files'].length).to eq(1)
        expect(todo_data['files'][0]['filename']).to eq('test_file.txt')
        expect(todo_data['files'][0]['content_type']).to eq('text/plain')
        expect(todo_data['files'][0]['url']).to be_present
      end

      it 'creates a todo with multiple files' do
        post api_v1_todos_path,
             params: valid_attributes.merge(todo: valid_attributes[:todo].merge(files: [text_file, image_file])),
             headers: auth_headers

        expect(response).to have_http_status(:created)

        json_response = response.parsed_body
        todo_data = json_response['data']
        expect(todo_data['files']).to be_an(Array)
        expect(todo_data['files'].length).to eq(2)

        filenames = todo_data['files'].pluck('filename')
        expect(filenames).to contain_exactly('test_file.txt', 'test_image.png')
      end
    end
  end

  describe 'PATCH /api/v1/todos/:id with files' do
    let(:text_file) { fixture_file_upload('test_file.txt', 'text/plain') }
    let(:image_file) { fixture_file_upload('test_image.png', 'image/png') }

    it 'adds files to existing todo' do
      patch api_v1_todo_path(todo),
            params: { todo: { files: [text_file] } },
            headers: auth_headers

      expect(response).to have_http_status(:ok)

      json_response = response.parsed_body
      todo_data = json_response['data']
      expect(todo_data['files']).to be_an(Array)
      expect(todo_data['files'].length).to eq(1)
      expect(todo_data['files'][0]['filename']).to eq('test_file.txt')
    end

    it 'preserves existing files when adding new ones' do
      # First add a file
      patch api_v1_todo_path(todo),
            params: { todo: { files: [text_file] } },
            headers: auth_headers

      expect(response).to have_http_status(:ok)

      # Then add another file
      patch api_v1_todo_path(todo),
            params: { todo: { files: [image_file] } },
            headers: auth_headers

      expect(response).to have_http_status(:ok)

      json_response = response.parsed_body
      todo_data = json_response['data']
      expect(todo_data['files']).to be_an(Array)
      expect(todo_data['files'].length).to eq(2)

      filenames = todo_data['files'].pluck('filename')
      expect(filenames).to contain_exactly('test_file.txt', 'test_image.png')
    end
  end

  describe 'DELETE /api/v1/todos/:id/files/:file_id' do
    let(:text_file) { fixture_file_upload('test_file.txt', 'text/plain') }
    let(:image_file) { fixture_file_upload('test_image.png', 'image/png') }

    before do
      # Add files to todo
      patch api_v1_todo_path(todo),
            params: { todo: { files: [text_file, image_file] } },
            headers: auth_headers

      todo.reload
    end

    it 'deletes a specific file' do
      expect(todo.files.count).to eq(2)

      file_to_delete = todo.files.first

      delete "/api/v1/todos/#{todo.id}/files/#{file_to_delete.id}",
             headers: auth_headers

      expect(response).to have_http_status(:ok)

      json_response = response.parsed_body
      todo_data = json_response['data']
      expect(todo_data['files'].length).to eq(1)

      todo.reload
      expect(todo.files.count).to eq(1)
    end

    it 'returns 404 for non-existent file' do
      delete "/api/v1/todos/#{todo.id}/files/non-existent-id",
             headers: auth_headers

      expect(response).to have_http_status(:not_found)
    end

    it 'returns 404 for file from another todo' do
      other_todo = create(:todo, user: user)
      patch api_v1_todo_path(other_todo),
            params: { todo: { files: [fixture_file_upload('test_file.txt', 'text/plain')] } },
            headers: auth_headers

      other_file = other_todo.files.first

      delete "/api/v1/todos/#{todo.id}/files/#{other_file.id}",
             headers: auth_headers

      expect(response).to have_http_status(:not_found)
    end
  end

  describe 'authorization' do
    it 'prevents other users from accessing todos' do
      # First verify basic todo access control
      other_headers = auth_headers_for(other_user)

      get api_v1_todo_path(todo), headers: other_headers
      expect(response).to have_http_status(:not_found)
    end

    it 'prevents other users from deleting files' do
      # Add file to todo
      text_file = fixture_file_upload('test_file.txt', 'text/plain')
      patch api_v1_todo_path(todo),
            params: { todo: { files: [text_file] } },
            headers: auth_headers

      expect(response).to have_http_status(:ok)
      todo.reload
      file = todo.files.first
      expect(file).to be_present

      # Store IDs to avoid lazy evaluation issues
      file_id = file.id
      todo_id = todo.id

      # Sign out current user to ensure clean authentication state
      # This prevents JWT token context pollution in tests
      delete '/auth/sign_out', headers: auth_headers

      # Authenticate as other_user and attempt to delete the file
      other_headers = auth_headers_for(other_user)

      delete "/api/v1/todos/#{todo_id}/files/#{file_id}",
             headers: other_headers

      # Should get 404 because other_user cannot access user's todo
      expect(response).to have_http_status(:not_found)

      # Verify file still exists
      todo.reload
      expect(todo.files.count).to eq(1)
    end
  end
end
