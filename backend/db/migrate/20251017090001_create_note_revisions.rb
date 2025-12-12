class CreateNoteRevisions < ActiveRecord::Migration[8.0]
  def change
    create_table :note_revisions do |t|
      t.references :note, null: false, foreign_key: true
      t.references :user, null: false, foreign_key: true
      t.string :title, limit: 150
      t.text :body_md

      t.timestamps
    end

    add_index :note_revisions, %i[note_id created_at]
  end
end
