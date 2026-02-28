## Context

SpendWise has ~35 sentinel errors in `model/errors.go`, ~20 inline error strings in handlers, and ~100+ hardcoded strings in the React frontend. All are English-only. The API returns `{"error": "human text"}` with no machine-readable codes — the frontend displays these strings directly via `getErrorMessage()` in `lib/errors.ts`. There is no `Accept-Language` handling, no user locale preference, and no translation infrastructure on either side.

The codebase is small enough that a single coordinated pass can convert everything without incremental migration tooling.

## Goals / Non-Goals

**Goals:**

- Support two locales: English (`en`) and Polish (`pl`), with English as the default.
- Frontend-driven translation: all user-visible text resolved client-side via `react-i18next`.
- Backend returns machine-readable error codes so the frontend maps them to localized strings.
- User can switch language in the UI; preference persists in `localStorage` and optionally in their user profile.
- Locale-aware date and number formatting (Polish uses `dd.MM.yyyy`, comma decimal separator).

**Non-Goals:**

- Server-side rendering of localized text (no SSR, email templates, or PDF generation in scope).
- Right-to-left (RTL) layout support.
- More than two locales in this change (the pattern supports it, but only `en`/`pl` JSON files are created).
- Translating user-generated content (workspace names, transaction descriptions).
- Locale-based currency formatting — SpendWise already uses `shopspring/decimal` with explicit currency.

## Decisions

### 1. Frontend-first translation strategy

**Decision:** All user-facing text is translated on the frontend. The backend provides error codes, not translated messages.

**Rationale:** The React SPA already owns all rendering. Pushing translations to the backend would require locale propagation through the context on every request, a server-side message catalog, and would couple the API to presentation concerns. Frontend-only translation is simpler, cacheable, and standard for SPAs.

**Alternatives considered:**
- *Backend returns localized messages:* Requires Accept-Language negotiation, server-side catalog, and complicates API caching. Adds complexity for no benefit since the SPA renders everything.
- *Hybrid (backend localizes some, frontend others):* Inconsistent, harder to maintain two catalogs.

### 2. Structured error codes in API responses

**Decision:** Change the error envelope from `{"error": "text"}` to `{"code": "SNAKE_UPPER_CODE", "message": "english fallback"}`. The `code` field is a stable machine-readable identifier. The `message` field is kept as a human-readable English fallback for debugging and non-i18n clients.

**Format:**
```json
{
  "code": "WORKSPACE_NOT_FOUND",
  "message": "workspace not found"
}
```

**Rationale:** The frontend needs a stable key to look up translations. Using the sentinel error's `.Error()` text is fragile — any wording change breaks the frontend. A code enum gives a stable contract.

**Alternatives considered:**
- *Use HTTP status codes only:* Not granular enough (multiple 400 errors with different meanings).
- *Numeric error codes:* Less readable, harder to grep, no advantage over string codes.
- *Keep `{"error": "text"}` and match on text client-side:* Fragile, breaks on any wording change, untranslatable.

### 3. Error code implementation approach

**Decision:** Add an `ErrorCode` string type and a `Code` field to each sentinel error using a custom error type. Map existing sentinel errors to codes in a single `errorcode.go` file within the `model` package.

```go
type AppError struct {
    Code    string
    Message string  // English default
}
```

Update `handleServiceError` to extract the code and pass it to a new `respondCodedError(w, status, code, message)` helper.

**Rationale:** Keeps error definitions co-located in `model/`. The handler layer extracts the code without needing a separate mapping table. Existing `errors.Is()` checks continue to work if `AppError` implements `Is()` or if we keep sentinel vars.

**Alternatives considered:**
- *Separate map[error]string in handler package:* Splits error knowledge across packages, easy to forget a mapping.
- *Error codes as constants + separate error text:* More indirection for no gain at this scale.

### 4. react-i18next for frontend translations

**Decision:** Use `i18next` + `react-i18next` + `i18next-browser-languagedetector`.

**Rationale:** `react-i18next` is the de facto standard for React i18n. It supports namespace-based splitting, interpolation, pluralization, and has excellent TypeScript support. `i18next-browser-languagedetector` handles `localStorage` → browser language → fallback chain automatically.

**Alternatives considered:**
- *react-intl (FormatJS):* More opinionated, ICU message syntax has steeper learning curve, smaller community.
- *Custom solution with React Context:* Under-featured, would need to build interpolation, plurals, detection from scratch.

### 5. Translation file structure

**Decision:** Flat JSON files organized by namespace under `frontend/src/locales/`:

```
frontend/src/locales/
├── en/
│   ├── common.json      # Shared: buttons, labels, nav
│   ├── errors.json      # Maps backend error codes → display text
│   ├── auth.json        # Login/logout strings
│   ├── workspace.json   # Workspace-specific UI
│   ├── transaction.json # Transaction-specific UI
│   └── validation.json  # Form validation messages
└── pl/
    ├── common.json
    ├── errors.json
    ├── auth.json
    ├── workspace.json
    ├── transaction.json
    └── validation.json
```

