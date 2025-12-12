require 'rails_helper'

RSpec.describe 'Note Revisions API', type: :request do
  let(:user) { create(:user) }
  let(:other_user) { create(:user) }
  let(:headers) { auth_headers_for(user) }

  def create_note_via_api
    payload = { note: { title: 'First', body_md: 'Original body' } }
    post '/api/v1/notes', params: payload.to_json, headers: headers
    response.parsed_body['data']
  end

  describe 'GET /api/v1/notes/:note_id/revisions' do
    it 'lists revisions for a note' do
      note = create_note_via_api
      patch "/api/v1/notes/#{note['id']}", params: { note: { body_md: 'Updated body' } }.to_json, headers: headers

      get "/api/v1/notes/#{note['id']}/revisions", headers: headers

      expect(response).to have_http_status(:ok)
      revisions = response.parsed_body['data']
      expect(revisions.size).to be >= 1
      expect(revisions.first).to include('body_md')
    end

    it 'does not allow access to other users note revisions' do
      note = create(:note, user: other_user)
      create(:note_revision, note: note, user: other_user)

      get "/api/v1/notes/#{note.id}/revisions", headers: headers

      expect(response).to have_http_status(:not_found)
    end
  end

  describe 'POST /api/v1/notes/:note_id/revisions/:id/restore' do
    it 'restores a previous revision' do
      note = create_note_via_api
      patch "/api/v1/notes/#{note['id']}", params: { note: { body_md: 'Second version' } }.to_json, headers: headers

      get "/api/v1/notes/#{note['id']}/revisions", headers: headers
      revisions = response.parsed_body['data']
      older_revision = revisions.last

      post "/api/v1/notes/#{note['id']}/revisions/#{older_revision['id']}/restore", headers: headers

      expect(response).to have_http_status(:ok)
      restored = response.parsed_body['data']
      expect(restored['body_md']).to eq('Original body')
    end
  end
end
