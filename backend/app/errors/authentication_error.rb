# frozen_string_literal: true

# Error raised when authentication fails
class AuthenticationError < ApiError
  def initialize(message = nil, details: nil)
    super(
      message,
      code: 'AUTHENTICATION_FAILED',
      status: :unauthorized,
      details: details
    )
  end

  class << self
    def default_message
      'Authentication failed. Please check your credentials.'
    end

    def default_code
      'AUTHENTICATION_FAILED'
    end
  end
end
