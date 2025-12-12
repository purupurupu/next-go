class ChangePositionColumnOrder < ActiveRecord::Migration[7.1]
  def up
    # 一時テーブルを作成
    create_table :new_todos do |t|
      t.string :title, null: false
      t.integer :position
      t.boolean :completed, default: false
      t.timestamps
    end

    # データを新しいテーブルにコピー
    execute <<-SQL
      INSERT INTO new_todos (id, title, position, completed, created_at, updated_at)
      SELECT id, title, position, completed, created_at, updated_at
      FROM todos
    SQL

    # 古いテーブルを削除
    drop_table :todos

    # 新しいテーブルの名前を変更
    rename_table :new_todos, :todos

    # インデックスを再作成（必要に応じて）
    add_index :todos, :position
  end

  def down
    # 元に戻すためのロジック（必要に応じて）
    raise ActiveRecord::IrreversibleMigration
  end
end
