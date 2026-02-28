## 1. Backend: AppError type and error codes

- [x] 1.1 Create `model/apperror.go` with `AppError` struct (Code string, Message string), `Error()` method, and `Is()` support for `errors.Is()` compatibility
- [x] 1.2 Define UPPER_SNAKE_CASE error code constants for all domain error groups: auth, workspace, transaction, category, funding, invite, rule, bank connection
- [x] 1.3 Add generic handler-level codes: `UNAUTHORIZED`, `INVALID_REQUEST`, `INTERNAL_ERROR`
- [x] 1.4 Migrate all sentinel errors in `model/errors.go` from `errors.New()` to `AppError` values, preserving existing variable names and `.Error()` output
- [x] 1.5 Write table-driven tests for `AppError`: `.Error()` returns message, `errors.Is()` matches sentinel vars, `Code` field is accessible

## 2. Backend: Coded error response helpers

- [x] 2.1 Add `codedErrorEnvelope` struct with `Code` and `Message` JSON fields to `handler/response.go`
- [x] 2.2 Add `respondCodedError(w, status, code, message)` function to `handler/response.go`
- [x] 2.3 Update `handleServiceError` in `handler/handler_workspace.go` to extract `AppError.Code` and call `respondCodedError` instead of `respondError`
- [x] 2.4 Update all inline `respondError` calls across handler files to use `respondCodedError` with appropriate codes (UNAUTHORIZED, INVALID_REQUEST, etc.)
- [x] 2.5 Remove or deprecate the old `respondError` function and `errorEnvelope` struct
- [x] 2.6 Write tests for `respondCodedError` output format and `handleServiceError` code extraction

## 3. Backend: User preferred language

- [x] 3.1 Add `PreferredLanguage string` field to `model.User` struct
- [x] 3.2 Create SQL migration: `ALTER TABLE users ADD COLUMN preferred_language VARCHAR(5) NOT NULL DEFAULT 'en'` with corresponding down migration
- [x] 3.3 Add sqlc query for `UpdatePreferredLanguage(ctx, userID, language)` and regenerate sqlc code
- [x] 3.4 Update existing sqlc queries (GetOrCreateByEmail, GetByID) to include `preferred_language` in SELECT and return mapping
- [x] 3.5 Add `ErrUnsupportedLanguage` AppError to `model/errors.go` with code `UNSUPPORTED_LANGUAGE`
- [x] 3.6 Add service-layer method to validate language against supported set (`en`, `pl`) and call store
- [x] 3.7 Write tests for language validation (valid, invalid, not found user)

## 4. Backend: PATCH /api/v1/auth/me endpoint

- [x] 4.1 Add `UpdateProfile` handler method to `AuthHandler` that decodes `{"preferred_language": "..."}`, validates, calls service, returns updated user
- [x] 4.2 Update `GET /api/v1/auth/me` response to include `preferred_language` field
- [x] 4.3 Register `PATCH /api/v1/auth/me` route in `handler/router.go` within the protected group
- [x] 4.4 Write handler tests: successful update, invalid language, empty body (no-op), unauthenticated

## 5. Backend: Verify and run all tests

- [x] 5.1 Run `go test -race -count=1 ./...` and fix any broken tests due to error envelope format change
- [x] 5.2 Update existing handler test assertions from `{"error":"..."}` to `{"code":"...","message":"..."}` format
- [x] 5.3 Run `go vet ./...` to verify no issues

## 6. Frontend: Install i18next dependencies

- [x] 6.1 Install `i18next`, `react-i18next`, `i18next-browser-languagedetector` via npm
- [x] 6.2 Create `frontend/src/i18n.ts` configuration file: init i18next with `fallbackLng: "en"`, supported languages `["en", "pl"]`, namespace list, detection order (localStorage → navigator → fallback), and resource imports

## 7. Frontend: Translation files

