namespace :coverage do
  desc 'Run RSpec tests with SimpleCov coverage'
  task rspec: :environment do
    ENV['COVERAGE'] = 'true'
    ENV['RAILS_ENV'] = 'test'

    puts 'Running RSpec with SimpleCov coverage...'
    puts 'Coverage report will be generated in ./coverage/index.html'

    # Run RSpec with coverage
    sh 'bundle exec rspec' do |ok, res|
      if ok
        puts "\n✅ Tests passed! Check coverage report at ./coverage/index.html"
      else
        puts "\n❌ Tests failed with exit code #{res.exitstatus}"
        exit res.exitstatus
      end
    end
  end

  desc 'Run RSpec tests with SimpleCov coverage and open report'
  task rspec_open: :rspec do
    if File.exist?('coverage/index.html')
      case RUBY_PLATFORM
      when /darwin/
        sh 'open coverage/index.html'
      when /linux/
        sh 'xdg-open coverage/index.html'
      else
        puts 'Please open coverage/index.html manually'
      end
    end
  end

  desc 'Clean coverage reports'
  task clean: :environment do
    puts 'Cleaning coverage reports...'
    sh 'rm -rf coverage/'
  end
end

# Alias for convenience
desc 'Run tests with coverage (alias for coverage:rspec)'
task coverage: 'coverage:rspec'
