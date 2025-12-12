# frozen_string_literal: true

RSpec.shared_examples 'api responses' do
  it 'sets X-API-Version header' do
    expect(response.headers['X-API-Version']).to eq('v1')
  end

  it 'sets proper content type' do
    expect(response.content_type).to match(%r{application/json})
  end

  it 'includes request ID in response' do
    expect(response.headers).to have_key('X-Request-Id')
  end
end

RSpec.shared_examples 'standard success response' do
  it 'returns standard success format' do
    json = response.parsed_body
    expect(json['status']).to include(
      'code' => response.status,
      'message' => be_a(String)
    )
    # Note: 'data' key may not be present if the data is empty
    expect(json).not_to have_key('error')
  end

  it_behaves_like 'api responses'
end

RSpec.shared_examples 'standard error response' do |error_code|
  it 'returns standard error format' do
    json = response.parsed_body
    expect(json['error']).to include(
      'code' => error_code,
      'message' => be_a(String)
    )
    expect(json['error']).to have_key('request_id')
    expect(json['error']).to have_key('timestamp')
    expect(json).not_to have_key('status')
    expect(json).not_to have_key('data')
  end

  it_behaves_like 'api responses'
end

RSpec.shared_examples 'requires authentication' do
  it 'returns unauthorized without auth headers' do
    expect(response).to have_http_status(:unauthorized)
  end
end