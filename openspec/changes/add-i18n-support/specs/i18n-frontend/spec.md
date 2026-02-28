## ADDED Requirements

### Requirement: i18next initialization with language detection
The frontend SHALL initialize `i18next` with `react-i18next` and `i18next-browser-languagedetector`. The configuration SHALL set `fallbackLng` to `"en"`, support `"en"` and `"pl"` locales, and use namespace-based resource loading. The detection order SHALL be: `localStorage` key `"i18nextLng"` → browser `navigator.language` → fallback `"en"`.

#### Scenario: i18next initializes with English default
- **WHEN** the application loads with no localStorage language preference and browser language is not Polish
- **THEN** i18next SHALL initialize with `lng: "en"`
- **AND** all UI text SHALL render in English

#### Scenario: i18next detects Polish browser language
- **WHEN** the application loads with no localStorage preference and `navigator.language` starts with `"pl"`
- **THEN** i18next SHALL initialize with `lng: "pl"`
- **AND** all UI text SHALL render in Polish

#### Scenario: localStorage preference takes priority
- **WHEN** localStorage contains `i18nextLng: "pl"` and browser language is English
- **THEN** i18next SHALL initialize with `lng: "pl"`

#### Scenario: Missing Polish key falls back to English
- **WHEN** a translation key exists in `en` but not in `pl` and the active locale is `pl`
- **THEN** i18next SHALL display the English translation (not the raw key)

### Requirement: Translation namespace files for English and Polish
The frontend SHALL maintain JSON translation files organized by namespace under `frontend/src/locales/{locale}/`. Both `en` and `pl` directories SHALL contain identical namespaces: `common`, `errors`, `auth`, `workspace`, `transaction`, `validation`.

#### Scenario: English translation files exist
- **WHEN** the application is built
- **THEN** the following files SHALL exist: `locales/en/common.json`, `locales/en/errors.json`, `locales/en/auth.json`, `locales/en/workspace.json`, `locales/en/transaction.json`, `locales/en/validation.json`

#### Scenario: Polish translation files exist
- **WHEN** the application is built
- **THEN** the following files SHALL exist: `locales/pl/common.json`, `locales/pl/errors.json`, `locales/pl/auth.json`, `locales/pl/workspace.json`, `locales/pl/transaction.json`, `locales/pl/validation.json`

#### Scenario: Key parity between locales
- **WHEN** the `en` and `pl` translation files are compared
- **THEN** every key present in any `en/{namespace}.json` SHALL also be present in the corresponding `pl/{namespace}.json`

### Requirement: errors namespace maps backend error codes
The `errors.json` namespace SHALL contain a key for every backend error code defined in `model/`. The key SHALL match the backend error code exactly (UPPER_SNAKE_CASE). The value SHALL be the localized user-facing message.

#### Scenario: Backend error code mapped to English
- **WHEN** the active locale is `en`
- **THEN** `t("errors:WORKSPACE_NOT_FOUND")` SHALL return `"Workspace not found"`

#### Scenario: Backend error code mapped to Polish
- **WHEN** the active locale is `pl`
- **THEN** `t("errors:WORKSPACE_NOT_FOUND")` SHALL return the Polish translation (e.g., `"Nie znaleziono przestrzeni roboczej"`)

### Requirement: All hardcoded UI strings replaced with t() calls
Every user-visible hardcoded string in React components SHALL be replaced with a `t()` call referencing the appropriate namespace and key. This includes button labels, dialog titles, placeholder text, toast messages, and static headings.

#### Scenario: Button labels use translations
- **WHEN** a button with text like "Save", "Cancel", or "Delete" is rendered
- **THEN** the text SHALL come from `t("common:save")`, `t("common:cancel")`, `t("common:delete")` respectively

#### Scenario: Login page uses translations
- **WHEN** the login page is rendered
- **THEN** "Sign in with Google" and "Sign in with GitHub" SHALL come from `t("auth:signInGoogle")` and `t("auth:signInGithub")`

#### Scenario: Toast error messages use translations
- **WHEN** a toast error is shown in response to a failed API call
- **THEN** the message SHALL be derived from the backend error code via the `errors` namespace, not hardcoded English

### Requirement: Updated error extraction utility
The `getErrorMessage` function in `lib/errors.ts` SHALL be updated to extract the `code` field from API error responses and return a translated message via `i18next.t("errors:{code}")`. If no `code` field is present, it SHALL fall back to the `message` field, then to the provided fallback string.

