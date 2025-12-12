# frozen_string_literal: true

# Error raised when rate limit is exceeded
class RateLimitError < ApiError
  def initialize(message = nil, limit: nil, remaining: 0, reset_at: nil, details: nil)
    details ||= {}
    details[:limit] = limit if limit
    details[:remaining] = remaining
    details[:reset_at] = reset_at.iso8601 if reset_at

    super(
      message,
      code: 'RATE_LIMIT_EXCEEDED',
      status: :too_many_requests,
      details: details
    )
  end

  class << self
    def default_message
      'Rate limit exceeded. Please try again later.'
    end

    def default_code
      'RATE_LIMIT_EXCEEDED'
    end
  end
end
