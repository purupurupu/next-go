class AddPositionToTodos < ActiveRecord::Migration[7.1]
  def change
    add_column :todos, :position, :integer
  end
end
