# This file is copied to spec/ when you run 'rails generate rspec:install'
require 'spec_helper'
ENV['RAILS_ENV'] ||= 'test'
require_relative '../config/environment'
# Prevent database truncation if the environment is production
abort("The Rails environment is running in production mode!") if Rails.env.production?
# Uncomment the line below in case you have `--require rails_helper` in the `.rspec` file
# that will avoid rails generators crashing because migrations haven't been run yet
# return unless Rails.env.test?
require 'rspec/rails'
# Add additional requires below this line. Rails is not loaded until this point!
require 'factory_bot_rails'
require 'database_cleaner/active_record'

# Include file upload helpers
include ActionDispatch::TestProcess::FixtureFile

# Configure host authorization to allow all hosts for tests
Rails.application.config.host_authorization = { exclude: ->(request) { true } }

# Configure Active Job to use inline adapter for tests to avoid Redis dependency
Rails.application.config.active_job.queue_adapter = :inline

# Disable SQL logging in tests for cleaner output
ActiveRecord::Base.logger.level = Logger::WARN if defined?(ActiveRecord::Base)
Rails.logger.level = Logger::WARN
ActiveJob::Base.logger = Logger.new(nil)
ActiveStorage.logger = Logger.new(nil) if defined?(ActiveStorage)

# Disable ActiveModelSerializers logging
ActiveModelSerializers.logger.level = Logger::WARN if defined?(ActiveModelSerializers)

# Requires supporting ruby files with custom matchers and macros, etc, in
# spec/support/ and its subdirectories. Files matching `spec/**/*_spec.rb` are
# run as spec files by default. This means that files in spec/support that end
# in _spec.rb will both be required and run as specs, causing the specs to be
# run twice. It is recommended that you do not name files matching this glob to
# end with _spec.rb. You can configure this pattern with the --pattern
# option on the command line or in ~/.rspec, .rspec or `.rspec-local`.
#
# The following line is provided for convenience purposes. It has the downside
# of increasing the boot-up time by auto-requiring all files in the support
# directory. Alternatively, in the individual `*_spec.rb` files, manually
# require only the support files necessary.
#
Rails.root.glob('spec/support/**/*.rb').sort_by(&:to_s).each { |f| require f }

# Checks for pending migrations and applies them before tests are run.
# If you are not using ActiveRecord, you can remove these lines.
begin
  ActiveRecord::Migration.maintain_test_schema!
rescue ActiveRecord::PendingMigrationError => e
  abort e.to_s.strip
end
RSpec.configure do |config|
  # Remove this line if you're not using ActiveRecord or ActiveRecord fixtures
  config.fixture_paths = [
    Rails.root.join('spec/fixtures')
  ]

  # If you're not using ActiveRecord, or you'd prefer not to run each of your
  # examples within a transaction, remove the following line or assign false
  # instead of true.
  config.use_transactional_fixtures = true
  
  # Set the host for request specs to avoid host authorization warnings
  config.before(:each, type: :request) do
    host! 'localhost:3001'
  end

  # You can uncomment this line to turn off ActiveRecord support entirely.
  # config.use_active_record = false

  # RSpec Rails uses metadata to mix in different behaviours to your tests,
  # for example enabling you to call `get` and `post` in request specs. e.g.:
  #
  #     RSpec.describe UsersController, type: :request do
  #       # ...
  #     end
  #
  # The different available types are documented in the features, such as in
  # https://rspec.info/features/7-1/rspec-rails
  #
  # You can also this infer these behaviours automatically by location, e.g.
  # /spec/models would pull in the same behaviour as `type: :model` but this
  # behaviour is considered legacy and will be removed in a future version.
  #
  # To enable this behaviour uncomment the line below.
  config.infer_spec_type_from_file_location!

  # Filter lines from Rails gems in backtraces.
  config.filter_rails_from_backtrace!
  # arbitrary gems may also be filtered via:
  # config.filter_gems_from_backtrace("gem name")
  
  # Factory Bot configuration
  config.include FactoryBot::Syntax::Methods
  
  # Devise test helpers for controller specs
  config.include Devise::Test::ControllerHelpers, type: :controller
  
  # Authentication helpers for both request and controller specs
  config.include AuthenticationHelpers, type: :request
  config.include AuthenticationHelpers, type: :controller
  
  # Clear authentication cache before each spec
  config.before(:each, type: :request) do
    clear_auth_cache if respond_to?(:clear_auth_cache)
  end
  
  config.before(:each, type: :controller) do
    clear_auth_cache if respond_to?(:clear_auth_cache)
  end
  
  # Database Cleaner configuration
  config.before(:suite) do
    DatabaseCleaner.strategy = :transaction
    # Allow remote database URLs for Docker testing
    DatabaseCleaner.allow_remote_database_url = true
    DatabaseCleaner.clean_with(:truncation)
    
    # Suppress Sidekiq logs in tests
    if defined?(Sidekiq)
      Sidekiq.logger.level = Logger::WARN
    end
  end

  config.around(:each) do |example|
    DatabaseCleaner.cleaning do
      example.run
    end
  end
end

# Shoulda Matchers configuration
begin
  require 'shoulda/matchers'
  Shoulda::Matchers.configure do |config|
    config.integrate do |with|
      with.test_framework :rspec
      with.library :rails
    end
  end
rescue LoadError
  # Shoulda Matchers not available
end
