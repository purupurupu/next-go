# Next.js + Go Todo Application

Full-stack Todoアプリケーション with JWT認証

## Tech Stack

| Layer | Technology |
|-------|------------|
| Frontend | Next.js 16, React 19, TypeScript, Tailwind CSS v4 |
| Backend | Go 1.25, Echo v4, GORM |
| Database | PostgreSQL 15 |
| Cache | Redis 7 |
| Infrastructure | Docker Compose |

## Prerequisites

- [Docker](https://www.docker.com/) & Docker Compose
- [mise](https://mise.jdx.dev/)（ローカル開発用、任意）

## Quick Start

### 1. リポジトリをクローン

```bash
git clone <repository-url>
cd next-go
```

### 2. 環境変数を設定

```bash
cp .env.example .env
```

`.env` を編集（本番環境では必ず `JWT_SECRET` を変更）:

```env
POSTGRES_DB=todo_next
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
JWT_SECRET=your-super-secret-key-change-in-production
```

### 3. Dockerで起動

```bash
docker compose up -d
```

### 4. 動作確認

- Frontend: http://localhost:3000
- Backend API: http://localhost:3001
- Health Check: http://localhost:3001/health

## Local Development（mise）

Dockerを使わずローカルで開発する場合：

```bash
# miseのインストール（未インストールの場合）
curl https://mise.run | sh

# Node.jsのGPG公開鍵をインポート（初回のみ）
gpg --keyserver hkps://keys.openpgp.org --recv-keys 86C8D74642E67846F8E120284DAA80D1E737BC9F

# ツールのインストール
mise trust
mise install

# バージョン確認
go version   # go1.25.5
node -v      # v24.12.0
```

### Cursor / VS Code で mise を認識させる

Cursor や VS Code が mise で管理している Go を認識しない場合、`.vscode/settings.json` に以下を追加：

```json
{
    "go.gopath": "/Users/<your-username>/go",
    "go.goroot": "/Users/<your-username>/.local/share/mise/installs/go/1.25.5",
    "go.alternateTools": {
        "go": "/Users/<your-username>/.local/share/mise/installs/go/1.25.5/bin/go"
    }
}
```

パスは `mise which go` で確認できます。

## Development Commands

### Docker操作

```bash
# 起動
docker compose up -d

# 停止
docker compose down

# ログ確認
docker compose logs -f backend
docker compose logs -f frontend

# 再ビルド（依存関係更新後）
docker compose build backend
docker compose build frontend
```

### Frontend

```bash
docker compose exec frontend pnpm run dev        # 開発サーバー
docker compose exec frontend pnpm run build      # ビルド
docker compose exec frontend pnpm run lint       # Lint
docker compose exec frontend pnpm run typecheck  # 型チェック
```

### Backend

```bash
docker compose exec backend go build -o bin/api cmd/api/main.go  # ビルド
docker compose exec backend go fmt ./...         # フォーマット
```

### テスト

```bash
# テスト用DBを起動
docker compose up -d db_test

# テスト実行（コンテナ内）
docker compose exec backend go test -v ./...

# 特定のパッケージのみ
docker compose exec backend go test -v ./internal/handler/...
```

#### `./...` とは？

Goのパッケージパス指定の構文です。

| パス | 意味 |
|-----|------|
| `./` | 現在のディレクトリ |
| `...` | このディレクトリ以下すべて（ワイルドカード） |

```bash
go test ./...                    # すべてのパッケージ
go test ./internal/handler/...   # handler以下すべて
go test ./internal/handler       # handlerのみ（サブディレクトリ除く）
```

#### ローカルでテスト実行

Dockerコンテナを経由せずローカルで実行する場合は、環境変数でDB接続先を指定：

```bash
cd backend
TEST_DATABASE_URL="host=localhost user=postgres password=password dbname=todo_next_test port=5433 sslmode=disable" \
  go test -v ./...
```

## Project Structure

```
.
├── frontend/          # Next.js アプリケーション
├── backend/           # Go API サーバー
├── docs/              # 詳細ドキュメント
├── compose.yml        # Docker Compose 設定
├── .mise.toml         # mise バージョン管理
└── CLAUDE.md          # Claude Code 用ガイド
```

## Documentation

詳細なドキュメントは [docs/](./docs/) を参照してください：

- [Architecture](./docs/architecture/) - システム設計
- [API Documentation](./docs/api/) - API仕様
- [Development Guide](./docs/guides/) - 開発ガイドライン

## License

MIT
