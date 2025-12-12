class TodoSerializer
  include JSONAPI::Serializer

  attributes :id, :title, :completed, :position, :due_date, :priority, :status,
             :description, :user_id, :created_at, :updated_at

  attribute :category do |object|
    next nil unless object.category

    {
      id: object.category.id,
      name: object.category.name,
      color: object.category.color
    }
  end

  attribute :tags do |object|
    object.tags.map do |tag|
      {
        id: tag.id,
        name: tag.name,
        color: tag.color
      }
    end
  end

  attribute :files do |object|
    next [] unless object.files.attached?

    object.files.map do |file|
      {
        id: file.id,
        filename: file.filename.to_s,
        content_type: file.content_type,
        byte_size: file.byte_size,
        url: Rails.application.routes.url_helpers.rails_blob_url(file, host: 'localhost:3001')
      }
    end
  end

  # 学習ポイント：コメント数を効率的に取得
  attribute :comments_count do |object|
    # N+1クエリを避けるため、counter_cacheを使用することも検討
    object.comments.count
  end

  # 学習ポイント：最新のコメントを3件まで取得
  attribute :latest_comments do |object, params|
    # 最新の3件のコメントを取得してシリアライズ
    comments = object.comments
                     .includes(:user)
                     .recent
                     .limit(3)

    comments.map do |comment|
      CommentSerializer.new(comment,
                            params: { current_user: params[:current_user] }).serializable_hash[:data][:attributes]
    end
  end

  # 学習ポイント：履歴数の取得
  attribute :history_count do |object|
    object.todo_histories.count
  end

  # 検索結果のハイライト情報
  attribute :highlights do |object, params|
    query = params[:highlight_query]
    next nil if query.blank?

    highlights = {}
    search_term = query.downcase

    # タイトルでのマッチ位置を検出
    if object.title.downcase.include?(search_term)
      highlights[:title] =
        TodoSerializer.find_match_positions(object.title, search_term)
    end

    # 説明文でのマッチ位置を検出
    if object.description.present? && object.description.downcase.include?(search_term)
      highlights[:description] = TodoSerializer.find_match_positions(object.description, search_term)
    end

    highlights.presence
  end

  def self.find_match_positions(text, search_term)
    positions = []
    text_lower = text.downcase
    index = 0

    while (match_index = text_lower.index(search_term, index))
      positions << {
        start: match_index,
        end: match_index + search_term.length,
        matched_text: text[match_index, search_term.length]
      }
      index = match_index + 1
    end

    positions
  end
end
