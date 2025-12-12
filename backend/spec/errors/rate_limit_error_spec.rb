# frozen_string_literal: true

require 'rails_helper'

RSpec.describe RateLimitError do
  describe '#initialize' do
    context 'with default values' do
      subject(:error) { described_class.new }

      it 'uses default message' do
        expect(error.message).to eq('Rate limit exceeded. Please try again later.')
      end

      it 'uses RATE_LIMIT_EXCEEDED code' do
        expect(error.code).to eq('RATE_LIMIT_EXCEEDED')
      end

      it 'has too_many_requests status' do
        expect(error.status).to eq(:too_many_requests)
      end

      it 'has details with remaining count' do
        expect(error.details).to eq({ remaining: 0 })
      end
    end

    context 'with all parameters' do
      subject(:error) do
        described_class.new(
          message,
          limit: limit,
          remaining: remaining,
          reset_at: reset_at,
          details: custom_details
        )
      end

      let(:message) { 'Too many API calls' }
      let(:limit) { 100 }
      let(:remaining) { 0 }
      let(:reset_at) { Time.parse('2024-01-15 12:00:00 UTC') }
      let(:custom_details) { { endpoint: '/api/todos' } }

      it 'uses custom message' do
        expect(error.message).to eq(message)
      end

      it 'includes all rate limit details' do
        expect(error.details).to include({
                                           limit: 100,
                                           remaining: 0,
                                           reset_at: '2024-01-15T12:00:00Z',
                                           endpoint: '/api/todos'
                                         })
      end
    end

    context 'with partial parameters' do
      subject(:error) { described_class.new(limit: 50) }

      it 'includes only provided parameters' do
        expect(error.details).to eq({
                                      limit: 50,
                                      remaining: 0
                                    })
      end
    end

    context 'with reset_at as Time object' do
      subject(:error) { described_class.new(reset_at: reset_time) }

      let(:reset_time) { 1.hour.from_now }

      it 'converts reset_at to ISO8601 format' do
        expect(error.details[:reset_at]).to eq(reset_time.iso8601)
      end
    end

    context 'without reset_at' do
      subject(:error) { described_class.new(limit: 100) }

      it 'does not include reset_at in details' do
        expect(error.details).not_to have_key(:reset_at)
      end
    end
  end

  describe '.default_message' do
    it 'returns the default message' do
      expect(described_class.default_message).to eq('Rate limit exceeded. Please try again later.')
    end
  end

  describe '.default_code' do
    it 'returns the default code' do
      expect(described_class.default_code).to eq('RATE_LIMIT_EXCEEDED')
    end
  end

  describe 'inheritance' do
    it 'inherits from ApiError' do
      expect(described_class.superclass).to eq(ApiError)
    end
  end

  describe 'HTTP status' do
    it 'returns 429 Too Many Requests status' do
      error = described_class.new
      expect(Rack::Utils.status_code(error.status)).to eq(429)
    end
  end
end
