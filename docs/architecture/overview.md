# System Architecture Overview

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Docker Compose                          │
├─────────────────────┬─────────────────────┬───────────────────┤
│                     │                     │                   │
│   Frontend          │   Backend           │   Database        │
│   Next.js 15.4.1    │   Rails 7.1.3+      │   PostgreSQL 15   │
│   Port: 3000        │   Port: 3001        │   Port: 5432      │
│                     │                     │                   │
│   - React 19        │   - Ruby 3.2.5      │                   │
│   - TypeScript 5    │   - Devise + JWT    │                   │
│   - Tailwind CSS 4  │   - API-only mode   │                   │
│   - pnpm            │   - RSpec tests     │                   │
└─────────────────────┴─────────────────────┴───────────────────┘
```

## Key Design Decisions

### 1. Frontend: Next.js with App Router
- **Why Next.js**: Server-side rendering capabilities, excellent developer experience, and strong TypeScript support
- **App Router**: Latest Next.js pattern for better performance and simplified data fetching
- **React 19**: Latest stable version with improved performance and concurrent features

### 2. Backend: Rails API-only
- **Why Rails**: Mature framework with excellent conventions and ActiveRecord ORM
- **API-only mode**: Lightweight, focused on serving JSON APIs
- **JWT Authentication**: Stateless authentication suitable for SPA architecture

### 3. Database: PostgreSQL
- **Why PostgreSQL**: Robust, feature-rich RDBMS with excellent Rails support
- **Version 15**: Latest stable version with improved performance

### 4. Infrastructure: Docker Compose
- **Why Docker**: Consistent development environment across team members
- **Compose**: Simple orchestration for local development
- **Hot reloading**: Both frontend and backend support live code updates

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
                    Rails Backend (:3001)
                            │
                    ┌───────┴────────┐
                    │                │
                JWT Auth        Todo CRUD
                    │                │
                    └───────┬────────┘
                            │
                            ▼
                    PostgreSQL (:5432)
```

## Security Considerations

1. **Authentication**: JWT tokens with secure storage in localStorage
2. **CORS**: Configured to allow only frontend origin
3. **Data Validation**: Both client-side and server-side validation
4. **User Isolation**: Each user can only access their own todos

## Scalability Considerations

1. **Stateless Architecture**: JWT allows horizontal scaling of backend
2. **Database Indexing**: Proper indexes on foreign keys and frequently queried fields
3. **API Design**: RESTful design allows for easy caching and CDN integration
4. **Frontend Optimization**: Next.js provides automatic code splitting and optimization