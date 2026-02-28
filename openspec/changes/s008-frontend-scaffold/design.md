## Context

SpendWise has a complete Go backend API with OAuth authentication, workspace management, transaction CRUD, categories, fundings, rules, and spending summaries. The backend will use httpOnly cookie-based auth (from s007-cookie-auth-cors). We need a web frontend that provides a polished, responsive UI for all these features.

The target users are household members tracking shared expenses — not developers. The UI must be intuitive, responsive, and fast.

## Goals / Non-Goals

**Goals:**
- Complete, functional web UI covering all backend API features
- Commercial-quality look and feel using shadcn/ui + Tailwind
- Responsive design (desktop sidebar, mobile hamburger)
- Type-safe API integration with proper loading/error states
- Fast development feedback loop (Vite HMR, API proxy)
- Production-ready Docker build

**Non-Goals:**
- PWA / offline support (future)
- Real-time updates / WebSockets (future)
- i18n / localization (future)
- Dark mode (future — shadcn supports it but we won't wire the toggle yet)
- E2E tests (future — unit tests for hooks/utils only in this change)
- Native mobile app

## Decisions

### Decision 1: React 19 + Vite + TypeScript

Standard SPA stack. Vite provides fast HMR and optimized builds. No SSR needed — this is a private app behind auth, not a content site.

**Alternatives considered:**
- Next.js: Adds SSR/SSG complexity we don't need. We already have a Go backend serving the API.
- SvelteKit: Smaller ecosystem, fewer UI component libraries for dashboards.

### Decision 2: shadcn/ui + Tailwind CSS v4

shadcn/ui copies accessible, well-designed components directly into the project. We own the code and can customize freely. Built on Radix UI primitives (accessibility-first).

Tailwind CSS v4 for utility-first styling — pairs natively with shadcn/ui.

**Alternatives considered:**
- MUI: Heavy, opinionated, hard to customize away from Material Design.
- Mantine: Good but smaller community, less content/examples available.
- Ant Design: Enterprise-focused, heavy bundle, harder to customize.

### Decision 3: TanStack Query v5 for server state

Every API call goes through TanStack Query. It handles:
- Caching (avoid redundant fetches)
- Background refetching (data stays fresh)
- Loading/error states (consistent UX)
- Optimistic updates (for mutations)
- Pagination (transactions list)

**Alternatives considered:**
- SWR: Similar but less feature-rich (no mutation helpers, less devtools).
- Redux Toolkit Query: Heavier, pulls in Redux. Overkill for this app.
- Raw fetch + useState: Manual cache management, error handling, loading states — lots of boilerplate.

### Decision 4: Axios with credentials

Axios instance configured with:
- `baseURL: "/api/v1"` (proxied in dev, served by nginx in prod)
- `withCredentials: true` (sends httpOnly cookies)
- Response interceptor: on 401, redirect to login page

No token management code needed — browser handles cookies automatically.

### Decision 5: React Router v7 for routing

File-convention-like routing using React Router v7 with a simple route config in `App.tsx`. Protected routes wrapped in an `<AuthGuard>` component that checks `/auth/me` on mount.

Route structure:
```
/login                    → LoginPage (public)
/auth/:provider/callback  → AuthCallback (public, handles OAuth redirect)
/                         → redirect to /dashboard
/dashboard                → DashboardPage (protected)
/transactions             → TransactionsPage (protected)
/categories               → CategoriesPage (protected)
/fundings                 → FundingsPage (protected)
/rules                    → RulesPage (protected)
/settings                 → WorkspaceSettingsPage (protected)
```

### Decision 6: React Hook Form + Zod for forms

All forms (create transaction, create category, create rule, import, workspace settings) use React Hook Form for performance (minimal re-renders) and Zod schemas for validation that mirrors backend constraints.

### Decision 7: Recharts for data visualization

Dashboard uses Recharts for:
- Pie/donut chart: spending by category
- Bar chart: spending over time (optional, based on summary data)

Recharts is the most popular React charting library, composable, and responsive.

### Decision 8: Project structure — flat pages + feature-grouped API

```
src/
├── api/          # Axios instance + per-resource API functions + TanStack Query hooks
├── components/   # ui/ (shadcn), layout/ (shell, sidebar), shared/ (reusable)
├── pages/        # One component per route
├── hooks/        # Custom hooks (useAuth, useWorkspace)
├── lib/          # Utilities (money formatting, date helpers)
├── types/        # TypeScript types mirroring Go models
├── App.tsx       # Router config
└── main.tsx      # Entry point
```

### Decision 9: Dev proxy + production nginx

**Development:** Vite's `server.proxy` sends `/api` requests to `localhost:8080`.

**Production:** Multi-stage Docker build:
1. Node stage: `npm run build` → static files
2. Nginx stage: serves static files + proxies `/api` to backend container

### Decision 10: Workspace context

The app operates within a single workspace at a time. A `WorkspaceContext` provider stores the selected workspace ID. The workspace selector in the header switches between workspaces. All API calls for resources (transactions, categories, etc.) use the current workspace ID.

On first login, if the user has no workspaces, prompt them to create one. If they have one, auto-select it. If multiple, show the selector.

## Risks / Trade-offs

- **[Risk] Large initial scope** — 10 capabilities is a big first change. → Mitigated by clean separation: each page is independent. Tasks are ordered so the app is functional early (auth + dashboard first, then other pages).

- **[Risk] Type drift between Go and TypeScript** — Types may diverge as backend evolves. → For now, manually maintained TypeScript types. Future: OpenAPI spec generation from Go handlers.

- **[Trade-off] No E2E tests in this change** — Would slow down initial delivery significantly. → Unit tests for critical hooks/utils. E2E tests as a follow-up change.

- **[Trade-off] No dark mode toggle** — shadcn supports it but wiring the toggle + persistence adds scope. → CSS variables are set up for it; toggle can be added later trivially.
