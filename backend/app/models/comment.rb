# frozen_string_literal: true

# == Schema Information
#
# Table name: comments
#
#  id               :bigint           not null, primary key
#  user_id          :bigint           not null
#  commentable_type :string           not null
#  commentable_id   :bigint           not null
#  content          :text             not null
#  deleted_at       :datetime
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#
# Indexes
#
#  index_comments_on_commentable_and_deleted_at  (commentable_type,commentable_id,deleted_at)
#  index_comments_on_deleted_at                  (deleted_at)
#  index_comments_on_user_id                     (user_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#

class Comment < ApplicationRecord
  # 学習ポイント：Polymorphic関連付け
  # コメントを様々なモデル（Todo、将来的にはProjectなど）に関連付けることができる
  belongs_to :user
  belongs_to :commentable, polymorphic: true

  # バリデーション
  validates :content, presence: true, length: { maximum: 1000 }

  # 学習ポイント：ソフトデリート機能
  # 履歴保持のため、実際にレコードを削除せずに論理削除を行う
  scope :active, -> { where(deleted_at: nil) }
  scope :deleted, -> { where.not(deleted_at: nil) }

  # デフォルトスコープで削除済みコメントを除外
  default_scope { active }

  # 学習ポイント：タイムスタンプ順のソート
  scope :chronological, -> { order(created_at: :asc) }
  scope :recent, -> { order(created_at: :desc) }

  # ソフトデリートメソッド
  def soft_delete!
    update!(deleted_at: Time.current)
  end

  # 復元メソッド（必要に応じて）
  def restore!
    update!(deleted_at: nil)
  end

  def deleted?
    deleted_at.present?
  end

  # コメントの編集可能時間（例：作成から15分間）
  def editable?
    !deleted? && created_at > 15.minutes.ago
  end

  # 学習ポイント：コメント作成者かどうかを判定
  def owned_by?(user)
    self.user == user
  end
end
