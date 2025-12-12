# frozen_string_literal: true

module Mcp
  class Server
    attr_reader :mcp_server

    def initialize
      # MCPサーバーインスタンスを作成
      @mcp_server = ::MCP::Server.new(
        name: 'todo-mcp-server',
        version: '1.0.0',
        tools: registered_tools
      )
    end

    def handle_request(request_body)
      # JSON-RPCリクエストを処理
      Rails.logger.info("MCP Request: #{request_body.truncate(200)}")

      response = @mcp_server.handle_json(request_body)

      Rails.logger.info("MCP Response: #{response.to_json.truncate(200)}")

      response
    rescue StandardError => e
      Rails.logger.error("MCP Server Error: #{e.message}")
      Rails.logger.error(e.backtrace.join("\n"))
      raise
    end

    private

    def registered_tools
      [
        Mcp::Tools::SearchTodos
      ]
    end
  end
end
