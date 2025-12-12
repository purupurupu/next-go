class CreateTodoHistories < ActiveRecord::Migration[7.1]
  def change
    create_table :todo_histories do |t|
      t.references :todo, null: false, foreign_key: true
      t.references :user, null: false, foreign_key: true
      t.string :field_name, null: false
      t.text :old_value
      t.text :new_value
      t.integer :action, null: false, default: 0

      # 学習ポイント：created_atのみ必要（履歴は更新されない）
      t.datetime :created_at, null: false
    end
    
    # 学習ポイント：履歴検索のためのインデックス
    add_index :todo_histories, [:todo_id, :created_at]
    add_index :todo_histories, :field_name
  end
end
