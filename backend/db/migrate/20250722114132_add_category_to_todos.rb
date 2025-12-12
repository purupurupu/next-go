class AddCategoryToTodos < ActiveRecord::Migration[7.1]
  def change
    add_reference :todos, :category, null: true, foreign_key: true
  end
end
