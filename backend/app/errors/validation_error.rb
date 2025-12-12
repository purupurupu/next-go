# frozen_string_literal: true

# Error raised when validation fails
class ValidationError < ApiError
  def initialize(message = nil, errors: nil, details: nil)
    details ||= {}
    details[:validation_errors] = format_errors(errors) if errors

    super(
      message,
      code: 'VALIDATION_FAILED',
      status: :unprocessable_content,
      details: details
    )
  end

  class << self
    def default_message
      'Validation failed. Please check your input.'
    end

    def default_code
      'VALIDATION_FAILED'
    end
  end

  private

  def format_errors(errors)
    case errors
    when ActiveModel::Errors
      errors.messages.transform_values(&:uniq)
    when Hash
      errors
    else
      { base: [errors.to_s] }
    end
  end
end
