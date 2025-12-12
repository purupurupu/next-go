class User < ApplicationRecord
  # Include default devise modules. Others available are:
  # :confirmable, :lockable, :timeoutable, :trackable and :omniauthable
  devise :database_authenticatable, :registerable,
         :recoverable, :rememberable, :validatable,
         :jwt_authenticatable, jwt_revocation_strategy: JwtDenylist

  # 関連付け（学習ポイント：has_many関連）
  has_many :todos, dependent: :destroy
  has_many :categories, dependent: :destroy
  has_many :tags, dependent: :destroy
  has_many :comments, dependent: :destroy
  has_many :notes, dependent: :destroy

  # バリデーション（学習ポイント：カスタムバリデーション）
  validates :name, presence: true, length: { minimum: 2, maximum: 50 }
  validates :email, uniqueness: { case_sensitive: false }
end
