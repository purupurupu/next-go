# frozen_string_literal: true

module Api
  module V1
    class NotesController < BaseController
      before_action :set_note, only: %i[show update destroy]

      def index
        notes = current_user.notes
        notes = apply_filters(notes)
        notes = notes.order(pinned: :desc).order(last_edited_at: :desc, updated_at: :desc)
        notes = paginate(notes)

        meta = pagination_meta(notes).merge(filters: filter_meta)

        render_json_response(
          data: notes,
          each_serializer: NoteSerializer,
          meta: meta,
          message: 'Notes retrieved successfully'
        )
      end

      def show
        render_json_response(
          data: @note,
          serializer: NoteSerializer,
          message: 'Note retrieved successfully'
        )
      end

      def create
        @note = current_user.notes.new(note_params)
        @note.current_user = current_user

        if @note.save
          render_json_response(
            data: @note,
            serializer: NoteSerializer,
            status: :created,
            message: 'Note created successfully'
          )
        else
          render_error_response(
            error: ::ValidationError.new(errors: @note.errors),
            status: :unprocessable_content
          )
        end
      end

      def update
        @note.current_user = current_user
        apply_state_flags(@note)

        if @note.update(note_params)
          render_json_response(
            data: @note,
            serializer: NoteSerializer,
            message: 'Note updated successfully'
          )
        else
          render_error_response(
            error: ::ValidationError.new(errors: @note.errors),
            status: :unprocessable_content
          )
        end
      end

      def destroy
        if truthy_param?(params[:force])
          @note.destroy!
          render_json_response(message: 'Note deleted', status: :no_content)
        else
          @note.update!(trashed_at: Time.current)
          render_json_response(
            data: @note,
            serializer: NoteSerializer,
            message: 'Note moved to trash'
          )
        end
      end

      private

      def set_note
        @note = current_user.notes.find(params[:id])
      end

      def note_params
        params.fetch(:note, {}).permit(:title, :body_md, :pinned)
      end

      def apply_filters(scope)
        scoped = default_state_scope(scope)
        scoped = filter_pinned(scoped)
        filter_search(scoped)
      end

      def default_state_scope(scope)
        return scope.trashed if truthy_param?(params[:trashed])
        return scope.archived if truthy_param?(params[:archived])

        scope.active
      end

      def filter_pinned(scope)
        return scope if params[:pinned].nil?

        truthy_param?(params[:pinned]) ? scope.where(pinned: true) : scope.where(pinned: false)
      end

      def filter_search(scope)
        query = params[:q].presence || params[:query].presence || params[:search].presence
        return scope unless query

        term = "%#{query.downcase}%"
        scope.where("LOWER(COALESCE(title, '')) LIKE :q OR LOWER(COALESCE(body_plain, '')) LIKE :q", q: term)
      end

      def apply_state_flags(note)
        return if params[:note].blank?

        state_params = params.require(:note).permit(:archived, :trashed)
        note.archived_at = truthy_param?(state_params[:archived]) ? Time.current : nil if state_params.key?(:archived)
        note.trashed_at = truthy_param?(state_params[:trashed]) ? Time.current : nil if state_params.key?(:trashed)
      end

      def paginate(scope)
        scope.page(page_param).per(per_page_param)
      end

      def page_param
        (params[:page] || 1).to_i
      end

      def per_page_param
        per_page = (params[:per_page] || 20).to_i
        return 1 if per_page < 1
        return 100 if per_page > 100

        per_page
      end

      def pagination_meta(scope)
        {
          total: scope.total_count,
          current_page: scope.current_page,
          total_pages: scope.total_pages,
          per_page: scope.limit_value
        }
      end

      def filter_meta
        {
          archived: truthy_param?(params[:archived]) || false,
          trashed: truthy_param?(params[:trashed]) || false,
          pinned: params[:pinned],
          query: params[:q] || params[:query] || params[:search]
        }.compact
      end

      def truthy_param?(value)
        ActiveModel::Type::Boolean.new.cast(value)
      end
    end
  end
end
