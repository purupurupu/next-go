# frozen_string_literal: true

class McpController < ApplicationController
  # 認証をスキップしてカスタム検証を使用
  skip_before_action :authenticate_user!
  before_action :verify_mcp_token

  # TODO: 本番環境では環境変数から取得すること
  MCP_TOKEN = 'your-secret-mcp-token-here'

  def handle
    # MCPサーバーインスタンスを作成
    mcp_server = Mcp::Server.new

    # リクエストボディを処理
    response = mcp_server.handle_request(request.body.read)

    # JSON-RPCレスポンスを返却
    render json: response, status: :ok
  rescue StandardError => e
    # エラーログを記録
    Rails.logger.error("MCP Controller Error: #{e.message}")
    Rails.logger.error(e.backtrace.join("\n"))

    # JSON-RPC エラーレスポンスを返却
    render json: {
      jsonrpc: '2.0',
      error: {
        code: -32_603,
        message: "Internal error: #{e.message}"
      },
      id: nil
    }, status: :internal_server_error
  end

  private

  def verify_mcp_token
    # Authorizationヘッダーからトークンを取得
    token = request.headers['Authorization']&.gsub(/^Bearer /, '')

    return if token == MCP_TOKEN

    render json: {
      jsonrpc: '2.0',
      error: {
        code: -32_001,
        message: 'Unauthorized: Invalid MCP token'
      },
      id: nil
    }, status: :unauthorized
  end
end
