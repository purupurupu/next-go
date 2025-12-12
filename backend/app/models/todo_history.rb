# frozen_string_literal: true

# == Schema Information
#
# Table name: todo_histories
#
#  id         :bigint           not null, primary key
#  todo_id    :bigint           not null
#  user_id    :bigint           not null
#  field_name :string           not null
#  old_value  :text
#  new_value  :text
#  action     :integer          not null, default: 0
#  created_at :datetime         not null
#
# Indexes
#
#  index_todo_histories_on_field_name           (field_name)
#  index_todo_histories_on_todo_id_and_created_at  (todo_id,created_at)
#  index_todo_histories_on_user_id              (user_id)
#
# Foreign Keys
#
#  fk_rails_...  (todo_id => todos.id)
#  fk_rails_...  (user_id => users.id)
#

class TodoHistory < ApplicationRecord
  # 学習ポイント：変更履歴の追跡
  belongs_to :todo
  belongs_to :user

  # バリデーション
  validates :field_name, presence: true
  validates :action, presence: true

  # 学習ポイント：アクションの種類を定義
  enum :action, {
    created: 0,
    updated: 1,
    deleted: 2,
    status_changed: 3,
    priority_changed: 4
  }

  # 学習ポイント：デフォルトスコープで最新順にソート
  default_scope { order(created_at: :desc) }

  # スコープ
  scope :recent, ->(limit = 10) { limit(limit) }
  scope :for_field, ->(field) { where(field_name: field) }
  scope :by_user, ->(user) { where(user: user) }

  # 学習ポイント：変更内容の可読化
  def human_readable_change
    case field_name
    when 'title'
      "タイトルを「#{old_value}」から「#{new_value}」に変更"
    when 'status'
      "ステータスを「#{translate_status(old_value)}」から「#{translate_status(new_value)}」に変更"
    when 'priority'
      "優先度を「#{translate_priority(old_value)}」から「#{translate_priority(new_value)}」に変更"
    when 'due_date'
      format_due_date_change
    when 'completed'
      new_value == 'true' ? 'タスクを完了にマーク' : 'タスクを未完了にマーク'
    when 'category_id'
      'カテゴリを変更'
    when 'description'
      "説明を#{old_value.present? ? '更新' : '追加'}"
    else
      "#{field_name}を変更"
    end
  end

  # 学習ポイント：履歴のグループ化（同一時刻の変更をまとめる）
  def self.grouped_by_timestamp
    all.group_by { |history| history.created_at.to_i }
  end

  private

  def translate_status(value)
    return '未設定' if value.blank?

    status_map = {
      'pending' => '未着手',
      'in_progress' => '進行中',
      'completed' => '完了'
    }
    status_map[value] || value
  end

  def translate_priority(value)
    return '未設定' if value.blank?

    priority_map = {
      'low' => '低',
      'medium' => '中',
      'high' => '高'
    }
    priority_map[value] || value
  end

  def format_due_date_change
    old_date = old_value.present? ? Date.parse(old_value).strftime('%Y年%m月%d日') : 'なし'
    new_date = new_value.present? ? Date.parse(new_value).strftime('%Y年%m月%d日') : 'なし'
    "期限日を「#{old_date}」から「#{new_date}」に変更"
  rescue Date::Error
    '期限日を変更'
  end
end
