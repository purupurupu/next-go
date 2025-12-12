# Authentication helpers for RSpec tests
# These helpers support both controller specs (using Devise helpers) and request specs (using JWT)

module AuthenticationHelpers
  # Cache for JWT tokens to avoid multiple login requests per user in a single test
  def jwt_tokens
    @jwt_tokens ||= {}
  end

  # Controller spec authentication using Devise test helpers
  def sign_in_user(user)
    unless defined?(sign_in)
      raise 'sign_in method not available. Make sure Devise test helpers are included for controller specs.'
    end

    sign_in user
  end

  # Helper method to generate JWT token for testing (cached per user)
  def jwt_token_for(user)
    return jwt_tokens[user.id] if jwt_tokens[user.id]

    # Ensure user has correct password for authentication
    user.password = 'password123' if user.encrypted_password.blank?
    user.password_confirmation = 'password123' if user.encrypted_password.blank?
    user.save! if user.changed?

    # Create a new request context for login
    post '/auth/sign_in',
         params: { user: { email: user.email, password: 'password123' } },
         as: :json,
         headers: { 'Content-Type' => 'application/json', 'Host' => 'localhost:3001' }

    raise "Failed to authenticate user #{user.email}: #{response.status} - #{response.body}" unless response.successful?

    token = response.headers['Authorization']
    raise "No Authorization header returned for user #{user.email}. Response: #{response.body}" unless token

    jwt_tokens[user.id] = token
    token
  end

  # Helper method to set authorization headers for request specs
  def auth_headers_for(user)
    # Always get a fresh token to avoid caching issues
    jwt_tokens.delete(user.id) if jwt_tokens[user.id]
    token = jwt_token_for(user)
    {
      'Authorization' => token,
      'Content-Type' => 'application/json',
      'Host' => 'localhost:3001'
    }
  end

  # Clear JWT token cache (useful for before blocks)
  def clear_auth_cache
    @jwt_tokens = {}
  end

  # Debug helper to inspect authentication state
  def debug_auth_for(user)
    Rails.logger.debug { "=== Auth Debug for #{user.email} ===" }
    Rails.logger.debug { "User ID: #{user.id}" }
    Rails.logger.debug { "User valid: #{user.valid?}" }
    Rails.logger.debug { "User errors: #{user.errors.full_messages}" } unless user.valid?

    if jwt_tokens[user.id]
      Rails.logger.debug { "Cached token exists: #{jwt_tokens[user.id][0..20]}..." }
    else
      Rails.logger.debug 'No cached token'
    end
    Rails.logger.debug '=================================='
  end
end
