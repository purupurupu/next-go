# Be sure to restart your server when you modify this file.

# Avoid CORS issues when API is called from the frontend app.
# Handle Cross-Origin Resource Sharing (CORS) in order to accept cross-origin Ajax requests.

# Read more: https://github.com/cyu/rack-cors

Rails.application.config.middleware.insert_before 0, Rack::Cors do
  # フロントエンド用のCORS設定
  allow do
    origins 'http://localhost:3000'

    resource '*',
      headers: :any,
      methods: [:get, :post, :put, :patch, :delete, :options, :head],
      credentials: true,
      expose: ['Authorization']
  end

  # MCP エンドポイント用のCORS設定
  # Claude DesktopなどのMCPクライアントからのアクセスを許可
  allow do
    origins '*'

    resource '/mcp',
      headers: :any,
      methods: [:post, :options],
      credentials: false
  end
end