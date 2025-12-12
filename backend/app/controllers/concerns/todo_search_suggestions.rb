# frozen_string_literal: true

module TodoSearchSuggestions
  extend ActiveSupport::Concern

  private

  def search_suggestions
    [].tap do |suggestions|
      add_search_query_suggestions(suggestions)
      add_filter_count_suggestion(suggestions)
      add_status_filter_suggestion(suggestions)
      add_tag_mode_suggestion(suggestions)
      add_date_range_suggestion(suggestions)
      add_general_suggestions(suggestions)
    end
  end

  def add_search_query_suggestions(suggestions)
    return unless search_query_present?

    suggestions << {
      type: 'spelling',
      message: '検索キーワードのスペルを確認してください。'
    }
    suggestions << {
      type: 'broader_search',
      message: 'より一般的なキーワードで検索してみてください。'
    }
  end

  def add_filter_count_suggestion(suggestions)
    return unless active_filters.size > 3

    suggestions << {
      type: 'reduce_filters',
      message: 'フィルター条件を減らしてみてください。',
      current_filters: active_filters.keys
    }
  end

  def add_status_filter_suggestion(suggestions)
    return unless search_params[:status].present? && Array(search_params[:status]).size > 1

    suggestions << {
      type: 'status_filter',
      message: 'ステータスフィルターを1つに絞ってみてください。'
    }
  end

  def add_tag_mode_suggestion(suggestions)
    return unless search_params[:tag_ids].present? && search_params[:tag_mode] == 'all'

    suggestions << {
      type: 'tag_mode',
      message: 'タグの検索モードを「いずれか」（ANY）に変更してみてください。'
    }
  end

  def add_date_range_suggestion(suggestions)
    return unless search_params[:due_date_from].present? && search_params[:due_date_to].present?

    suggestions << {
      type: 'date_range',
      message: '日付範囲を広げてみてください。'
    }
  end

  def add_general_suggestions(suggestions)
    suggestions << {
      type: 'clear_filters',
      message: 'すべてのフィルターをクリアして、もう一度お試しください。'
    }
  end

  def search_query_present?
    search_params[:q].present? || search_params[:query].present? || search_params[:search].present?
  end
end
