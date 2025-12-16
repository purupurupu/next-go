# System Architecture Overview

## Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                              Docker Compose                                   │
├──────────────┬──────────────┬────────────┬──────────┬──────────────────────┤
│              │              │            │          │                      │
│  Frontend    │  Backend     │  Database  │  Cache   │  Storage             │
│  Next.js 16  │  Go 1.25     │  Postgres  │  Redis 7 │  RustFS              │
│  Port: 3000  │  Port: 3001  │  Port:5432 │  :6379   │  Port: 9000          │
│              │              │            │          │                      │
│  - React 19  │  - Echo v4   │            │          │  - S3互換            │
│  - TypeScript│  - GORM      │            │          │  - ファイル保存      │
│  - Tailwind 4│  - JWT       │            │          │  - サムネイル        │
│  - pnpm      │  - zerolog   │            │          │                      │
└──────────────┴──────────────┴────────────┴──────────┴──────────────────────┘
```

## Key Design Decisions

### 1. Frontend: Next.js with App Router
- **Why Next.js**: Server-side rendering capabilities, excellent developer experience, and strong TypeScript support
- **App Router**: Latest Next.js pattern for better performance and simplified data fetching
- **React 19**: Latest stable version with improved performance and concurrent features
- **pnpm**: Fast, disk-efficient package manager

### 2. Backend: Go with Echo Framework
- **Why Go**: High performance, strong concurrency support, simple deployment (single binary)
- **Echo v4**: Lightweight, high-performance web framework with excellent middleware support
- **GORM**: Feature-rich ORM with migration support
- **JWT Authentication**: Stateless authentication suitable for SPA architecture

### 3. Database: PostgreSQL
- **Why PostgreSQL**: Robust, feature-rich RDBMS
- **Version 15**: Latest stable version with improved performance

### 4. Storage: RustFS (S3 Compatible)
- **Why RustFS**: S3互換のローカル開発用ストレージ
- **File handling**: Todo添付ファイル、サムネイル自動生成

### 5. Infrastructure: Docker Compose
- **Why Docker**: Consistent development environment across team members
- **Compose**: Simple orchestration for local development
- **Hot reloading**: Air (Go) + Next.js dev server

## Communication Flow

```
User Browser
    │
    ▼
Next.js Frontend (:3000)
    │
    ├─── Static Assets (JS, CSS)
    │
    └─── API Requests ──────┐
                            │
                            ▼
                    Go Backend (:3001)
                            │
                    ┌───────┴────────┐
                    │                │
                JWT Auth        Business Logic
                    │                │
                    │         ┌──────┴──────┐
                    │         │             │
                    │      Todo CRUD    Notes API
                    │         │             │
                    └────┬────┴─────────────┘
                         │
              ┌──────────┼──────────┐
              │          │          │
              ▼          ▼          ▼
         PostgreSQL   Redis     RustFS
           (:5432)   (:6379)   (:9000)
```

## Security Considerations

1. **Authentication**: JWT tokens with secure storage
2. **CORS**: Configured to allow only frontend origin
3. **Data Validation**: Both client-side (Zod) and server-side (go-playground/validator)
4. **User Isolation**: Each user can only access their own data
5. **Soft Delete**: コメント削除時はソフトデリート

## Scalability Considerations

1. **Stateless Architecture**: JWT allows horizontal scaling of backend
2. **Database Indexing**: Proper indexes on foreign keys and frequently queried fields
3. **API Design**: RESTful design allows for easy caching and CDN integration
4. **Frontend Optimization**: Next.js provides automatic code splitting and optimization
5. **S3 Compatible Storage**: プロダクション環境では AWS S3 等に移行可能
