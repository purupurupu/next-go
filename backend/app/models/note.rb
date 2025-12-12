class Note < ApplicationRecord
  belongs_to :user
  has_many :note_revisions, dependent: :destroy

  attr_accessor :current_user

  validates :title, length: { maximum: 150 }, allow_blank: true
  validates :body_md, length: { maximum: 100_000 }, allow_nil: true

  scope :not_trashed, -> { where(trashed_at: nil) }
  scope :trashed, -> { where.not(trashed_at: nil) }
  scope :archived, -> { where.not(archived_at: nil) }
  scope :active, -> { not_trashed.where(archived_at: nil) }
  scope :pinned_first, -> { order(pinned: :desc, updated_at: :desc) }

  before_save :set_body_plain
  before_save :touch_last_edited_at

  after_create_commit :record_initial_revision
  after_update_commit :record_revision_if_changed

  private

  def set_body_plain
    return unless new_record? || will_save_change_to_body_md?

    self.body_plain = body_md.to_s
  end

  def touch_last_edited_at
    return unless will_save_change_to_title? || will_save_change_to_body_md?

    self.last_edited_at = Time.current
  end

  def record_initial_revision
    return unless current_user

    create_revision
  end

  def record_revision_if_changed
    return unless current_user
    return unless saved_change_to_title? || saved_change_to_body_md?

    create_revision
  end

  def create_revision
    note_revisions.create!(user: current_user, title: title, body_md: body_md)
    prune_revisions
  end

  def prune_revisions(limit = 50)
    excess = note_revisions.order(created_at: :desc).offset(limit)
    excess.destroy_all if excess.exists?
  end
end
