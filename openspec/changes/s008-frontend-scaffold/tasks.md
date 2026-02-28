## 1. Project Setup

- [x] 1.1 Initialize Vite + React + TypeScript project in `frontend/` with `npm create vite@latest`
- [x] 1.2 Configure TypeScript strict mode and `@/` path alias in `tsconfig.json` and `vite.config.ts`
- [x] 1.3 Install and configure Tailwind CSS v4
- [x] 1.4 Initialize shadcn/ui with `components.json`, install base components (Button, Card, Dialog, DropdownMenu, Input, Label, Select, Sheet, Separator, Skeleton, Table, Tabs, Toast/Sonner, Avatar, Badge)
- [x] 1.5 Install project dependencies: axios, @tanstack/react-query, react-router-dom, react-hook-form, @hookform/resolvers, zod, recharts, lucide-react, date-fns, sonner
- [x] 1.6 Configure Vite dev proxy (`/api` → `localhost:8080`) in `vite.config.ts`
- [x] 1.7 Configure ESLint for TypeScript + React

## 2. Types and API Client

- [x] 2.1 Create `src/types/index.ts` with TypeScript interfaces: User, Workspace, WorkspaceMember, Category, Transaction, TransactionEntry, Funding, Rule, Summary, Money
- [x] 2.2 Create `src/api/client.ts` — Axios instance with `baseURL: "/api/v1"`, `withCredentials: true`, and 401 redirect interceptor
- [x] 2.3 Create `src/api/auth.ts` — API functions: `getAuthURL(provider)`, `handleCallback(provider, code, state)`, `getMe()`, `logout()`
- [x] 2.4 Create `src/api/workspaces.ts` — API functions + TanStack Query hooks: `useWorkspaces()`, `useWorkspace(id)`, `useCreateWorkspace()`, `useUpdateWorkspace()`, `useDeleteWorkspace()`, `useAddMember()`, `useRemoveMember()`
- [x] 2.5 Create `src/api/categories.ts` — API functions + hooks: `useCategories(workspaceId)`, `useCreateCategory()`, `useUpdateCategory()`, `useDeleteCategory()`
- [x] 2.6 Create `src/api/transactions.ts` — API functions + hooks: `useTransactions(workspaceId, params)`, `useTransaction(workspaceId, txId)`, `useCreateTransaction()`, `useDeleteTransaction()`, `useImportTransactions()`
- [x] 2.7 Create `src/api/summary.ts` — API functions + hooks: `useSummary(workspaceId, from, to)`
- [x] 2.8 Create `src/api/fundings.ts` — API functions + hooks: `useFundings(workspaceId)`, `useRecordFunding()`
- [x] 2.9 Create `src/api/rules.ts` — API functions + hooks: `useRules(workspaceId)`, `useCreateRule()`, `useUpdateRule()`, `useDeleteRule()`, `useToggleRule()`

## 3. Auth and Routing

- [x] 3.1 Create `src/hooks/useAuth.ts` — auth context provider using `getMe()` query, exposes `user`, `isLoading`, `isAuthenticated`, `logout()`
- [x] 3.2 Create `src/components/auth/AuthGuard.tsx` — redirects to `/login` if unauthenticated, shows loading skeleton while checking
- [x] 3.3 Create `src/pages/LoginPage.tsx` — centered card with Google and GitHub OAuth buttons, app logo
- [x] 3.4 Create `src/pages/AuthCallbackPage.tsx` — extracts `code`/`state` from URL, calls callback API, redirects to dashboard or shows error
- [x] 3.5 Configure React Router in `src/App.tsx` — public routes (login, callback), protected routes wrapped in AuthGuard

## 4. Layout

- [x] 4.1 Create `src/components/layout/Sidebar.tsx` — navigation links with Lucide icons, active route highlighting, responsive (hidden on mobile)
- [x] 4.2 Create `src/components/layout/Header.tsx` — app name, workspace selector dropdown, user avatar dropdown with logout
- [x] 4.3 Create `src/components/layout/AppShell.tsx` — combines Sidebar + Header + main content area, mobile Sheet for sidebar
- [x] 4.4 Create `src/hooks/useWorkspace.ts` — workspace context provider storing selected workspace ID, auto-select logic (single workspace → auto, none → onboarding prompt)

