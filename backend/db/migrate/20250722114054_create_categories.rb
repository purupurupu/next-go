class CreateCategories < ActiveRecord::Migration[7.1]
  def change
    create_table :categories do |t|
      t.string :name, null: false
      t.string :color, null: false, default: '#6B7280'
      t.references :user, null: false, foreign_key: true

      t.timestamps
    end

    add_index :categories, [:user_id, :name], unique: true
  end
end
