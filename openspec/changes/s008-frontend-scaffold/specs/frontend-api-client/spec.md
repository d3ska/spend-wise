## ADDED Requirements

### Requirement: Axios instance with credentials
The API client SHALL be an Axios instance configured with `baseURL: "/api/v1"` and `withCredentials: true` so httpOnly cookies are sent automatically.

#### Scenario: Cookies included in requests
- **WHEN** the API client makes any request
- **THEN** the browser SHALL include the httpOnly `sw_token` cookie

#### Scenario: 401 response triggers redirect
- **WHEN** the API client receives a 401 response
- **THEN** the response interceptor SHALL redirect the user to `/login`

### Requirement: TypeScript types for API models
The `types/` directory SHALL contain TypeScript interfaces mirroring all Go models: `User`, `Workspace`, `Category`, `Transaction`, `TransactionEntry`, `Funding`, `Rule`, `Summary`, `Money`, `DateRange`.

#### Scenario: Type safety on API responses
- **WHEN** an API function returns a workspace
- **THEN** the return type SHALL be `Workspace` with fields matching the backend JSON shape

### Requirement: TanStack Query hooks for all endpoints
Each API resource SHALL have a dedicated file with query/mutation hooks using TanStack Query.

#### Scenario: Workspaces query hook
- **WHEN** `useWorkspaces()` is called in a component
- **THEN** it SHALL return `{ data, isLoading, error }` with data typed as `Workspace[]`

#### Scenario: Create transaction mutation
- **WHEN** `useCreateTransaction()` is used and the mutation succeeds
- **THEN** it SHALL invalidate the transactions query cache so the list refetches

#### Scenario: Loading state displayed
- **WHEN** a query is in-flight
- **THEN** `isLoading` SHALL be `true` so components can show loading indicators

#### Scenario: Error state displayed
- **WHEN** a query fails
- **THEN** `error` SHALL contain the error details so components can show error messages
