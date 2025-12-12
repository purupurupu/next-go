class Users::RegistrationsController < Devise::RegistrationsController
  include ApiResponseFormatter

  respond_to :json

  private

  def respond_with(resource, _opts = {})
    if resource.persisted?
      success_response(
        message: 'Signed up successfully.',
        data: user_data(resource),
        status: :created
      )
    else
      error = ::ValidationError.new(
        "User couldn't be created successfully",
        errors: resource.errors
      )
      render_error_response(error: error, status: :unprocessable_content)
    end
  end
end
