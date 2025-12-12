class AddPriorityAndStatusToTodos < ActiveRecord::Migration[7.1]
  def change
    add_column :todos, :priority, :integer, default: 1, null: false
    add_column :todos, :status, :integer, default: 0, null: false
    
    add_index :todos, :priority
    add_index :todos, :status
  end
end
