# frozen_string_literal: true

# Middleware for handling API errors consistently
class ErrorHandler
  def initialize(app)
    @app = app
  end

  def call(env)
    request = ActionDispatch::Request.new(env)

    # Add request ID if not present
    env['action_dispatch.request_id'] ||= SecureRandom.uuid

    begin
      @app.call(env)
    rescue StandardError => e
      handle_error(e, request)
    end
  end

  private

  def handle_error(error, request)
    case error
    when ::ApiError
      render_api_error(error, request)
    when ActiveRecord::RecordNotFound
      handle_record_not_found(error, request)
    when ActiveRecord::RecordInvalid
      handle_validation_error(error, request)
    when ActionController::ParameterMissing
      handle_parameter_missing(error, request)
    when JWT::DecodeError, JWT::ExpiredSignature
      handle_jwt_error(error, request)
    else
      handle_standard_error(error, request)
    end
  end

  def render_api_error(error, request)
    log_error(error, request, :warn)

    response_body = {
      error: {
        code: error.code,
        message: error.message,
        details: error.details,
        request_id: request.request_id,
        timestamp: Time.current.iso8601
      }
    }

    build_response(response_body, error.status)
  end

  def handle_record_not_found(error, request)
    log_error(error, request, :warn)

    # Extract model name from error message
    model_name = error.model || extract_model_from_message(error.message)

    not_found_error = ::NotFoundError.new(
      resource: model_name,
      id: error.id
    )

    render_api_error(not_found_error, request)
  end

  def handle_validation_error(error, request)
    log_error(error, request, :warn)

    validation_error = ::ValidationError.new(
      errors: error.record.errors
    )

    render_api_error(validation_error, request)
  end

  def handle_parameter_missing(error, request)
    log_error(error, request, :warn)

    response_body = {
      error: {
        code: 'PARAMETER_MISSING',
        message: "Required parameter missing: #{error.param}",
        details: { missing_parameter: error.param },
        request_id: request.request_id,
        timestamp: Time.current.iso8601
      }
    }

    build_response(response_body, :bad_request)
  end

  def handle_jwt_error(error, request)
    log_error(error, request, :warn)

    auth_error = ::AuthenticationError.new(
      'Invalid or expired token',
      details: { error_type: error.class.name }
    )

    render_api_error(auth_error, request)
  end

  def handle_standard_error(error, request)
    log_error(error, request, :error)

    # In production, hide internal error details
    message = Rails.env.production? ? 'An unexpected error occurred' : error.message

    response_body = {
      error: {
        code: 'INTERNAL_ERROR',
        message: message,
        details: Rails.env.production? ? {} : { error_class: error.class.name },
        request_id: request.request_id,
        timestamp: Time.current.iso8601
      }
    }

    build_response(response_body, :internal_server_error)
  end

  def build_response(body, status)
    status_code = Rack::Utils.status_code(status)
    headers = {
      'Content-Type' => 'application/json',
      'X-Request-Id' => body.dig(:error, :request_id) || SecureRandom.uuid
    }

    [status_code, headers, [body.to_json]]
  end

  def log_error(error, request, level)
    # Skip logging in test environment or when running specs
    return if Rails.env.test? || ENV['RAILS_ENV'] == 'test' || defined?(RSpec)

    Rails.logger.send(level) do
      {
        error_class: error.class.name,
        error_message: error.message,
        request_id: request.request_id,
        request_method: request.request_method,
        request_path: request.path,
        request_params: filtered_params(request),
        backtrace: level == :error ? error.backtrace&.first(10) : nil
      }.compact.to_json
    end
  end

  def filtered_params(request)
    # Use Rails parameter filtering
    filter = ActionDispatch::Http::ParameterFilter.new(Rails.application.config.filter_parameters)
    filter.filter(request.params)
  rescue StandardError
    {}
  end

  def extract_model_from_message(message)
    # Try to extract model name from standard Rails error messages
    match = message.match(/Couldn't find (\w+)/)
    match ? match[1] : nil
  end
end
