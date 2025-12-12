class Todo < ApplicationRecord
  # 関連付け（学習ポイント：belongs_to関連）
  belongs_to :user
  belongs_to :category, optional: true, counter_cache: :todos_count
  has_many :todo_tags, dependent: :destroy
  has_many :tags, through: :todo_tags

  # 学習ポイント：ポリモーフィック関連
  # as: :commentableにより、Commentモデルからポリモーフィックに参照される
  has_many :comments, as: :commentable, dependent: :destroy

  # 学習ポイント：変更履歴の追跡
  has_many :todo_histories, dependent: :destroy

  # Active Storage files
  has_many_attached :files do |attachable|
    attachable.variant :thumb, resize_to_limit: [300, 300]
    attachable.variant :medium, resize_to_limit: [800, 800]
  end

  # Enums
  enum :priority, { low: 0, medium: 1, high: 2 }, default: :medium
  enum :status, { pending: 0, in_progress: 1, completed: 2 }, default: :pending

  validates :title, presence: true
  validates :completed, inclusion: { in: [true, false] }
  validate :due_date_cannot_be_in_the_past, if: -> { due_date.present? }
  validate :validate_file_attachments

  scope :ordered, -> { order(position: :asc) }
  before_create :set_position

  # 学習ポイント：変更履歴の記録
  # 作成時と更新時に履歴を記録
  after_create :record_creation
  around_update :track_changes_with_user

  # 学習ポイント：現在のユーザーを一時的に保持（履歴記録用）
  attr_accessor :current_user

  private

  def set_position
    last_position = user.todos.maximum(:position) || 0
    self.position = last_position + 1
  end

  def due_date_cannot_be_in_the_past
    return unless due_date.present? && due_date < Date.current

    errors.add(:due_date, 'は過去の日付にできません')
  end

  def validate_file_attachments
    return unless files.attached?

    files.each do |file|
      # File size validation (max 10MB)
      errors.add(:files, "ファイルサイズは10MB以下にしてください (#{file.filename})") if file.byte_size > 10.megabytes

      # File type validation
      allowed_types = %w[
        image/jpeg image/png image/gif image/webp
        application/pdf
        text/plain text/csv
        application/msword application/vnd.openxmlformats-officedocument.wordprocessingml.document
        application/vnd.ms-excel application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
      ]

      unless allowed_types.include?(file.content_type)
        errors.add(:files, "許可されていないファイルタイプです (#{file.filename}: #{file.content_type})")
      end
    end
  end

  # 学習ポイント：作成時の履歴記録
  def record_creation
    return unless current_user

    todo_histories.create!(
      user: current_user,
      field_name: 'created',
      action: 'created',
      new_value: title
    )
  end

  # 学習ポイント：around_updateコールバックで変更を追跡
  def track_changes_with_user
    # 変更前の状態を保存
    changes_to_track = {}
    tracked_attributes = %w[title status priority due_date completed category_id description]

    tracked_attributes.each do |attr|
      next unless will_save_change_to_attribute?(attr)

      changes_to_track[attr] = {
        old_value: attribute_in_database(attr),
        new_value: send(attr)
      }
    end

    # 実際の更新を実行
    yield

    # 学習ポイント：更新後に履歴を記録
    track_changes(changes_to_track) if current_user && changes_to_track.any?
  end

  # 学習ポイント：更新時の変更追跡（内部メソッド）
  def track_changes(changes_to_track)
    return unless current_user

    changes_to_track.each do |attr, changes|
      old_val = changes[:old_value]
      new_val = changes[:new_value]

      # 学習ポイント：特別なアクションの判定
      action_type = determine_action_type(attr, old_val, new_val)

      todo_histories.create!(
        user: current_user,
        field_name: attr,
        old_value: format_value_for_history(attr, old_val),
        new_value: format_value_for_history(attr, new_val),
        action: action_type
      )
    end
  end

  # 学習ポイント：アクションタイプの判定
  def determine_action_type(attr, _old_val, _new_val)
    case attr
    when 'status'
      'status_changed'
    when 'priority'
      'priority_changed'
    else
      'updated'
    end
  end

  # 学習ポイント：履歴用の値フォーマット
  def format_value_for_history(attr, value)
    case attr
    when 'category_id'
      value.present? ? Category.find_by(id: value)&.name : nil
    when 'due_date'
      value&.to_s
    else
      value.to_s
    end
  end
end
