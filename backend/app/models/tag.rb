class Tag < ApplicationRecord
  belongs_to :user
  has_many :todo_tags, dependent: :destroy
  has_many :todos, through: :todo_tags

  validates :name, presence: true, length: { maximum: 30 }
  validates :name, uniqueness: { scope: :user_id, case_sensitive: false }
  validates :color, format: { with: /\A#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})\z/, message: 'must be a valid hex color' },
                    allow_blank: true

  before_validation :normalize_name
  before_validation :normalize_color

  scope :ordered, -> { order(name: :asc) }

  private

  def normalize_name
    self.name = name&.strip&.downcase if name.present?
  end

  def normalize_color
    self.color = color&.upcase if color.present?
  end
end
