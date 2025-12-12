# frozen_string_literal: true

# FactoryBot Profiler
# This module helps identify slow factories by tracking their execution time
module FactoryBotProfiler
  class << self
    attr_accessor :enabled
    attr_reader :results

    def reset!
      @results = {}
      @enabled = false
    end

    def enable!
      @enabled = true
      reset_results!
    end

    def disable!
      @enabled = false
    end

    def reset_results!
      @results = {}
    end

    def track(factory_name, strategy)
      return yield unless enabled

      start_time = Process.clock_gettime(Process::CLOCK_MONOTONIC)
      result = yield
      end_time = Process.clock_gettime(Process::CLOCK_MONOTONIC)

      duration = (end_time - start_time) * 1000 # Convert to milliseconds

      key = "#{factory_name}##{strategy}"
      @results[key] ||= { count: 0, total_time: 0.0, strategy: strategy, factory: factory_name }
      @results[key][:count] += 1
      @results[key][:total_time] += duration

      result
    end

    def report
      return Rails.logger.debug 'FactoryBot profiling is disabled. Enable it with FactoryBotProfiler.enable!' unless results.any?

      Rails.logger.debug { "\n#{'=' * 80}" }
      Rails.logger.debug 'FactoryBot Performance Report'
      Rails.logger.debug '=' * 80

      sorted_results = results.values.sort_by { |r| -r[:total_time] }

      Rails.logger.debug 'Factory                        Strategy        Count      Total (ms)        Avg (ms)'
      Rails.logger.debug '-' * 80

      sorted_results.each do |result|
        avg_time = result[:total_time] / result[:count]
        Rails.logger.debug format('%-30s %-10s %10d %15.2f %15.2f',
                                  result[:factory],
                                  result[:strategy],
                                  result[:count],
                                  result[:total_time],
                                  avg_time)
      end

      total_time = sorted_results.sum { |r| r[:total_time] }
      total_count = sorted_results.sum { |r| r[:count] }

      Rails.logger.debug '-' * 80
      Rails.logger.debug format('%-30s %-10s %10d %15.2f', 'TOTAL', '', total_count, total_time)
      Rails.logger.debug '=' * 80

      # Identify problematic factories
      slow_factories = sorted_results.select { |r| r[:total_time] / r[:count] > 50 } # > 50ms average

      if slow_factories.any?
        Rails.logger.debug "\nWARNING: The following factories are slow (>50ms average):"
        slow_factories.each do |result|
          avg_time = result[:total_time] / result[:count]
          Rails.logger.debug { "  - #{result[:factory]} (#{result[:strategy]}): #{avg_time.round(2)}ms average" }
        end
      end

      Rails.logger.debug "\n"
    end
  end

  reset!
end

# Monkey patch FactoryBot to add profiling
module FactoryBot
  class << self
    # Store original methods
    alias original_create create
    alias original_build build
    alias original_build_stubbed build_stubbed
    alias original_attributes_for attributes_for

    def create(name, *traits_and_overrides, &block)
      FactoryBotProfiler.track(name, :create) do
        original_create(name, *traits_and_overrides, &block)
      end
    end

    def build(name, *traits_and_overrides, &block)
      FactoryBotProfiler.track(name, :build) do
        original_build(name, *traits_and_overrides, &block)
      end
    end

    def build_stubbed(name, *traits_and_overrides, &block)
      FactoryBotProfiler.track(name, :build_stubbed) do
        original_build_stubbed(name, *traits_and_overrides, &block)
      end
    end

    def attributes_for(name, *traits_and_overrides, &block)
      FactoryBotProfiler.track(name, :attributes_for) do
        original_attributes_for(name, *traits_and_overrides, &block)
      end
    end
  end
end

# RSpec configuration
if defined?(RSpec)
  RSpec.configure do |config|
    config.before(:suite) do
      if ENV['PROFILE_FACTORIES']
        FactoryBotProfiler.enable!
        Rails.logger.debug 'FactoryBot profiling enabled'
      end
    end

    config.after(:suite) do
      FactoryBotProfiler.report if FactoryBotProfiler.enabled
    end
  end
end