**Rationale:** Namespace-per-feature keeps files small and reviewable. Flat keys (no deep nesting) are easier to search and less error-prone. The `errors.json` namespace maps directly to backend error codes.

**Alternatives considered:**
- *Single file per locale:* Gets unwieldy quickly (200+ keys in one file).
- *Co-located with components:* Harder to review completeness across locales, translation vendors need all strings in one place.

### 6. User locale preference storage

**Decision:** Add `preferred_language VARCHAR(5) NOT NULL DEFAULT 'en'` to the `users` table. Expose via `GET /api/v1/auth/me` response and `PATCH /api/v1/auth/me` for updates. Frontend persists to `localStorage` for immediate use and syncs to the profile on change.

**Language resolution order:**
1. `localStorage` value (fastest, no network)
2. User profile `preferred_language` (loaded on auth, written to `localStorage`)
3. `i18next-browser-languagedetector` (browser `navigator.language`)
4. Fallback: `en`

**Rationale:** localStorage gives instant language switching without waiting for an API call. Profile storage syncs across devices. The browser detector handles first-visit experience.

**Alternatives considered:**
- *localStorage only (no DB):* Doesn't persist across devices or browsers.
- *DB only:* Requires API call before first render, causes flash of wrong language.

### 7. Locale-aware formatting

**Decision:** Use `date-fns` locale support (already installed) for date formatting. For numbers, use `Intl.NumberFormat` (built into all modern browsers).

```ts
import { pl, enUS } from "date-fns/locale";
format(date, "PPP", { locale: currentLocale === "pl" ? pl : enUS });
```

**Rationale:** `date-fns` already supports locale-aware formatting via its `locale` parameter — no new dependency needed. `Intl.NumberFormat` is native and handles comma vs. dot decimal separators correctly.

### 8. Zod validation message localization

**Decision:** Use Zod's `errorMap` feature to provide translated validation messages. Create a custom error map that reads from `i18next` translations.

```ts
const zodErrorMap: ZodErrorMap = (issue, ctx) => {
  const key = `validation.${issue.code}`;
  const translated = i18next.t(key, { defaultValue: ctx.defaultError });
  return { message: translated };
};
z.setErrorMap(zodErrorMap);
```

**Rationale:** Centralizes validation message translation in one place rather than adding `{ message: t("...") }` to every Zod field. The error map is set once at app init.

### 9. PATCH /api/v1/auth/me endpoint

**Decision:** Add a `PATCH /api/v1/auth/me` endpoint to update user profile fields (starting with `preferred_language`). This avoids creating a separate `/users/:id` endpoint just for language preference.

**Rationale:** The `/auth/me` resource already represents "the current user." PATCH semantics allow partial updates. This is extensible for future profile fields (display name, avatar override, etc.).

## Risks / Trade-offs

**[Breaking API change]** → The error envelope format change breaks any client parsing `data.error`. **Mitigation:** The only client is our React frontend, updated in the same change. The `message` field preserves the English text for API debugging tools (curl, Postman).

**[Translation completeness]** → Missing translation keys would show raw keys like `workspace.createButton` in the UI. **Mitigation:** i18next `fallbackLng: 'en'` ensures English is always shown if a Polish key is missing. Add a CI check that compares `en/*.json` and `pl/*.json` key sets.

**[Bundle size]** → Translation JSON adds ~5-15 KB per locale. **Mitigation:** i18next supports lazy loading namespaces, but at two locales and ~200 keys this is negligible. No lazy loading needed now.

**[Stale localStorage]** → User changes language on one device, other device still uses old localStorage value. **Mitigation:** On each login/page load, sync `preferred_language` from the `/auth/me` response to localStorage. Profile is the source of truth; localStorage is a cache.

## Migration Plan

1. **Backend first:** Add `AppError` type, error codes, new error envelope, `preferred_language` column, PATCH endpoint. All in one migration + code change.
2. **Frontend second:** Install i18next, extract strings, update `getErrorMessage()` to use `data.code`, add language switcher.
3. **Rollback:** Revert migration (`ALTER TABLE users DROP COLUMN preferred_language`), revert error envelope to `{"error": "text"}`. Frontend would need to revert to hardcoded strings.

No staged rollout needed — this is a single-SPA app with one backend, deployed atomically.

## Open Questions

- **Polish translations source:** Who provides the Polish translations? Developer-written first pass, then reviewed by a native speaker? Or wait for professional translations?
- **Date format preference:** Should Polish users see `dd.MM.yyyy` (standard Polish) or should the date format be a separate user preference independent of language?
