# frozen_string_literal: true

# Error raised when a requested resource is not found
class NotFoundError < ApiError
  def initialize(message = nil, resource: nil, id: nil, details: nil)
    details ||= {}
    details[:resource] = resource if resource
    details[:id] = id if id

    message ||= if resource && id
                  "#{resource.to_s.humanize} with ID '#{id}' not found"
                elsif resource
                  "#{resource.to_s.humanize} not found"
                else
                  self.class.default_message
                end

    super(
      message,
      code: 'RESOURCE_NOT_FOUND',
      status: :not_found,
      details: details
    )
  end

  class << self
    def default_message
      'The requested resource was not found.'
    end

    def default_code
      'RESOURCE_NOT_FOUND'
    end
  end
end
