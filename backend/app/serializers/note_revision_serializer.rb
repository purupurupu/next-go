class NoteRevisionSerializer
  include JSONAPI::Serializer

  attributes :id, :note_id, :title, :body_md, :created_at
end
