# Frontend

Next.js 16 + React 19.2 + TypeScript + Tailwind CSS v4

## 技術スタック

- **Framework**: Next.js 16 (App Router)
- **UI Library**: React 19.2
- **Language**: TypeScript 5
- **Styling**: Tailwind CSS v4
- **Components**: shadcn/ui
- **Package Manager**: pnpm

## セットアップ

```bash
# Docker環境（推奨）
docker compose up -d

# ローカル開発
cd frontend
pnpm install
pnpm run dev
```

## 開発コマンド

```bash
# Docker環境
docker compose exec frontend pnpm run dev        # 開発サーバー
docker compose exec frontend pnpm run build      # 本番ビルド
docker compose exec frontend pnpm run lint       # ESLint
docker compose exec frontend pnpm run lint:fix   # ESLint (自動修正)
docker compose exec frontend pnpm run typecheck  # TypeScript チェック

# ローカル環境
pnpm run dev
pnpm run build
pnpm run lint
pnpm run typecheck
```

## ディレクトリ構造

```
src/
├── app/                  # Next.js App Router
│   ├── (auth)/          # 認証関連ページ
│   ├── (dashboard)/     # ダッシュボード
│   └── layout.tsx       # Root layout
├── components/           # 共通UIコンポーネント (shadcn/ui)
├── contexts/             # React Context
│   └── auth-context.tsx # 認証状態管理
├── features/             # 機能別モジュール
│   ├── todo/            # Todo機能
│   ├── category/        # カテゴリ管理
│   ├── tag/             # タグ管理
│   ├── comment/         # コメント
│   ├── history/         # 変更履歴
│   ├── file/            # ファイル添付
│   └── notes/           # ノート機能
├── hooks/                # 共通フック
├── lib/                  # ユーティリティ
│   ├── api-client.ts    # API クライアント
│   ├── auth-client.ts   # 認証 API
│   └── constants.ts     # 定数
└── types/                # 型定義
```

## Feature モジュール構成

各 feature は以下の構成:

```
features/todo/
├── components/          # UI コンポーネント
├── hooks/               # カスタムフック
├── lib/
│   └── api.ts          # API 関数
└── types.ts            # 型定義
```

## API クライアント

```typescript
// ApiClient を使用（/api/v1 プレフィックス付き）
import { apiClient } from '@/lib/api-client';

// GET
const todos = await apiClient.get<TodosResponse>('/todos');

// POST
const todo = await apiClient.post<Todo>('/todos', { todo: { title: 'New' } });

// PATCH
const updated = await apiClient.patch<Todo>(`/todos/${id}`, { todo: data });

// DELETE
await apiClient.delete<void>(`/todos/${id}`);
```

## 環境変数

```env
NEXT_PUBLIC_API_URL=http://localhost:3001
```

## Path Aliases

```typescript
// tsconfig.json で設定済み
import { Button } from '@/components/ui/button';
import { useTodos } from '@/features/todo/hooks/useTodos';
```

## 関連ドキュメント

- [Architecture Overview](../docs/architecture/overview.md)
- [API Documentation](../docs/api/README.md)
- [CLAUDE.md](../CLAUDE.md) - 開発ガイドライン
