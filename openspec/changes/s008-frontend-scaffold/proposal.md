## Why

The SpendWise backend API is feature-complete (workspaces, transactions, categories, summaries, fundings, rules, OAuth auth). Users need a web interface to interact with the system. This change scaffolds the complete frontend application with all pages, routing, auth flow, and API integration.

## What Changes

- Create a `frontend/` directory with a React + Vite + TypeScript application
- Implement OAuth login flow (Google, GitHub) using cookie-based auth from the backend
- Build all pages: Dashboard (summary + charts), Transactions (table + import), Categories, Fundings, Rules, Workspace Settings
- Responsive layout with sidebar navigation that collapses on mobile
- API client layer using Axios + TanStack Query for all backend endpoints
- Type-safe TypeScript types mirroring Go models
- Vite dev server proxy to backend for seamless development
- Docker integration for production build (nginx serving static files + API proxy)

## Capabilities

### New Capabilities
- `frontend-project`: Vite + React + TypeScript project setup, Tailwind CSS, shadcn/ui configuration, ESLint, and build tooling
- `frontend-auth`: OAuth login/logout flow, protected routes, auth state management using cookie-based sessions
- `frontend-api-client`: Axios instance with credentials, TanStack Query hooks for all API endpoints, TypeScript request/response types
- `frontend-layout`: App shell with responsive sidebar, header with workspace selector and user avatar, mobile hamburger menu
- `frontend-dashboard`: Dashboard page with spending summary cards, category breakdown chart (Recharts), recent transactions list
- `frontend-transactions`: Transactions page with data table, date range filtering, bank import functionality
- `frontend-categories`: Categories page with CRUD operations, icon display
- `frontend-fundings`: Fundings page for recording and listing workspace fundings
- `frontend-rules`: Rules page for creating, editing, toggling, and deleting automation rules
- `frontend-workspace-settings`: Workspace settings page with member management (add/remove), workspace name/description editing

### Modified Capabilities
- (none — this is a new frontend, no existing frontend specs to modify)

## Impact

- **Code**: New `frontend/` directory with ~40-50 source files
- **Dependencies**: Node.js ecosystem (React, Vite, Tailwind, shadcn/ui, TanStack Query, Axios, Recharts, React Hook Form, Zod, date-fns, Lucide icons)
- **Infrastructure**: Docker Compose updated to build and serve frontend via nginx
- **Backend dependency**: Requires `s007-cookie-auth-cors` to be implemented first (cookie-based auth + CORS)
