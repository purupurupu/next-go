# frozen_string_literal: true

# 学習ポイント：jsonapi-serializerを使用したレスポンスのカスタマイズ
class CommentSerializer
  include JSONAPI::Serializer

  attributes :id, :content, :created_at, :updated_at

  # 学習ポイント：関連モデルの情報も含める
  attribute :user do |object|
    UserSerializer.new(object.user).serializable_hash[:data][:attributes]
  end

  # 学習ポイント：カスタム属性の追加
  # 現在のユーザーがコメントを編集できるかどうかを返す
  attribute :editable do |object, params|
    # current_userはコントローラーから渡される
    current_user = params[:current_user]
    return false unless current_user

    object.editable? && object.owned_by?(current_user)
  end
end
