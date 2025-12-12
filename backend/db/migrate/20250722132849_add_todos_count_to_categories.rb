class AddTodosCountToCategories < ActiveRecord::Migration[7.1]
  def up
    add_column :categories, :todos_count, :integer, default: 0, null: false
    
    # Reset counter cache for existing categories
    Category.find_each do |category|
      Category.reset_counters(category.id, :todos)
    end
  end

  def down
    remove_column :categories, :todos_count
  end
end
