class TagSerializer
  include JSONAPI::Serializer

  attributes :id, :name, :color, :created_at, :updated_at
end
