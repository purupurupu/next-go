# frozen_string_literal: true

module TodoSearchFilters
  extend ActiveSupport::Concern

  private

  def active_filters
    {}.tap do |filters|
      add_search_filter(filters)
      add_category_filter(filters)
      add_status_filter(filters)
      add_priority_filter(filters)
      add_tag_filter(filters)
      add_date_range_filter(filters)
    end
  end

  def add_search_filter(filters)
    search_query = search_params[:q] || search_params[:query] || search_params[:search]
    filters[:search] = search_query if search_query.present?
  end

  def add_category_filter(filters)
    filters[:category_id] = search_params[:category_id] if search_params[:category_id].present?
  end

  def add_status_filter(filters)
    filters[:status] = Array(search_params[:status]) if search_params[:status].present?
  end

  def add_priority_filter(filters)
    filters[:priority] = Array(search_params[:priority]) if search_params[:priority].present?
  end

  def add_tag_filter(filters)
    filters[:tag_ids] = search_params[:tag_ids] if search_params[:tag_ids].present?
  end

  def add_date_range_filter(filters)
    return unless search_params[:due_date_from].present? || search_params[:due_date_to].present?

    filters[:date_range] = {
      from: search_params[:due_date_from],
      to: search_params[:due_date_to]
    }
  end
end
