# Sidekiq configuration for background job processing
# This configuration sets up Redis connection for job queues

Sidekiq.configure_server do |config|
  config.redis = { url: ENV.fetch('REDIS_URL', 'redis://localhost:6379/0') }
  
  # Disable Sidekiq logging in test environment
  config.logger.level = Logger::WARN if Rails.env.test?
end

Sidekiq.configure_client do |config|
  config.redis = { url: ENV.fetch('REDIS_URL', 'redis://localhost:6379/0') }
  
  # Disable Sidekiq logging in test environment
  config.logger.level = Logger::WARN if Rails.env.test?
end