## Why

SpendWise targets Polish-speaking households as its primary user base, but the entire application — error messages, validation text, UI labels, date formatting — is hardcoded in English with no localization infrastructure. Adding internationalization (i18n) for English and Polish enables the app to serve its core audience in their native language while establishing a scalable pattern for future locales.

## What Changes

- **Backend: structured error codes** — Replace plain-text `{"error": "message"}` responses with machine-readable error codes (`{"code": "WORKSPACE_NOT_FOUND", "message": "..."}`) so the frontend can map codes to localized strings instead of displaying raw server text.
- **Backend: Accept-Language middleware** — Parse the `Accept-Language` request header in Chi middleware, resolve to a supported locale (`en`, `pl`), and propagate it via request context.
- **Backend: server-side message catalog** — Provide a Go translation layer so backend-rendered messages (e.g., email templates, SSR fallback text) can be emitted in the negotiated locale.
- **Frontend: react-i18next integration** — Install `i18next` + `react-i18next`, extract all hardcoded UI strings (~100+) into JSON translation files (`en.json`, `pl.json`), and wrap the app in an `I18nextProvider`.
- **Frontend: language switcher** — Add a UI control that lets users switch between English and Polish at runtime, persisted to `localStorage` and synced to their user profile.
- **Frontend: locale-aware formatting** — Configure `date-fns` locale and number formatting to respect the active language (e.g., Polish date format `dd.MM.yyyy`, comma decimal separator).
- **User model: preferred locale** — Add an optional `preferred_language` column to the `users` table so the selected language persists across devices. Defaults to `en`.
- **API contract: Content-Language header** — Return `Content-Language` in every response so clients know which locale the server used.

## Capabilities

### New Capabilities

- `i18n-backend`: Backend internationalization infrastructure — error code enum, locale negotiation middleware, server-side message catalog, Content-Language response header.
- `i18n-frontend`: Frontend internationalization infrastructure — react-i18next setup, translation JSON files (en/pl), language switcher component, locale-aware date/number formatting.

### Modified Capabilities

- `user-management`: Add optional `preferred_language` field (VARCHAR, default `'en'`) to User model, persisted in DB, included in profile GET/PATCH responses.
- `http-server`: Add `Accept-Language` parsing middleware and `Content-Language` response header to the Chi router middleware stack.

## Impact

- **API responses** — Error envelope changes from `{"error": "text"}` to `{"code": "ERROR_CODE", "message": "text"}`. **BREAKING** for any clients parsing the `error` field directly. Frontend error extraction utility (`lib/errors.ts`) must be updated.
- **Database** — New migration: `ALTER TABLE users ADD COLUMN preferred_language VARCHAR(5) NOT NULL DEFAULT 'en'`.
- **Dependencies** — Backend: add a Go i18n library (e.g., `nicksnyder/go-i18n/v2`) or lightweight custom catalog. Frontend: add `i18next`, `react-i18next`, `i18next-browser-languagedetector`.
- **Handler layer** — Every `respondError` call site gains an error code; `handleServiceError` maps sentinel errors to codes.
- **Frontend components** — All ~100+ hardcoded strings replaced with `t('key')` calls; new translation JSON files under `frontend/src/locales/`.
- **Testing** — New table-driven tests for locale negotiation, error code mapping, and translation completeness (missing-key detection).
