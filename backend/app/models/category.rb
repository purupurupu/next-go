class Category < ApplicationRecord
  belongs_to :user
  has_many :todos, dependent: :nullify

  validates :name, presence: true, length: { maximum: 50 }
  validates :name, uniqueness: { scope: :user_id, case_sensitive: false }
  validates :color, presence: true,
                    format: { with: /\A#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})\z/, message: 'must be a valid hex color' }

  before_validation :normalize_color

  private

  def normalize_color
    self.color = color&.upcase if color.present?
  end
end
