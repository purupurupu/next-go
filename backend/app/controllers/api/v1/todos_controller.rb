# frozen_string_literal: true

module Api
  module V1
    class TodosController < BaseController
      include TodoSearchFilters
      include TodoSearchSuggestions

      before_action :set_todo, only: %i[show update destroy update_tags destroy_file]

      def index
        @todos = current_user.todos.includes(:category, :tags, :comments).ordered
        render_json_response(
          data: @todos,
          each_serializer: TodoSerializer,
          message: 'Todos retrieved successfully'
        )
      end

      def search
        @todos = TodoSearchService.new(current_user, search_params).call

        meta_data = {
          total: @todos.total_count,
          current_page: @todos.current_page,
          total_pages: @todos.total_pages,
          per_page: @todos.limit_value,
          search_query: search_params[:q] || search_params[:query] || search_params[:search],
          filters_applied: active_filters
        }

        # Add helpful feedback when no results found
        meta_data[:suggestions] = search_suggestions if @todos.total_count.zero?

        render_json_response(
          data: @todos,
          each_serializer: TodoSerializer,
          highlight_query: search_params[:q] || search_params[:query] || search_params[:search],
          meta: meta_data,
          message: 'Search completed successfully'
        )
      end

      def show
        render_json_response(
          data: @todo,
          serializer: TodoSerializer,
          message: 'Todo retrieved successfully'
        )
      end

      def create
        @todo = current_user.todos.build(todo_params.except(:tag_ids, :files))

        # 学習ポイント：履歴記録のためにcurrent_userを設定
        @todo.current_user = current_user

        if params[:todo][:tag_ids].present?
          valid_tag_ids = current_user.tags.where(id: params[:todo][:tag_ids]).pluck(:id)
          @todo.tag_ids = valid_tag_ids
        end

        if @todo.save
          # Attach files after successful save
          @todo.files.attach(params[:todo][:files]) if params[:todo][:files].present?

          render_json_response(
            data: @todo,
            serializer: TodoSerializer,
            status: :created,
            message: 'Todo created successfully'
          )
        else
          render_error_response(
            error: ::ValidationError.new(errors: @todo.errors),
            status: :unprocessable_content
          )
        end
      end

      def update
        # 学習ポイント：履歴記録のためにcurrent_userを設定
        @todo.current_user = current_user

        if params[:todo][:tag_ids].present?
          valid_tag_ids = current_user.tags.where(id: params[:todo][:tag_ids]).pluck(:id)
          @todo.tag_ids = valid_tag_ids
        end

        # Handle file attachments separately to append rather than replace
        @todo.files.attach(params[:todo][:files]) if params[:todo][:files].present?

        if @todo.update(todo_params.except(:tag_ids, :files))
          render_json_response(
            data: @todo,
            serializer: TodoSerializer,
            message: 'Todo updated successfully'
          )
        else
          render_error_response(
            error: ::ValidationError.new(errors: @todo.errors),
            status: :unprocessable_content
          )
        end
      end

      def destroy
        @todo.destroy
        render_json_response(
          message: 'Todo deleted successfully',
          status: :no_content
        )
      end

      def update_tags
        tag_ids = (params[:tag_ids] || []).map(&:to_i)

        # Validate that all tags belong to current user
        user_tag_ids = current_user.tags.where(id: tag_ids).pluck(:id)

        if user_tag_ids.sort == tag_ids.sort
          @todo.tag_ids = tag_ids
          render_json_response(
            data: @todo,
            serializer: TodoSerializer,
            message: 'Tags updated successfully'
          )
        else
          render_error_response(
            error: 'Invalid tag IDs',
            status: :unprocessable_content,
            details: { invalid_tags: tag_ids - user_tag_ids }
          )
        end
      end

      def update_order
        todo_ids = params[:todos].pluck(:id)
        user_todos = current_user.todos.where(id: todo_ids).index_by(&:id)

        # Validate all todos belong to current user
        if user_todos.size != todo_ids.size
          return render_error_response(
            error: 'Some todos not found',
            status: :not_found,
            details: { missing_todos: todo_ids - user_todos.keys }
          )
        end

        # Prepare bulk update data
        updates = params[:todos].map do |todo_data|
          todo = user_todos[todo_data[:id]]
          { id: todo.id, position: todo_data[:position] }
        end

        # Perform bulk update using case statement
        ActiveRecord::Base.transaction do
          updates.each do |update_data|
            current_user.todos.where(id: update_data[:id]).update_all(position: update_data[:position])
          end
        end

        render_json_response(
          message: 'Todo order updated successfully'
        )
      rescue StandardError => e
        Rails.logger.error "Failed to update todo order: #{e.message}" unless Rails.env.test? || defined?(RSpec)
        render_error_response(
          error: 'Failed to update todo order',
          status: :unprocessable_content,
          details: { error: e.message }
        )
      end

      def destroy_file
        file = @todo.files.find(params[:file_id])
        file.purge
        render_json_response(
          data: @todo,
          serializer: TodoSerializer,
          message: 'File deleted successfully'
        )
      rescue ActiveRecord::RecordNotFound
        render_error_response(
          error: 'File not found',
          status: :not_found
        )
      end

      private

      def set_todo
        @todo = current_user.todos.find(params[:id])
      rescue ActiveRecord::RecordNotFound
        render_error_response(
          error: 'Todo not found',
          status: :not_found
        )
      end

      def todo_params
        params.require(:todo).permit(:title, :completed, :position, :due_date, :priority, :status, :description,
                                     :category_id, tag_ids: [], files: [])
      end

      def search_params
        # Handle both single values and arrays for status and priority
        permitted = params.permit(
          :q, :query, :search,
          :category_id,
          :due_date_from, :due_date_to,
          :sort_by, :sort_order,
          :tag_mode,
          :page, :per_page,
          :status, :priority, # Allow single values
          status: [],
          priority: [],
          tag_ids: []
        )

        # Convert single values to arrays if needed
        permitted[:status] = [permitted[:status]] if permitted[:status].present? && !permitted[:status].is_a?(Array)

        if permitted[:priority].present? && !permitted[:priority].is_a?(Array)
          permitted[:priority] = [permitted[:priority]]
        end

        permitted
      end
    end
  end
end
