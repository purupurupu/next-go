class NoteSerializer
  include JSONAPI::Serializer

  attributes :id, :title, :body_md, :pinned,
             :archived_at, :trashed_at, :last_edited_at,
             :created_at, :updated_at

  attribute :archived do |object|
    object.archived_at.present?
  end

  attribute :trashed do |object|
    object.trashed_at.present?
  end
end
