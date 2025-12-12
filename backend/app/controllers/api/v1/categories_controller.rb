# frozen_string_literal: true

module Api
  module V1
    class CategoriesController < BaseController
      before_action :authenticate_user!
      before_action :set_category, only: %i[show update destroy]

      # GET /api/v1/categories
      def index
        @categories = current_user.categories.order(:name)
        render_json_response(
          data: @categories,
          each_serializer: CategorySerializer,
          message: 'Categories retrieved successfully'
        )
      end

      # GET /api/v1/categories/:id
      def show
        render_json_response(
          data: @category,
          serializer: CategorySerializer,
          message: 'Category retrieved successfully'
        )
      end

      # POST /api/v1/categories
      def create
        @category = current_user.categories.build(category_params)

        if @category.save
          render_json_response(
            data: @category,
            serializer: CategorySerializer,
            status: :created,
            message: 'Category created successfully'
          )
        else
          render_error_response(
            error: ::ValidationError.new(errors: @category.errors),
            status: :unprocessable_content
          )
        end
      end

      # PATCH/PUT /api/v1/categories/:id
      def update
        if @category.update(category_params)
          render_json_response(
            data: @category,
            serializer: CategorySerializer,
            message: 'Category updated successfully'
          )
        else
          render_error_response(
            error: ::ValidationError.new(errors: @category.errors),
            status: :unprocessable_content
          )
        end
      end

      # DELETE /api/v1/categories/:id
      def destroy
        @category.destroy
        render_json_response(
          message: 'Category deleted successfully',
          status: :ok
        )
      end

      private

      def set_category
        @category = current_user.categories.find(params[:id])
      rescue ActiveRecord::RecordNotFound
        render_error_response(
          error: 'Category not found',
          status: :not_found
        )
      end

      def category_params
        params.require(:category).permit(:name, :color)
      end
    end
  end
end
