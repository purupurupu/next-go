# frozen_string_literal: true

# Error raised when business logic constraints are violated
class BusinessLogicError < ApiError
  def initialize(message = nil, code: nil, details: nil)
    super(
      message,
      code: code || 'BUSINESS_LOGIC_ERROR',
      status: :unprocessable_content,
      details: details
    )
  end

  class << self
    def default_message
      'Business logic constraint violated.'
    end

    def default_code
      'BUSINESS_LOGIC_ERROR'
    end
  end

  # Specific business logic errors
  class DuplicateResourceError < BusinessLogicError
    def initialize(message = nil, resource: nil, field: nil, value: nil)
      details = {}
      details[:resource] = resource if resource
      details[:field] = field if field
      details[:value] = value if value

      message ||= if resource && field
                    "#{resource.to_s.humanize} with #{field} '#{value}' already exists"
                  else
                    'Resource already exists'
                  end

      super(message, code: 'DUPLICATE_RESOURCE', details: details)
    end
  end

  class InvalidStateTransitionError < BusinessLogicError
    def initialize(message = nil, from_state: nil, to_state: nil, allowed_states: nil)
      details = {}
      details[:from_state] = from_state if from_state
      details[:to_state] = to_state if to_state
      details[:allowed_states] = allowed_states if allowed_states

      message ||= if from_state && to_state
                    "Cannot transition from '#{from_state}' to '#{to_state}'"
                  else
                    'Invalid state transition'
                  end

      super(message, code: 'INVALID_STATE_TRANSITION', details: details)
    end
  end

  class ResourceLimitExceededError < BusinessLogicError
    def initialize(message = nil, resource: nil, limit: nil, current: nil)
      details = {}
      details[:resource] = resource if resource
      details[:limit] = limit if limit
      details[:current] = current if current

      message ||= if resource && limit
                    "Maximum limit of #{limit} #{resource.to_s.pluralize} exceeded"
                  else
                    'Resource limit exceeded'
                  end

      super(message, code: 'RESOURCE_LIMIT_EXCEEDED', details: details)
    end
  end
end
