class CategorySerializer
  include JSONAPI::Serializer

  attributes :id, :name, :color, :created_at, :updated_at

  attribute :todo_count, &:todos_count
end
