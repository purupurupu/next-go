# frozen_string_literal: true

# Error raised when user lacks permission to perform an action
class AuthorizationError < ApiError
  def initialize(message = nil, resource: nil, action: nil, details: nil)
    details ||= {}
    details[:resource] = resource if resource
    details[:action] = action if action

    super(
      message,
      code: 'AUTHORIZATION_FAILED',
      status: :forbidden,
      details: details
    )
  end

  class << self
    def default_message
      'You are not authorized to perform this action.'
    end

    def default_code
      'AUTHORIZATION_FAILED'
    end
  end
end
