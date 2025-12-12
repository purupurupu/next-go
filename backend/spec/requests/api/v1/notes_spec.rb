require 'rails_helper'

RSpec.describe 'Notes API', type: :request do
  let(:user) { create(:user) }
  let(:other_user) { create(:user) }
  let(:headers) { auth_headers_for(user) }

  describe 'authentication' do
    it 'rejects unauthenticated access' do
      get '/api/v1/notes'

      expect(response).to have_http_status(:unauthorized)
    end
  end

  describe 'GET /api/v1/notes' do
    before do
      create(:note, user: user, title: 'Active')
      create(:note, user: user, title: 'Archived', archived_at: 1.day.ago)
      create(:note, user: user, title: 'Trash', trashed_at: 1.day.ago)
      create(:note, user: other_user, title: 'Other user')
    end

    it 'returns only active notes by default' do
      get '/api/v1/notes', headers: headers

      expect(response).to have_http_status(:ok)
      data = response.parsed_body['data']
      expect(data.size).to eq(1)
      expect(data.first['title']).to eq('Active')
    end

    it 'filters trashed notes' do
      get '/api/v1/notes', params: { trashed: true }, headers: headers

      expect(response).to have_http_status(:ok)
      titles = response.parsed_body['data'].pluck('title')
      expect(titles).to contain_exactly('Trash')
    end

    it 'filters archived notes' do
      get '/api/v1/notes', params: { archived: true }, headers: headers

      expect(response).to have_http_status(:ok)
      titles = response.parsed_body['data'].pluck('title')
      expect(titles).to contain_exactly('Archived')
    end

    it 'filters by pinned flag' do
      create(:note, user: user, title: 'Pinned note', pinned: true)

      get '/api/v1/notes', params: { pinned: true }, headers: headers
      titles = response.parsed_body['data'].pluck('title')
      expect(titles).to include('Pinned note')
      expect(titles).not_to include('Active')
    end

    it 'searches by title/body' do
      create(:note, user: user, title: 'Meeting notes', body_md: 'Discuss roadmap')

      get '/api/v1/notes', params: { q: 'roadmap' }, headers: headers

      data = response.parsed_body['data']
      expect(data.size).to eq(1)
      expect(data.first['title']).to eq('Meeting notes')
    end
  end

  describe 'POST /api/v1/notes' do
    it 'creates a note' do
      payload = { note: { title: 'New note', body_md: 'Hello' } }

      post '/api/v1/notes', params: payload.to_json, headers: headers

      expect(response).to have_http_status(:created)
      data = response.parsed_body['data']
      expect(data['title']).to eq('New note')
      expect(data['body_md']).to eq('Hello')
    end
  end

  describe 'PATCH /api/v1/notes/:id' do
    let(:note) { create(:note, user: user, pinned: false) }

    it 'updates content and pinned flag' do
      payload = { note: { title: 'Updated', body_md: 'Changed', pinned: true } }

      patch "/api/v1/notes/#{note.id}", params: payload.to_json, headers: headers

      expect(response).to have_http_status(:ok)
      data = response.parsed_body['data']
      expect(data['title']).to eq('Updated')
      expect(data['pinned']).to be(true)
    end

    it 'archives and restores' do
      patch "/api/v1/notes/#{note.id}", params: { note: { archived: true } }.to_json, headers: headers
      expect(response).to have_http_status(:ok)
      archived_at = response.parsed_body['data']['archived_at']
      expect(archived_at).to be_present

      patch "/api/v1/notes/#{note.id}", params: { note: { archived: false } }.to_json, headers: headers
      expect(response.parsed_body['data']['archived_at']).to be_nil
    end

    it 'trashes and restores' do
      patch "/api/v1/notes/#{note.id}", params: { note: { trashed: true } }.to_json, headers: headers
      expect(response.parsed_body['data']['trashed_at']).to be_present

      patch "/api/v1/notes/#{note.id}", params: { note: { trashed: false } }.to_json, headers: headers
      expect(response.parsed_body['data']['trashed_at']).to be_nil
    end
  end

  describe 'DELETE /api/v1/notes/:id' do
    let!(:note) { create(:note, user: user) }

    it 'soft deletes by default' do
      delete "/api/v1/notes/#{note.id}", headers: headers

      expect(response).to have_http_status(:ok)
      expect(response.parsed_body['data']['trashed_at']).to be_present
      expect(note.reload.trashed_at).to be_present
    end

    it 'force deletes when requested' do
      delete "/api/v1/notes/#{note.id}?force=true", headers: headers

      expect(response).to have_http_status(:no_content)
      expect { note.reload }.to raise_error(ActiveRecord::RecordNotFound)
    end
  end

  describe 'authorization' do
    let(:note) { create(:note, user: other_user) }

    it 'does not allow accessing another users note' do
      get "/api/v1/notes/#{note.id}", headers: headers

      expect(response).to have_http_status(:not_found)
    end
  end
end
