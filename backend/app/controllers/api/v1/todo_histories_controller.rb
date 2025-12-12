# frozen_string_literal: true

module Api
  module V1
    # 学習ポイント：Todoの変更履歴を表示するコントローラー
    class TodoHistoriesController < ApplicationController
      before_action :authenticate_user!
      before_action :set_todo

      # GET /api/v1/todos/:todo_id/histories
      def index
        # 学習ポイント：履歴を新しい順に取得し、関連データを含める
        @histories = @todo.todo_histories
                          .includes(:user)
                          .recent(50) # 最新50件まで

        serialized_histories = @histories.map do |history|
          serialized = TodoHistorySerializer.new(history,
                                                 params: { current_user: current_user }).serializable_hash[:data]
          serialized[:attributes]
        end
        render json: serialized_histories
      end

      private

      def set_todo
        # 学習ポイント：ユーザースコープでのセキュリティ確保
        @todo = current_user.todos.find(params[:todo_id])
      rescue ActiveRecord::RecordNotFound
        render json: { error: 'Todo not found' }, status: :not_found
      end
    end
  end
end
