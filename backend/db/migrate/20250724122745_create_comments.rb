class CreateComments < ActiveRecord::Migration[7.1]
  def change
    create_table :comments do |t|
      t.references :user, null: false, foreign_key: true
      t.references :commentable, polymorphic: true, null: false
      t.text :content, null: false
      t.datetime :deleted_at

      t.timestamps
    end

    # 学習ポイント：ポリモーフィック関連のためのインデックス
    # deleted_atを含めた複合インデックスで、ソフトデリート時のパフォーマンスを向上
    add_index :comments, [:commentable_type, :commentable_id, :deleted_at], 
              name: 'index_comments_on_commentable_and_deleted_at'
    add_index :comments, :deleted_at
  end
end
