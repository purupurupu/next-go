class AddSearchIndexesToTodos < ActiveRecord::Migration[7.1]
  def change
    # Add indexes for search and filtering performance
    
    # Text search indexes (only add if not exists)
    add_index :todos, :title unless index_exists?(:todos, :title)
    add_index :todos, :description unless index_exists?(:todos, :description)
    
    # Filter indexes (only add if not exists)
    add_index :todos, :status unless index_exists?(:todos, :status)
    add_index :todos, :priority unless index_exists?(:todos, :priority)
    add_index :todos, :due_date unless index_exists?(:todos, :due_date)
    add_index :todos, :category_id unless index_exists?(:todos, :category_id)
    
    # Composite indexes for common query patterns
    add_index :todos, [:user_id, :status] unless index_exists?(:todos, [:user_id, :status])
    add_index :todos, [:user_id, :priority] unless index_exists?(:todos, [:user_id, :priority])
    add_index :todos, [:user_id, :due_date] unless index_exists?(:todos, [:user_id, :due_date])
    add_index :todos, [:user_id, :category_id] unless index_exists?(:todos, [:user_id, :category_id])
    
    # Index for sorting by created_at and updated_at
    add_index :todos, :created_at unless index_exists?(:todos, :created_at)
    add_index :todos, :updated_at unless index_exists?(:todos, :updated_at)
    
    # Composite index for user + position (for drag-and-drop ordering)
    add_index :todos, [:user_id, :position] unless index_exists?(:todos, [:user_id, :position])
    
    # Add indexes for todo_tags junction table
    add_index :todo_tags, [:todo_id, :tag_id], unique: true unless index_exists?(:todo_tags, [:todo_id, :tag_id])
    add_index :todo_tags, :tag_id unless index_exists?(:todo_tags, :tag_id)
  end
end
