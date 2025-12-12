class CreateTags < ActiveRecord::Migration[7.1]
  def change
    create_table :tags do |t|
      t.string :name, null: false
      t.references :user, null: false, foreign_key: true
      t.string :color, default: "#6B7280"

      t.timestamps
    end

    add_index :tags, [:user_id, :name], unique: true
  end
end
