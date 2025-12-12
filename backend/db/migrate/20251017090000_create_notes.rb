class CreateNotes < ActiveRecord::Migration[8.0]
  def change
    create_table :notes do |t|
      t.references :user, null: false, foreign_key: true
      t.string :title, limit: 150
      t.text :body_md
      t.text :body_plain
      t.boolean :pinned, default: false, null: false
      t.datetime :archived_at
      t.datetime :trashed_at
      t.datetime :last_edited_at, null: false, default: -> { 'CURRENT_TIMESTAMP' }

      t.timestamps
    end

    add_index :notes, :pinned
    add_index :notes, :archived_at
    add_index :notes, :trashed_at
    add_index :notes, :last_edited_at
    add_index :notes, :body_plain
    add_index :notes, %i[user_id pinned]
    add_index :notes, %i[user_id archived_at]
    add_index :notes, %i[user_id trashed_at]
    add_index :notes, %i[user_id last_edited_at]
  end
end
