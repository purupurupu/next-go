# frozen_string_literal: true

module Api
  module V1
    class TagsController < BaseController
      before_action :authenticate_user!
      before_action :set_tag, only: %i[show update destroy]

      # GET /api/v1/tags
      def index
        @tags = current_user.tags.ordered
        render_json_response(
          data: @tags,
          each_serializer: TagSerializer,
          message: 'Tags retrieved successfully'
        )
      end

      # GET /api/v1/tags/:id
      def show
        render_json_response(
          data: @tag,
          serializer: TagSerializer,
          message: 'Tag retrieved successfully'
        )
      end

      # POST /api/v1/tags
      def create
        @tag = current_user.tags.build(tag_params)

        if @tag.save
          render_json_response(
            data: @tag,
            serializer: TagSerializer,
            status: :created,
            message: 'Tag created successfully'
          )
        else
          render_error_response(
            error: ::ValidationError.new(errors: @tag.errors),
            status: :unprocessable_content
          )
        end
      end

      # PATCH/PUT /api/v1/tags/:id
      def update
        if @tag.update(tag_params)
          render_json_response(
            data: @tag,
            serializer: TagSerializer,
            message: 'Tag updated successfully'
          )
        else
          render_error_response(
            error: ::ValidationError.new(errors: @tag.errors),
            status: :unprocessable_content
          )
        end
      end

      # DELETE /api/v1/tags/:id
      def destroy
        @tag.destroy
        render_json_response(
          message: 'Tag deleted successfully',
          status: :ok
        )
      end

      private

      def set_tag
        @tag = current_user.tags.find(params[:id])
      rescue ActiveRecord::RecordNotFound
        render_error_response(
          error: 'Tag not found',
          status: :not_found
        )
      end

      def tag_params
        params.require(:tag).permit(:name, :color)
      end
    end
  end
end
