class NoteRevision < ApplicationRecord
  belongs_to :note
  belongs_to :user

  validates :title, length: { maximum: 150 }, allow_blank: true
end