- [x] 7.1 Create `frontend/src/locales/en/common.json` — shared UI strings: buttons (save, cancel, delete, edit, close, confirm, back, loading), nav labels, generic headings
- [x] 7.2 Create `frontend/src/locales/en/errors.json` — one key per backend error code (WORKSPACE_NOT_FOUND, INSUFFICIENT_PERMISSION, etc.) with English display text
- [x] 7.3 Create `frontend/src/locales/en/auth.json` — login page strings (sign in with Google/GitHub, redirecting, logout)
- [x] 7.4 Create `frontend/src/locales/en/workspace.json` — workspace UI strings (create, edit, delete workspace, member management, invite)
- [x] 7.5 Create `frontend/src/locales/en/transaction.json` — transaction UI strings (create, edit, delete, import, categories, amounts, date labels)
- [x] 7.6 Create `frontend/src/locales/en/validation.json` — Zod validation messages (required, too_small, too_big, invalid_type, invalid_string)
- [x] 7.7 Create `frontend/src/locales/pl/common.json` — Polish translations matching all `en/common.json` keys
- [x] 7.8 Create `frontend/src/locales/pl/errors.json` — Polish translations matching all `en/errors.json` keys
- [x] 7.9 Create `frontend/src/locales/pl/auth.json` — Polish translations matching all `en/auth.json` keys
- [x] 7.10 Create `frontend/src/locales/pl/workspace.json` — Polish translations matching all `en/workspace.json` keys
- [x] 7.11 Create `frontend/src/locales/pl/transaction.json` — Polish translations matching all `en/transaction.json` keys
- [x] 7.12 Create `frontend/src/locales/pl/validation.json` — Polish translations matching all `en/validation.json` keys

## 8. Frontend: i18next provider and app integration

- [x] 8.1 Import `frontend/src/i18n.ts` in the app entry point (before React render)
- [x] 8.2 Wrap the root React component with `I18nextProvider` or verify Suspense-based loading works
- [x] 8.3 Configure Zod global error map in app init that reads from `validation` namespace via `i18next.t()`

## 9. Frontend: Replace hardcoded strings with t() calls

- [x] 9.1 Update `LoginPage.tsx` — replace "Sign in with Google", "Sign in with GitHub", "Redirecting..." with `t("auth:...")` calls
- [x] 9.2 Update `ConfirmDialog.tsx` — replace "Delete", "Cancel", "Deleting..." with `t("common:...")` calls
- [x] 9.3 Update `lib/errors.ts` (`getErrorMessage`) — extract `data.code`, return `i18next.t("errors:{code}")`, fall back to `data.message` then fallback string
- [x] 9.4 Update workspace-related components — replace hardcoded workspace labels, form placeholders, and toast messages with `t("workspace:...")` calls
- [x] 9.5 Update transaction-related components — replace hardcoded transaction labels, import messages, category text with `t("transaction:...")` calls
- [x] 9.6 Update remaining shared components — replace button text, nav labels, headings, and other common strings with `t("common:...")` calls
- [x] 9.7 Update Zod schema inline messages (e.g., `.min(1, "Required")`) to remove hardcoded text (the global error map handles it)

## 10. Frontend: Language switcher component

- [x] 10.1 Create `LanguageSwitcher` component that toggles between "EN" and "PL", calls `i18next.changeLanguage()`, and persists to localStorage
- [x] 10.2 Add API call in switcher: if user is authenticated, fire `PATCH /api/v1/auth/me` with `{ preferred_language }` on change
- [x] 10.3 Add `LanguageSwitcher` to the app layout (header/sidebar) so it's accessible on all pages

## 11. Frontend: Locale-aware formatting

- [x] 11.1 Create a `useLocale` hook (or utility) that returns the current `date-fns` locale object (`enUS` or `pl`) based on `i18next.language`
- [x] 11.2 Update all `date-fns` `format()` calls across components to pass the locale from the hook
- [x] 11.3 Create a `formatAmount` utility using `Intl.NumberFormat` with the active i18next locale for monetary display
- [x] 11.4 Replace hardcoded number formatting in transaction/summary components with `formatAmount` calls

## 12. Frontend: Profile language sync on login

- [x] 12.1 Update the auth/login flow: after `GET /api/v1/auth/me` succeeds, read `preferred_language` from the response
- [x] 12.2 Implement sync logic: if profile language differs from localStorage and profile is not default (`en`), switch i18next to profile language; if profile is default and localStorage differs, PATCH profile to match localStorage
- [x] 12.3 Verify language persists across page refresh and new browser sessions
