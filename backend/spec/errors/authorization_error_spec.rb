# frozen_string_literal: true

require 'rails_helper'

RSpec.describe AuthorizationError do
  describe '#initialize' do
    context 'with default values' do
      subject(:error) { described_class.new }

      it 'uses default message' do
        expect(error.message).to eq('You are not authorized to perform this action.')
      end

      it 'uses AUTHORIZATION_FAILED code' do
        expect(error.code).to eq('AUTHORIZATION_FAILED')
      end

      it 'has forbidden status' do
        expect(error.status).to eq(:forbidden)
      end

      it 'has empty details' do
        expect(error.details).to eq({})
      end
    end

    context 'with custom message and details' do
      subject(:error) { described_class.new(message, resource: resource, action: action) }

      let(:message) { 'Admin access required' }
      let(:resource) { 'admin_dashboard' }
      let(:action) { 'view' }

      it 'uses custom message' do
        expect(error.message).to eq(message)
      end

      it 'includes resource and action in details' do
        expect(error.details).to eq({
                                      resource: 'admin_dashboard',
                                      action: 'view'
                                    })
      end
    end

    context 'with only resource' do
      subject(:error) { described_class.new(resource: 'project') }

      it 'includes only resource in details' do
        expect(error.details).to eq({ resource: 'project' })
      end
    end

    context 'with only action' do
      subject(:error) { described_class.new(action: 'delete') }

      it 'includes only action in details' do
        expect(error.details).to eq({ action: 'delete' })
      end
    end
  end

  describe '.default_message' do
    it 'returns the default message' do
      expect(described_class.default_message).to eq('You are not authorized to perform this action.')
    end
  end

  describe '.default_code' do
    it 'returns the default code' do
      expect(described_class.default_code).to eq('AUTHORIZATION_FAILED')
    end
  end

  describe 'inheritance' do
    it 'inherits from ApiError' do
      expect(described_class.superclass).to eq(ApiError)
    end
  end
end
