module ApiResponseFormatter
  extend ActiveSupport::Concern

  private

  def success_response(message:, data: nil, status: :ok)
    response_data = {
      status: {
        code: status_code_for(status),
        message: message
      }
    }
    response_data[:data] = data if data.present?

    render json: response_data, status: status
  end

  def error_response(message:, status: :unprocessable_content)
    render json: {
      status: {
        code: status_code_for(status),
        message: message
      }
    }, status: status
  end

  def status_code_for(status)
    case status
    when :ok then 200
    when :created then 201
    when :unauthorized then 401
    when :forbidden then 403
    when :not_found then 404
    when :unprocessable_content then 422
    when :internal_server_error then 500
    else
      Rack::Utils::SYMBOL_TO_STATUS_CODE[status] || 500
    end
  end

  def user_data(user)
    {
      id: user.id,
      email: user.email,
      name: user.name,
      created_at: user.created_at
    }
  end
end