## 5. Shared Components

- [x] 5.1 Create `src/components/shared/MoneyDisplay.tsx` — formats Money value object with currency symbol and thousands separator
- [x] 5.2 Create `src/components/shared/DateRangePicker.tsx` — date range selection using shadcn Calendar/Popover, with presets (This Month, Last Month, This Year)
- [x] 5.3 Create `src/components/shared/ConfirmDialog.tsx` — reusable confirmation dialog for delete actions
- [x] 5.4 Create `src/lib/money.ts` — formatMoney utility function
- [x] 5.5 Create `src/lib/dates.ts` — date range helpers (thisMonth, lastMonth, thisYear), date formatting with date-fns

## 6. Dashboard Page

- [x] 6.1 Create `src/pages/DashboardPage.tsx` — layout with summary cards, chart, and recent transactions
- [x] 6.2 Implement summary stat cards (Total Spent, date range display) using shadcn Card
- [x] 6.3 Implement category breakdown donut chart using Recharts PieChart
- [x] 6.4 Implement recent transactions list (last 5) with link to full transactions page
- [x] 6.5 Wire date range picker to refetch summary data

## 7. Transactions Page

- [x] 7.1 Create `src/pages/TransactionsPage.tsx` — data table with date range filter and action buttons
- [x] 7.2 Implement transactions data table using shadcn Table with columns: Date, Description, Category, Amount
- [x] 7.3 Implement pagination controls for the transactions list
- [x] 7.4 Implement create transaction dialog with React Hook Form + Zod validation (description, date, category select, entries with amounts)
- [x] 7.5 Implement delete transaction with ConfirmDialog
- [x] 7.6 Implement bank import dialog — file upload, POST to import endpoint, show imported/skipped counts

## 8. Categories Page

- [x] 8.1 Create `src/pages/CategoriesPage.tsx` — grid or list of categories with CRUD actions
- [x] 8.2 Implement create/edit category dialog with name and icon fields
- [x] 8.3 Implement delete category with ConfirmDialog

## 9. Fundings Page

- [x] 9.1 Create `src/pages/FundingsPage.tsx` — list of fundings with record form
- [x] 9.2 Implement record funding dialog with amount, member, description fields
- [x] 9.3 Implement fundings list display

## 10. Rules Page

- [x] 10.1 Create `src/pages/RulesPage.tsx` — list of rules with CRUD and toggle
- [x] 10.2 Implement create/edit rule dialog with form fields
- [x] 10.3 Implement rule toggle switch using shadcn Switch
- [x] 10.4 Implement delete rule with ConfirmDialog

## 11. Workspace Settings Page

- [x] 11.1 Create `src/pages/WorkspaceSettingsPage.tsx` — workspace details form + member list
- [x] 11.2 Implement workspace name/description edit form (owner-only)
- [x] 11.3 Implement member list display with roles
- [x] 11.4 Implement add member dialog (owner-only)
- [x] 11.5 Implement remove member with ConfirmDialog (owner-only, cannot remove self)
- [x] 11.6 Implement delete workspace with ConfirmDialog and warning (owner-only)

## 12. Docker Integration

- [x] 12.1 Create `frontend/Dockerfile` — multi-stage build: Node for `npm run build`, nginx for serving static files + API proxy
- [x] 12.2 Create `frontend/nginx.conf` — serve static files at `/`, proxy `/api` to backend container
- [x] 12.3 Update `docker-compose.yml` to add frontend service, expose port 3000

## 13. Polish and Verification

- [x] 13.1 Add `NotFoundPage.tsx` for 404 routes
- [x] 13.2 Add loading skeletons for all pages using shadcn Skeleton
- [x] 13.3 Add toast notifications (Sonner) for success/error on mutations (create, delete, update)
- [x] 13.4 Verify responsive layout on mobile viewport sizes
- [x] 13.5 Run `npm run build` and `npm run lint` to verify clean build
