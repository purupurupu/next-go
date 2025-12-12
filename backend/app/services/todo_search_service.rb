# frozen_string_literal: true

class TodoSearchService
  attr_reader :user, :params

  def initialize(user, params = {})
    @user = user
    @params = params.to_h.with_indifferent_access
  end

  def call
    # Build query
    scope = user.todos
    scope = apply_search(scope)
    scope = apply_filters(scope)
    scope = apply_sorting(scope)
    scope = apply_includes(scope)
    apply_pagination(scope)
  end

  private

  def apply_search(scope)
    return scope if search_query.blank?

    search_term = "%#{search_query.downcase}%"

    scope.where(
      'LOWER(todos.title) LIKE :query OR LOWER(todos.description) LIKE :query',
      query: search_term
    )
  end

  def search_query
    params[:q] || params[:query] || params[:search]
  end

  def apply_filters(scope)
    scope = filter_by_category(scope)
    scope = filter_by_status(scope)
    scope = filter_by_priority(scope)
    scope = filter_by_tags(scope)
    filter_by_date_range(scope)
  end

  def filter_by_category(scope)
    return scope if params[:category_id].blank?

    category_ids = Array(params[:category_id]).map(&:to_i).reject(&:zero?)

    if category_ids.include?(-1) || params[:category_id] == 'null'
      # Include todos without category
      if category_ids.size > 1
        scope.where(category_id: category_ids.reject { |id| id == -1 }).or(scope.where(category_id: nil))
      else
        scope.where(category_id: nil)
      end
    else
      scope.where(category_id: category_ids)
    end
  end

  def filter_by_status(scope)
    return scope if params[:status].blank?

    # Handle both single string and array of strings
    status_values = params[:status].is_a?(Array) ? params[:status] : [params[:status]]
    statuses = status_values.map(&:to_s).select { |s| Todo.statuses.key?(s) }
    return scope if statuses.empty?

    scope.where(status: statuses)
  end

  def filter_by_priority(scope)
    return scope if params[:priority].blank?

    # Handle both single string and array of strings
    priority_values = params[:priority].is_a?(Array) ? params[:priority] : [params[:priority]]
    priorities = priority_values.map(&:to_s).select { |p| Todo.priorities.key?(p) }
    return scope if priorities.empty?

    scope.where(priority: priorities)
  end

  def filter_by_tags(scope)
    return scope if params[:tag_ids].blank?

    tag_ids = Array(params[:tag_ids]).map(&:to_i).reject(&:zero?)
    return scope if tag_ids.empty?

    # Validate tag IDs belong to current user
    valid_tag_ids = user.tags.where(id: tag_ids).pluck(:id)

    # If no valid tags were found, return empty result
    # This handles both non-existent tags and tags from other users
    return scope.none if valid_tag_ids.empty?

    # Filter by tag IDs with AND/OR logic
    tag_mode = params[:tag_mode] || 'any' # 'any' for OR, 'all' for AND

    if tag_mode == 'all'
      # Find todos that have ALL specified tags
      todo_ids = user.todos.joins(:todo_tags)
                     .where(todo_tags: { tag_id: valid_tag_ids })
                     .group('todos.id')
                     .having('COUNT(DISTINCT todo_tags.tag_id) = ?', valid_tag_ids.size)
                     .pluck(:id)
      scope.where(id: todo_ids)
    else
      # Find todos that have ANY of the specified tags (OR)
      scope.joins(:todo_tags).where(todo_tags: { tag_id: valid_tag_ids }).distinct
    end
  end

  def filter_by_date_range(scope)
    scope = filter_by_due_date_from(scope)
    filter_by_due_date_to(scope)
  end

  def filter_by_due_date_from(scope)
    return scope if params[:due_date_from].blank?

    begin
      date_from = Date.parse(params[:due_date_from])
      scope.where(due_date: date_from..)
    rescue ArgumentError
      # Invalid date format, skip filter
      scope
    end
  end

  def filter_by_due_date_to(scope)
    return scope if params[:due_date_to].blank?

    begin
      date_to = Date.parse(params[:due_date_to])
      scope.where(due_date: ..date_to)
    rescue ArgumentError
      # Invalid date format, skip filter
      scope
    end
  end

  def apply_sorting(scope)
    sort_by = params[:sort_by] || 'position'
    sort_order = normalize_sort_order

    sorting_methods = {
      'created_at' => -> { sort_by_timestamp(scope, :created_at, sort_order) },
      'updated_at' => -> { sort_by_timestamp(scope, :updated_at, sort_order) },
      'due_date' => -> { sort_by_due_date(scope, sort_order) },
      'title' => -> { sort_by_title(scope, sort_order) },
      'priority' => -> { sort_by_enum_field(scope, :priority, sort_order) },
      'status' => -> { sort_by_enum_field(scope, :status, sort_order) }
    }

    method = sorting_methods[sort_by]
    method ? method.call : scope.ordered
  end

  def normalize_sort_order
    params[:sort_order]&.downcase == 'desc' ? 'DESC' : 'ASC'
  end

  def sort_by_timestamp(scope, field, order)
    scope.order(field => order)
  end

  def sort_by_due_date(scope, order)
    scope.order(Arel.sql("due_date IS NULL, due_date #{order}"))
  end

  def sort_by_title(scope, order)
    scope.order(Arel.sql("LOWER(title) #{order}"))
  end

  def sort_by_enum_field(scope, field, order)
    scope.order(field => order)
  end

  def apply_includes(scope)
    scope.includes(:category, :tags, :comments, :user, files_attachments: :blob)
  end

  def apply_pagination(scope)
    page = (params[:page] || 1).to_i
    per_page = (params[:per_page] || 20).to_i

    # Limit per_page to prevent abuse
    per_page = 100 if per_page > 100
    per_page = 1 if per_page < 1

    scope.page(page).per(per_page)
  end
end
