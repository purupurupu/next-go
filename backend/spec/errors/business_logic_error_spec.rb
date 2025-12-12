# frozen_string_literal: true

require 'rails_helper'

RSpec.describe BusinessLogicError do
  describe '#initialize' do
    context 'with default values' do
      subject(:error) { described_class.new }

      it 'uses default message' do
        expect(error.message).to eq('Business logic constraint violated.')
      end

      it 'uses default code' do
        expect(error.code).to eq('BUSINESS_LOGIC_ERROR')
      end

      it 'has unprocessable_entity status' do
        expect(error.status).to eq(:unprocessable_content)
      end

      it 'has empty details' do
        expect(error.details).to eq({})
      end
    end

    context 'with custom values' do
      subject(:error) { described_class.new(message, code: code, details: details) }

      let(:message) { 'Custom business error' }
      let(:code) { 'CUSTOM_ERROR' }
      let(:details) { { field: 'value' } }

      it 'uses custom message' do
        expect(error.message).to eq(message)
      end

      it 'uses custom code' do
        expect(error.code).to eq(code)
      end

      it 'uses custom details' do
        expect(error.details).to eq(details)
      end
    end
  end

  describe '.default_message' do
    it 'returns the default message' do
      expect(described_class.default_message).to eq('Business logic constraint violated.')
    end
  end

  describe '.default_code' do
    it 'returns the default code' do
      expect(described_class.default_code).to eq('BUSINESS_LOGIC_ERROR')
    end
  end
end

RSpec.describe BusinessLogicError::DuplicateResourceError do
  describe '#initialize' do
    context 'with all parameters' do
      subject(:error) do
        described_class.new(resource: resource, field: field, value: value)
      end

      let(:resource) { 'user' }
      let(:field) { 'email' }
      let(:value) { 'test@example.com' }

      it 'generates appropriate message' do
        expect(error.message).to eq("User with email 'test@example.com' already exists")
      end

      it 'uses DUPLICATE_RESOURCE code' do
        expect(error.code).to eq('DUPLICATE_RESOURCE')
      end

      it 'includes all details' do
        expect(error.details).to eq({
                                      resource: 'user',
                                      field: 'email',
                                      value: 'test@example.com'
                                    })
      end
    end

    context 'with custom message' do
      subject(:error) { described_class.new(message, resource: 'user', field: 'email') }

      let(:message) { 'This email is taken' }

      it 'uses custom message' do
        expect(error.message).to eq(message)
      end
    end

    context 'with minimal parameters' do
      subject(:error) { described_class.new }

      it 'uses default message' do
        expect(error.message).to eq('Resource already exists')
      end

      it 'has empty details' do
        expect(error.details).to eq({})
      end
    end

    context 'with partial parameters' do
      subject(:error) { described_class.new(resource: 'todo') }

      it 'includes only provided details' do
        expect(error.details).to eq({ resource: 'todo' })
      end
    end
  end
end

RSpec.describe BusinessLogicError::InvalidStateTransitionError do
  describe '#initialize' do
    context 'with all parameters' do
      subject(:error) do
        described_class.new(
          from_state: from_state,
          to_state: to_state,
          allowed_states: allowed_states
        )
      end

      let(:from_state) { 'pending' }
      let(:to_state) { 'completed' }
      let(:allowed_states) { %w[in_progress cancelled] }

      it 'generates appropriate message' do
        expect(error.message).to eq("Cannot transition from 'pending' to 'completed'")
      end

      it 'uses INVALID_STATE_TRANSITION code' do
        expect(error.code).to eq('INVALID_STATE_TRANSITION')
      end

      it 'includes all details' do
        expect(error.details).to eq({
                                      from_state: 'pending',
                                      to_state: 'completed',
                                      allowed_states: %w[in_progress cancelled]
                                    })
      end
    end

    context 'with custom message' do
      subject(:error) { described_class.new(message, from_state: 'draft', to_state: 'published') }

      let(:message) { 'Invalid workflow transition' }

      it 'uses custom message' do
        expect(error.message).to eq(message)
      end
    end

    context 'with minimal parameters' do
      subject(:error) { described_class.new }

      it 'uses default message' do
        expect(error.message).to eq('Invalid state transition')
      end

      it 'has empty details' do
        expect(error.details).to eq({})
      end
    end
  end
end

RSpec.describe BusinessLogicError::ResourceLimitExceededError do
  describe '#initialize' do
    context 'with all parameters' do
      subject(:error) do
        described_class.new(
          resource: resource,
          limit: limit,
          current: current
        )
      end

      let(:resource) { 'todo' }
      let(:limit) { 100 }
      let(:current) { 105 }

      it 'generates appropriate message with pluralization' do
        expect(error.message).to eq('Maximum limit of 100 todos exceeded')
      end

      it 'uses RESOURCE_LIMIT_EXCEEDED code' do
        expect(error.code).to eq('RESOURCE_LIMIT_EXCEEDED')
      end

      it 'includes all details' do
        expect(error.details).to eq({
                                      resource: 'todo',
                                      limit: 100,
                                      current: 105
                                    })
      end
    end

    context 'with custom message' do
      subject(:error) { described_class.new(message, resource: 'project', limit: 5) }

      let(:message) { 'You have reached your plan limit' }

      it 'uses custom message' do
        expect(error.message).to eq(message)
      end
    end

    context 'with minimal parameters' do
      subject(:error) { described_class.new }

      it 'uses default message' do
        expect(error.message).to eq('Resource limit exceeded')
      end

      it 'has empty details' do
        expect(error.details).to eq({})
      end
    end

    context 'with singular resource' do
      subject(:error) { described_class.new(resource: 'category', limit: 1) }

      it 'pluralizes correctly' do
        expect(error.message).to eq('Maximum limit of 1 categories exceeded')
      end
    end
  end

  describe 'inheritance' do
    it 'inherits from BusinessLogicError' do
      expect(described_class.superclass).to eq(BusinessLogicError)
    end

    it 'has unprocessable_entity status' do
      error = described_class.new
      expect(error.status).to eq(:unprocessable_content)
    end
  end
end
