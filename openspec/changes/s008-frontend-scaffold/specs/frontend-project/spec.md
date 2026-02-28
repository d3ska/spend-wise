## ADDED Requirements

### Requirement: Vite + React + TypeScript project initialization
The `frontend/` directory SHALL contain a Vite project with React 19, TypeScript in strict mode, and proper `tsconfig.json` configuration.

#### Scenario: Project builds successfully
- **WHEN** `npm run build` is executed in the `frontend/` directory
- **THEN** it SHALL produce optimized static files in `frontend/dist/`

#### Scenario: Dev server starts with HMR
- **WHEN** `npm run dev` is executed in the `frontend/` directory
- **THEN** a development server SHALL start on port 3000 with hot module replacement

### Requirement: Tailwind CSS v4 configuration
The project SHALL use Tailwind CSS v4 for utility-first styling, configured to scan all `.tsx` files in `src/`.

#### Scenario: Tailwind classes applied
- **WHEN** a component uses Tailwind utility classes (e.g., `className="flex items-center gap-2"`)
- **THEN** the corresponding CSS SHALL be generated in the build output

### Requirement: shadcn/ui component library
The project SHALL include shadcn/ui configured with the `components.json` file, using the "default" style and Tailwind CSS. Components SHALL be installed into `src/components/ui/`.

#### Scenario: shadcn components available
- **WHEN** a shadcn component (e.g., Button, Card, Dialog) is imported from `@/components/ui/button`
- **THEN** it SHALL render correctly with default styling

### Requirement: Path aliases
The TypeScript and Vite configurations SHALL support the `@/` path alias resolving to `src/`.

#### Scenario: Import with alias
- **WHEN** a file imports `from "@/components/ui/button"`
- **THEN** it SHALL resolve to `src/components/ui/button`

### Requirement: Vite dev proxy to backend
The Vite dev server SHALL proxy requests matching `/api/**` to `http://localhost:8080`.

#### Scenario: API request proxied in development
- **WHEN** the frontend makes a request to `/api/v1/workspaces` during development
- **THEN** Vite SHALL proxy the request to `http://localhost:8080/api/v1/workspaces`

### Requirement: ESLint configuration
The project SHALL include ESLint configured for TypeScript and React with recommended rules.

#### Scenario: Lint check passes
- **WHEN** `npm run lint` is executed
- **THEN** the codebase SHALL pass linting without errors
