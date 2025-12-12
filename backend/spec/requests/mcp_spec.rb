require 'rails_helper'

RSpec.describe 'MCP Server', type: :request do
  let(:headers) { { 'Content-Type' => 'application/json', 'Host' => 'localhost:3001' } }
  let(:valid_token) { 'your-secret-mcp-token-here' }
  let(:invalid_token) { 'invalid-token' }

  # テストデータ作成
  let!(:user) { create(:user, email: 'demo@example.com') }
  let!(:category) { create(:category, user: user, name: '学習') }
  # rubocop:disable RSpec/LetSetup
  let!(:todo1) do
    create(:todo,
           user: user,
           title: 'Ruby on Rails学習',
           description: 'Rails 7の新機能について学習する',
           status: :in_progress,
           priority: :high,
           category: category)
  end
  let!(:todo2) do
    create(:todo,
           user: user,
           title: 'テストコード作成',
           description: 'RSpecでテストを書く',
           status: :pending,
           priority: :medium)
  end
  # rubocop:enable RSpec/LetSetup

  describe 'POST /mcp' do
    context 'トークン認証' do
      context 'トークンなし' do
        it '401 Unauthorizedを返す' do
          post '/mcp',
               params: { jsonrpc: '2.0', method: 'tools/list', id: 1 }.to_json,
               headers: headers

          expect(response).to have_http_status(:unauthorized)

          json_response = response.parsed_body
          expect(json_response['jsonrpc']).to eq('2.0')
          expect(json_response['error']['code']).to eq(-32_001)
          expect(json_response['error']['message']).to eq('Unauthorized: Invalid MCP token')
        end
      end

      context '無効なトークン' do
        it '401 Unauthorizedを返す' do
          post '/mcp',
               params: { jsonrpc: '2.0', method: 'tools/list', id: 1 }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{invalid_token}" })

          expect(response).to have_http_status(:unauthorized)

          json_response = response.parsed_body
          expect(json_response['error']['code']).to eq(-32_001)
          expect(json_response['error']['message']).to eq('Unauthorized: Invalid MCP token')
        end
      end

      context '有効なトークン' do
        it 'リクエストを処理する' do
          post '/mcp',
               params: { jsonrpc: '2.0', method: 'tools/list', id: 1 }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

          expect(response).to have_http_status(:ok)

          json_response = response.parsed_body
          expect(json_response['jsonrpc']).to eq('2.0')
          expect(json_response['id']).to eq(1)
        end
      end
    end

    context 'tools/list' do
      it 'ツール一覧を返す' do
        post '/mcp',
             params: { jsonrpc: '2.0', method: 'tools/list', id: 1 }.to_json,
             headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

        expect(response).to have_http_status(:ok)

        json_response = response.parsed_body
        expect(json_response['jsonrpc']).to eq('2.0')
        expect(json_response['id']).to eq(1)
        expect(json_response['result']['tools']).to be_an(Array)
        expect(json_response['result']['tools'].size).to eq(1)

        tool = json_response['result']['tools'].first
        expect(tool['name']).to eq('search_todos')
        expect(tool['description']).to be_present
        expect(tool['inputSchema']).to be_present
        expect(tool['inputSchema']['required']).to eq(['query'])
      end
    end

    context 'tools/call - search_todos' do
      context 'TODO検索成功' do
        it 'タイトルでTODOを検索できる' do
          post '/mcp',
               params: {
                 jsonrpc: '2.0',
                 method: 'tools/call',
                 params: {
                   name: 'search_todos',
                   arguments: { query: '学習' }
                 },
                 id: 1
               }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

          expect(response).to have_http_status(:ok)

          json_response = response.parsed_body
          expect(json_response['jsonrpc']).to eq('2.0')
          expect(json_response['id']).to eq(1)

          content = json_response['result']['content'].first
          expect(content['type']).to eq('text')

          result = JSON.parse(content['text'])
          expect(result['count']).to eq(1)
          expect(result['total_found']).to eq(1)
          expect(result['todos'].first['title']).to eq('Ruby on Rails学習')
          expect(result['todos'].first['user_email']).to eq('demo@example.com')
        end

        it '部分一致でTODOを検索できる' do
          post '/mcp',
               params: {
                 jsonrpc: '2.0',
                 method: 'tools/call',
                 params: {
                   name: 'search_todos',
                   arguments: { query: 'Rails' }
                 },
                 id: 1
               }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

          expect(response).to have_http_status(:ok)

          json_response = response.parsed_body
          content = json_response['result']['content'].first
          result = JSON.parse(content['text'])

          expect(result['count']).to eq(1)
          expect(result['todos'].first['title']).to include('Rails')
        end
      end

      context 'フィルタリング' do
        it 'ステータスでフィルタできる' do
          post '/mcp',
               params: {
                 jsonrpc: '2.0',
                 method: 'tools/call',
                 params: {
                   name: 'search_todos',
                   arguments: { query: '', status: 'in_progress' }
                 },
                 id: 1
               }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

          expect(response).to have_http_status(:ok)

          json_response = response.parsed_body
          content = json_response['result']['content'].first
          result = JSON.parse(content['text'])

          expect(result['count']).to be >= 1
          result['todos'].each do |todo|
            expect(todo['status']).to eq('in_progress')
          end
        end

        it '優先度でフィルタできる' do
          post '/mcp',
               params: {
                 jsonrpc: '2.0',
                 method: 'tools/call',
                 params: {
                   name: 'search_todos',
                   arguments: { query: '', priority: 'high' }
                 },
                 id: 1
               }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

          expect(response).to have_http_status(:ok)

          json_response = response.parsed_body
          content = json_response['result']['content'].first
          result = JSON.parse(content['text'])

          expect(result['count']).to be >= 1
          result['todos'].each do |todo|
            expect(todo['priority']).to eq('high')
          end
        end

        it 'カテゴリーでフィルタできる' do
          post '/mcp',
               params: {
                 jsonrpc: '2.0',
                 method: 'tools/call',
                 params: {
                   name: 'search_todos',
                   arguments: { query: '', category_id: category.id }
                 },
                 id: 1
               }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

          expect(response).to have_http_status(:ok)

          json_response = response.parsed_body
          content = json_response['result']['content'].first
          result = JSON.parse(content['text'])

          expect(result['count']).to be >= 1
          result['todos'].each do |todo|
            expect(todo['category']).to eq('学習')
          end
        end
      end

      context 'リミット制限' do
        it 'limitパラメータで結果数を制限できる' do
          post '/mcp',
               params: {
                 jsonrpc: '2.0',
                 method: 'tools/call',
                 params: {
                   name: 'search_todos',
                   arguments: { query: '', limit: 1 }
                 },
                 id: 1
               }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

          expect(response).to have_http_status(:ok)

          json_response = response.parsed_body
          content = json_response['result']['content'].first
          result = JSON.parse(content['text'])

          expect(result['count']).to eq(1)
        end

        it 'limitが50を超える場合は50に制限される' do
          # 多くのTODOを作成
          55.times do |i|
            create(:todo, user: user, title: "TODO #{i}")
          end

          post '/mcp',
               params: {
                 jsonrpc: '2.0',
                 method: 'tools/call',
                 params: {
                   name: 'search_todos',
                   arguments: { query: 'TODO', limit: 100 }
                 },
                 id: 1
               }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

          expect(response).to have_http_status(:ok)

          json_response = response.parsed_body
          content = json_response['result']['content'].first
          result = JSON.parse(content['text'])

          expect(result['count']).to eq(50)
        end
      end

      context '検索結果なし' do
        it '0件の結果を返す' do
          post '/mcp',
               params: {
                 jsonrpc: '2.0',
                 method: 'tools/call',
                 params: {
                   name: 'search_todos',
                   arguments: { query: '存在しないキーワード' }
                 },
                 id: 1
               }.to_json,
               headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

          expect(response).to have_http_status(:ok)

          json_response = response.parsed_body
          content = json_response['result']['content'].first
          result = JSON.parse(content['text'])

          expect(result['count']).to eq(0)
          expect(result['total_found']).to eq(0)
          expect(result['todos']).to eq([])
        end
      end
    end

    context 'エラーハンドリング' do
      it '無効なJSON-RPCリクエストでエラーを返す' do
        post '/mcp',
             params: 'invalid json',
             headers: headers.merge({ 'Authorization' => "Bearer #{valid_token}" })

        expect(response).to have_http_status(:ok)

        json_response = response.parsed_body
        expect(json_response['jsonrpc']).to eq('2.0')
        expect(json_response['error']['code']).to eq(-32_700)
        expect(json_response['error']['message']).to eq('Parse error')
      end
    end
  end
end
