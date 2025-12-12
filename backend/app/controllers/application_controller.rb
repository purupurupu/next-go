class ApplicationController < ActionController::API
  before_action :authenticate_user!, unless: :devise_registration_or_session_controller?
  before_action :configure_permitted_parameters, if: :devise_controller?
  before_action :set_request_id

  # Unified error handling using custom error classes
  rescue_from ::ApiError, with: :handle_api_error
  rescue_from ActiveRecord::RecordNotFound, with: :handle_not_found
  rescue_from ActiveRecord::RecordInvalid, with: :handle_unprocessable_entity
  rescue_from ActionController::ParameterMissing, with: :handle_parameter_missing
  rescue_from JWT::DecodeError, JWT::ExpiredSignature, with: :handle_authentication_error

  private

  def devise_registration_or_session_controller?
    devise_controller? && (
      controller_name == 'registrations' ||
      (controller_name == 'sessions' && action_name.in?(%w[create new destroy]))
    )
  end

  def configure_permitted_parameters
    devise_parameter_sanitizer.permit(:sign_up, keys: [:name])
    devise_parameter_sanitizer.permit(:account_update, keys: [:name])
  end

  # Set request ID for tracking
  def set_request_id
    response.headers['X-Request-Id'] = request.request_id
  end

  # Error handlers
  def handle_api_error(exception)
    log_error(exception, :warn)
    render_error_response(
      error: exception,
      status: exception.status
    )
  end

  def handle_not_found(exception)
    log_error(exception, :warn)
    # Extract model and id from exception if possible
    model_name = exception.model if exception.respond_to?(:model)
    record_id = exception.id if exception.respond_to?(:id)

    error = ::NotFoundError.new(
      resource: model_name,
      id: record_id
    )

    render_error_response(error: error, status: :not_found)
  end

  def handle_unprocessable_entity(exception)
    log_error(exception, :warn)
    error = ::ValidationError.new(
      errors: exception.record&.errors
    )

    render_error_response(error: error, status: :unprocessable_content)
  end

  def handle_parameter_missing(exception)
    log_error(exception, :warn)
    render_error_response(
      error: "Required parameter missing: #{exception.param}",
      status: :bad_request,
      details: { missing_parameter: exception.param }
    )
  end

  def handle_authentication_error(exception)
    log_error(exception, :warn)
    error = ::AuthenticationError.new(
      'Invalid or expired token',
      details: { error_type: exception.class.name }
    )

    render_error_response(error: error, status: :unauthorized)
  end

  # Common error response rendering
  def render_error_response(error:, status: :unprocessable_content, details: nil)
    error_body = if error.is_a?(::ApiError)
                   {
                     error: {
                       code: error.code,
                       message: error.message,
                       details: error.details,
                       request_id: request.request_id,
                       timestamp: Time.current.iso8601
                     }
                   }
                 else
                   {
                     error: {
                       code: error_code_for(error),
                       message: error.is_a?(String) ? error : error.message,
                       details: details || {},
                       request_id: request.request_id,
                       timestamp: Time.current.iso8601
                     }
                   }
                 end

    render json: error_body, status: status
  end

  def error_code_for(error)
    case error
    when String
      'ERROR'
    else
      error.class.name.underscore.upcase
    end
  end

  def log_error(exception, level = :error)
    # Skip logging in test environment or when running specs
    return if Rails.env.test? || ENV['RAILS_ENV'] == 'test' || defined?(RSpec)

    Rails.logger.send(level) do
      {
        error_class: exception.class.name,
        error_message: exception.message,
        request_id: request.request_id,
        controller: "#{controller_name}##{action_name}",
        params: filtered_params,
        backtrace: level == :error ? exception.backtrace&.first(5) : nil
      }.compact.to_json
    end
  end

  def filtered_params
    # Use Rails parameter filtering
    filter = ActionDispatch::Http::ParameterFilter.new(Rails.application.config.filter_parameters)
    filter.filter(params.to_unsafe_h)
  rescue StandardError
    {}
  end
end
