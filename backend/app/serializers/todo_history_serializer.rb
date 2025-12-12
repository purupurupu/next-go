# frozen_string_literal: true

# 学習ポイント：履歴データのシリアライズ
class TodoHistorySerializer
  include JSONAPI::Serializer

  attributes :id, :field_name, :old_value, :new_value, :action, :created_at

  # 学習ポイント：関連ユーザー情報の含める
  attribute :user do |object|
    UserSerializer.new(object.user).serializable_hash[:data][:attributes]
  end

  # 学習ポイント：可読性の高い変更内容を含める
  attribute :human_readable_change, &:human_readable_change
end