#### Scenario: Error with code field
- **WHEN** an API response contains `{"code": "WORKSPACE_NOT_FOUND", "message": "workspace not found"}`
- **THEN** `getErrorMessage` SHALL return the result of `t("errors:WORKSPACE_NOT_FOUND")`

#### Scenario: Error without code field (legacy)
- **WHEN** an API response contains `{"message": "unexpected error"}` with no `code` field
- **THEN** `getErrorMessage` SHALL return `"unexpected error"`

#### Scenario: Non-API error
- **WHEN** the error is not an Axios error
- **THEN** `getErrorMessage` SHALL return the fallback string

### Requirement: Language switcher component
The frontend SHALL provide a `LanguageSwitcher` component that allows the user to toggle between English and Polish. Changing the language SHALL update `i18next` language, persist to `localStorage`, and call `PATCH /api/v1/auth/me` to sync the preference to the server (if the user is authenticated).

#### Scenario: Switch from English to Polish
- **WHEN** the user selects Polish in the language switcher
- **THEN** `i18next.changeLanguage("pl")` SHALL be called
- **AND** `localStorage` key `i18nextLng` SHALL be set to `"pl"`
- **AND** all rendered text SHALL update to Polish without a page reload

#### Scenario: Authenticated user preference synced
- **WHEN** an authenticated user switches language
- **THEN** a `PATCH /api/v1/auth/me` request SHALL be sent with `{"preferred_language": "pl"}`

#### Scenario: Unauthenticated user preference stored locally
- **WHEN** a non-authenticated user switches language (e.g., on the login page)
- **THEN** the preference SHALL be stored in `localStorage` only (no API call)

### Requirement: Locale-aware date formatting
All date display in the frontend SHALL use `date-fns` with the locale matching the active `i18next` language. English SHALL use `enUS` locale, Polish SHALL use `pl` locale.

#### Scenario: Date formatted in English
- **WHEN** the active language is `en` and a date `2026-03-15` is displayed
- **THEN** it SHALL be formatted using `date-fns` `enUS` locale (e.g., `"Mar 15, 2026"` or `"03/15/2026"` depending on format string)

#### Scenario: Date formatted in Polish
- **WHEN** the active language is `pl` and a date `2026-03-15` is displayed
- **THEN** it SHALL be formatted using `date-fns` `pl` locale (e.g., `"15 mar 2026"` or `"15.03.2026"`)

### Requirement: Locale-aware number formatting
Monetary amounts and numeric values displayed in the UI SHALL use `Intl.NumberFormat` with the active locale. English SHALL use dot decimal separator, Polish SHALL use comma decimal separator.

#### Scenario: Amount formatted in English
- **WHEN** the active language is `en` and amount `1234.56` is displayed
- **THEN** it SHALL render as `"1,234.56"` (dot decimal, comma thousands)

#### Scenario: Amount formatted in Polish
- **WHEN** the active language is `pl` and amount `1234.56` is displayed
- **THEN** it SHALL render as `"1 234,56"` (comma decimal, space thousands)

### Requirement: Zod validation messages use translations
The frontend SHALL configure a global Zod error map that translates validation messages via `i18next`. Validation error codes (e.g., `too_small`, `invalid_type`) SHALL map to keys in the `validation` namespace.

#### Scenario: Required field validation in English
- **WHEN** a required field is left empty and the active language is `en`
- **THEN** the validation message SHALL be the English translation from `validation` namespace (e.g., `"This field is required"`)

#### Scenario: Required field validation in Polish
- **WHEN** a required field is left empty and the active language is `pl`
- **THEN** the validation message SHALL be the Polish translation (e.g., `"To pole jest wymagane"`)

### Requirement: Language synced from user profile on login
When the user authenticates and the `GET /api/v1/auth/me` response includes `preferred_language`, the frontend SHALL set `i18next` language to that value and update `localStorage`. This ensures cross-device consistency.

#### Scenario: Profile language overrides localStorage
- **WHEN** the user logs in with `preferred_language: "pl"` in their profile and `localStorage` has `"en"`
- **THEN** i18next SHALL switch to `"pl"` and localStorage SHALL be updated to `"pl"`

#### Scenario: Profile language not set
- **WHEN** the user logs in with `preferred_language: "en"` (default) and localStorage has `"pl"`
- **THEN** i18next SHALL keep `"pl"` (localStorage is more recent user intent)
- **AND** a `PATCH /api/v1/auth/me` SHALL be sent with `{"preferred_language": "pl"}` to sync
