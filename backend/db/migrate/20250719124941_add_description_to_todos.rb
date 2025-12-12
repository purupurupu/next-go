class AddDescriptionToTodos < ActiveRecord::Migration[7.1]
  def change
    add_column :todos, :description, :text
  end
end
