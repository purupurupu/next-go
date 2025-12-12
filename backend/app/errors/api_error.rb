# frozen_string_literal: true

# Base class for all API errors
# Provides structured error responses with codes, messages, and details
class ApiError < StandardError
  attr_reader :code, :status, :details

  def initialize(message = nil, code: nil, status: :bad_request, details: nil)
    @message = message || self.class.default_message
    @code = code || self.class.default_code
    @status = status
    @details = details || {}
    super(@message)
  end

  def to_json(*_args)
    {
      error: {
        code: @code,
        message: @message,
        details: @details
      }
    }.to_json
  end

  def to_h
    {
      error: {
        code: @code,
        message: @message,
        details: @details
      }
    }
  end

  class << self
    def default_message
      'An error occurred'
    end

    def default_code
      'API_ERROR'
    end
  end
end
