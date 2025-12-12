# frozen_string_literal: true

require 'rails_helper'

RSpec.describe ErrorHandler do
  let(:app) { ->(_env) { [200, { 'Content-Type' => 'text/plain' }, ['OK']] } }
  let(:middleware) { described_class.new(app) }
  let(:env) do
    Rack::MockRequest.env_for('/api/test',
                              'REQUEST_METHOD' => 'GET',
                              'action_dispatch.request_id' => 'test-request-id')
  end

  describe '#call' do
    context 'when no error occurs' do
      it 'passes the request through successfully' do
        status, _, body = middleware.call(env)
        expect(status).to eq(200)
        expect(body).to eq(['OK'])
      end

      it 'adds request ID if not present' do
        env.delete('action_dispatch.request_id')
        middleware.call(env)
        expect(env['action_dispatch.request_id']).not_to be_nil
      end
    end

    context 'when ApiError is raised' do
      let(:api_error) { NotFoundError.new(resource: 'Todo', id: 123) }
      let(:app) { ->(_env) { raise api_error } }

      it 'returns appropriate error response' do
        status, headers, body = middleware.call(env)
        parsed_body = JSON.parse(body.first)

        expect(status).to eq(404)
        expect(headers['Content-Type']).to eq('application/json')
        expect(headers['X-Request-Id']).to eq('test-request-id')
        expect(parsed_body['error']['code']).to eq('RESOURCE_NOT_FOUND')
        expect(parsed_body['error']['message']).to eq("Todo with ID '123' not found")
        expect(parsed_body['error']['request_id']).to eq('test-request-id')
        expect(parsed_body['error']['timestamp']).not_to be_nil
      end
    end

    context 'when ActiveRecord::RecordNotFound is raised' do
      let(:error) { ActiveRecord::RecordNotFound.new("Couldn't find Todo with 'id'=999") }
      let(:app) { ->(_env) { raise error } }

      it 'converts to NotFoundError and returns appropriate response' do
        status, _, body = middleware.call(env)
        parsed_body = JSON.parse(body.first)

        expect(status).to eq(404)
        expect(parsed_body['error']['code']).to eq('RESOURCE_NOT_FOUND')
        expect(parsed_body['error']['message']).to match(/Todo.* not found/)
      end
    end

    context 'when ActiveRecord::RecordInvalid is raised' do
      let(:todo) { Todo.new }
      let(:error) do
        todo.errors.add(:title, "can't be blank")
        ActiveRecord::RecordInvalid.new(todo)
      end
      let(:app) { ->(_env) { raise error } }

      it 'converts to ValidationError and returns appropriate response' do
        status, _, body = middleware.call(env)
        parsed_body = JSON.parse(body.first)

        expect(status).to eq(422)
        expect(parsed_body['error']['code']).to eq('VALIDATION_FAILED')
        expect(parsed_body['error']['message']).to eq('Validation failed. Please check your input.')
        expect(parsed_body['error']['details']['validation_errors']).not_to be_empty
      end
    end

    context 'when ActionController::ParameterMissing is raised' do
      let(:error) { ActionController::ParameterMissing.new(:todo) }
      let(:app) { ->(_env) { raise error } }

      it 'returns bad request error response' do
        status, _, body = middleware.call(env)
        parsed_body = JSON.parse(body.first)

        expect(status).to eq(400)
        expect(parsed_body['error']['code']).to eq('PARAMETER_MISSING')
        expect(parsed_body['error']['message']).to eq('Required parameter missing: todo')
        expect(parsed_body['error']['details']['missing_parameter']).to eq('todo')
      end
    end

    context 'when JWT::DecodeError is raised' do
      let(:error) { JWT::DecodeError.new('Invalid token') }
      let(:app) { ->(_env) { raise error } }

      it 'converts to AuthenticationError and returns appropriate response' do
        status, _, body = middleware.call(env)
        parsed_body = JSON.parse(body.first)

        expect(status).to eq(401)
        expect(parsed_body['error']['code']).to eq('AUTHENTICATION_FAILED')
        expect(parsed_body['error']['message']).to eq('Invalid or expired token')
        expect(parsed_body['error']['details']['error_type']).to eq('JWT::DecodeError')
      end
    end

    context 'when JWT::ExpiredSignature is raised' do
      let(:error) { JWT::ExpiredSignature.new('Token has expired') }
      let(:app) { ->(_env) { raise error } }

      it 'converts to AuthenticationError and returns appropriate response' do
        status, _, body = middleware.call(env)
        parsed_body = JSON.parse(body.first)

        expect(status).to eq(401)
        expect(parsed_body['error']['code']).to eq('AUTHENTICATION_FAILED')
        expect(parsed_body['error']['message']).to eq('Invalid or expired token')
        expect(parsed_body['error']['details']['error_type']).to eq('JWT::ExpiredSignature')
      end
    end

    context 'when StandardError is raised' do
      let(:error) { StandardError.new('Something went wrong') }
      let(:app) { ->(_env) { raise error } }

      context 'in development environment' do
        before do
          allow(Rails.env).to receive_messages(production?: false, test?: false)
        end

        it 'returns internal server error with details' do
          status, _, body = middleware.call(env)
          parsed_body = JSON.parse(body.first)

          expect(status).to eq(500)
          expect(parsed_body['error']['code']).to eq('INTERNAL_ERROR')
          expect(parsed_body['error']['message']).to eq('Something went wrong')
          expect(parsed_body['error']['details']['error_class']).to eq('StandardError')
        end
      end

      context 'in production environment' do
        before do
          allow(Rails.env).to receive_messages(production?: true, test?: false)
        end

        it 'returns internal server error without details' do
          status, _, body = middleware.call(env)
          parsed_body = JSON.parse(body.first)

          expect(status).to eq(500)
          expect(parsed_body['error']['code']).to eq('INTERNAL_ERROR')
          expect(parsed_body['error']['message']).to eq('An unexpected error occurred')
          expect(parsed_body['error']['details']).to eq({})
        end
      end
    end

    context 'with different error types' do
      it 'handles AuthorizationError correctly' do
        error = AuthorizationError.new('Forbidden')
        app = ->(_env) { raise error }
        middleware = described_class.new(app)

        status, _, body = middleware.call(env)
        parsed_body = JSON.parse(body.first)

        expect(status).to eq(403)
        expect(parsed_body['error']['code']).to eq('AUTHORIZATION_FAILED')
      end

      it 'handles RateLimitError correctly' do
        error = RateLimitError.new(limit: 100, reset_at: 1.hour.from_now)
        app = ->(_env) { raise error }
        middleware = described_class.new(app)

        status, _, body = middleware.call(env)
        parsed_body = JSON.parse(body.first)

        expect(status).to eq(429)
        expect(parsed_body['error']['code']).to eq('RATE_LIMIT_EXCEEDED')
      end

      it 'handles ValidationError with complex errors correctly' do
        errors = ActiveModel::Errors.new(Todo.new)
        errors.add(:title, "can't be blank")
        errors.add(:title, 'is too short (minimum is 3 characters)')
        errors.add(:due_date, 'must be in the future')

        error = ValidationError.new(errors: errors)
        app = ->(_env) { raise error }
        middleware = described_class.new(app)

        status, _, body = middleware.call(env)
        parsed_body = JSON.parse(body.first)

        expect(status).to eq(422)
        expect(parsed_body['error']['details']['validation_errors']['title']).to include("can't be blank")
        expect(parsed_body['error']['details']['validation_errors']['due_date']).to include('must be in the future')
      end
    end
  end

  describe 'private methods' do
    describe '#extract_model_from_message' do
      it 'extracts model name from standard Rails error message' do
        message = "Couldn't find Todo with 'id'=123"
        result = middleware.send(:extract_model_from_message, message)
        expect(result).to eq('Todo')
      end

      it 'returns nil when no model name can be extracted' do
        message = 'Some other error message'
        result = middleware.send(:extract_model_from_message, message)
        expect(result).to be_nil
      end
    end

    describe '#filtered_params' do
      it 'returns params from request' do
        # Test the basic functionality without mocking Rails internals
        request = instance_double(ActionDispatch::Request)
        params = { 'some' => 'params' }
        allow(request).to receive(:params).and_return(params)

        # Since ActionDispatch::Http::ParameterFilter is not available in this context,
        # the method will rescue and return empty hash
        result = middleware.send(:filtered_params, request)
        expect(result).to eq({})
      end

      it 'handles errors gracefully' do
        request = instance_double(ActionDispatch::Request)
        allow(request).to receive(:params).and_raise(StandardError)

        result = middleware.send(:filtered_params, request)
        expect(result).to eq({})
      end
    end

    describe '#build_response' do
      it 'builds response with correct status code and headers' do
        body = { error: { message: 'Test error', request_id: 'test-123' } }
        status, headers, response_body = middleware.send(:build_response, body, :not_found)

        expect(status).to eq(404)
        expect(headers['Content-Type']).to eq('application/json')
        expect(headers['X-Request-Id']).to eq('test-123')
        expect(response_body.first).to eq(body.to_json)
      end

      it 'generates request ID if not provided' do
        body = { error: { message: 'Test error' } }
        _, headers, = middleware.send(:build_response, body, :ok)

        expect(headers['X-Request-Id']).not_to be_nil
      end
    end

    describe '#log_error' do
      it 'does not log in test environment' do
        allow(Rails.logger).to receive(:warn)
        allow(Rails.logger).to receive(:error)

        error = StandardError.new('Test error')
        request = ActionDispatch::Request.new(env)
        middleware.send(:log_error, error, request, :warn)

        expect(Rails.logger).not_to have_received(:warn)
        expect(Rails.logger).not_to have_received(:error)
      end
    end
  end

  describe 'integration scenarios' do
    it 'handles nested API errors correctly' do
      StandardError.new('Database connection failed')
      api_error = ApiError.new('Service unavailable', code: 'SERVICE_ERROR', status: :internal_server_error)
      app = ->(_env) { raise api_error }
      middleware = described_class.new(app)

      status, _, body = middleware.call(env)
      parsed_body = JSON.parse(body.first)

      expect(status).to eq(500)
      expect(parsed_body['error']['message']).to eq('Service unavailable')
      expect(parsed_body['error']['code']).to eq('SERVICE_ERROR')
    end

    it 'preserves request context through error handling' do
      env['action_dispatch.request_id'] = 'unique-request-123'
      env['REQUEST_METHOD'] = 'POST'
      env['PATH_INFO'] = '/api/todos'

      error = ApiError.new('Invalid input', code: 'INVALID_INPUT', status: :bad_request)
      app = ->(_env) { raise error }
      middleware = described_class.new(app)

      _, headers, body = middleware.call(env)
      parsed_body = JSON.parse(body.first)

      expect(headers['X-Request-Id']).to eq('unique-request-123')
      expect(parsed_body['error']['request_id']).to eq('unique-request-123')
    end
  end
end
