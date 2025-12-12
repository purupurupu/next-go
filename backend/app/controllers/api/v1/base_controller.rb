# frozen_string_literal: true

module Api
  module V1
    # Base controller for API v1 endpoints
    # Provides common functionality for all v1 API controllers
    class BaseController < ApplicationController
      # API version information
      API_VERSION = 'v1'
      API_VERSION_HEADER = 'X-API-Version'

      before_action :set_api_version_header

      private

      # Set API version in response headers
      def set_api_version_header
        response.headers[API_VERSION_HEADER] = API_VERSION
      end

      # Override render to ensure consistent API responses
      def render_json_response(data: nil, message: nil, status: :ok, serializer: nil, each_serializer: nil, **options)
        # For no_content status, use head method to avoid body
        if status == :no_content
          head :no_content
        else
          response_options = {
            data: data,
            message: message,
            status: status,
            serializer: serializer,
            each_serializer: each_serializer
          }.merge(options)
          response_body = build_response_body(response_options)
          render json: response_body, status: status
        end
      end

      # render_error_response is inherited from ApplicationController

      def build_response_body(options)
        response_body = {
          status: {
            code: Rack::Utils.status_code(options[:status]),
            message: options[:message] || default_message_for(options[:status])
          }
        }

        if options[:data].present?
          serialized_data = if options[:serializer] || options[:each_serializer]
                              serialize_data(options[:data], options[:serializer], options[:each_serializer], options)
                            else
                              options[:data]
                            end

          response_body[:data] = serialized_data
        end

        response_body[:meta] = options[:meta] if options[:meta].present?
        response_body
      end

      def serialize_data(data, serializer, each_serializer, options)
        # Extract custom params from options
        custom_params = options.except(:data, :message, :status, :serializer, :each_serializer, :meta, :params)

        # Merge all params including current_user and any custom params
        all_params = { current_user: current_user }.merge(custom_params).merge(options[:params] || {})
        serializer_options = { params: all_params }

        if each_serializer
          # For collections, serialize each item and extract attributes
          data.map do |item|
            serialized = each_serializer.new(item, serializer_options).serializable_hash[:data]
            # Return attributes directly with id from the serialized data
            serialized[:attributes]
          end
        elsif serializer
          # For single items
          serialized = serializer.new(data, serializer_options).serializable_hash[:data]
          # Return attributes directly with id from the serialized data
          serialized[:attributes]
        else
          data
        end
      end

      # error_code_for and extract_error_details are inherited from ApplicationController

      def extract_error_details(error)
        case error
        when ActiveRecord::RecordInvalid
          error.record.errors.details
        when ActionController::ParameterMissing
          { missing_parameter: error.param }
        else
          {}
        end
      end

      def default_message_for(status)
        case status
        when :ok, :no_content
          'Request processed successfully'
        when :created
          'Resource created successfully'
        else
          'Request processed'
        end
      end
    end
  end
end
