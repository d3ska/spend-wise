## Context

SpendWise currently supports manual transaction entry and CSV import. Users want automatic bank transaction syncing to reduce friction. GoCardless Bank Account Data API (formerly Nordigen) provides free PSD2 access to European bank accounts — up to 24 months of history, 90 days of continuous access per authorization, with date-range filtering on the transactions endpoint.

The current codebase follows a flat `model ← store ← service ← handler ← cmd/api` chain. Transactions already support `source` (`manual`|`import`) and `fingerprint`-based deduplication. The composition root in `cmd/api/main.go` wires everything together with graceful shutdown.

## Goals / Non-Goals

**Goals:**
- Allow users to connect European bank accounts via GoCardless and automatically sync booked transactions into workspaces
- Support multiple banks per user and multiple bank accounts per workspace
- Let users reuse a single bank authentication across multiple workspaces (no re-auth)
- Provide clear UX around the 90-day PSD2 re-authentication requirement
- Sync transactions on a background schedule (~every 6 hours) with minimal API calls

**Non-Goals:**
- Real-time transaction notifications (bank rate limits make this impractical — max 4 calls/day/account)
- Premium/enriched GoCardless data (stay on free `/api/v2/` endpoints only)
- Auto-categorization of bank transactions (future enhancement; transactions land as "Uncategorized")
- Non-EEA bank support (PSD2 is Europe-only)
- Balance syncing (transactions only in this iteration)

## Decisions

### 1. GoCardless client as a standalone `banksync` package

**Decision:** Create a new `backend/banksync/` package containing the GoCardless HTTP client, separate from the existing `store`/`service`/`handler` layers.

**Why:** The GoCardless API is an external HTTP dependency with its own auth token lifecycle (24h access token, 30-day refresh token). Isolating it keeps the external API boundary clean and testable via interface. The package exposes a `Client` interface consumed by the bank sync service.

**Alternatives considered:**
- Embed GoCardless calls directly in the service layer → rejected because it couples external API details to business logic and makes testing harder.
- Use a third-party Go SDK → none exists with active maintenance; the API is simple enough (6 endpoints) to wrap directly.

### 2. Bank connections are user-owned, linked to workspaces via bank accounts

**Decision:** Three new tables: `bank_connections` (user-scoped), `bank_accounts` (child of connection), `workspace_bank_accounts` (join table).

**Why:** A user authenticates with their bank once. The resulting connection yields one or more bank accounts. The user then picks which accounts feed into which workspaces. This avoids re-authentication when adding the same bank to a second workspace.

**Data model:**
```
bank_connections
├── id              BIGSERIAL PK
├── user_id         FK → users
├── institution_id  TEXT (GoCardless institution ID, e.g. "PKO_BP_BPKOPLPW")
├── institution_name TEXT ("PKO BP")
├── requisition_id  TEXT (GoCardless requisition UUID)
├── status          TEXT (active | expired | revoked)
├── auth_expires_at TIMESTAMPTZ
├── created_at      TIMESTAMPTZ
└── updated_at      TIMESTAMPTZ

bank_accounts
├── id                  BIGSERIAL PK
├── bank_connection_id  FK → bank_connections ON DELETE CASCADE
├── external_id         TEXT UNIQUE (GoCardless account UUID)
├── iban                TEXT (nullable, not all banks expose this)
├── name                TEXT (user-editable display name)
├── last_synced_at      TIMESTAMPTZ (nullable, null = never synced)
├── created_at          TIMESTAMPTZ
└── updated_at          TIMESTAMPTZ

workspace_bank_accounts
├── workspace_id    FK → workspaces ON DELETE CASCADE
├── bank_account_id FK → bank_accounts ON DELETE CASCADE
└── UNIQUE(workspace_id, bank_account_id)
```

**Alternatives considered:**
- Workspace-owned connections (user re-auths per workspace) → rejected because it wastes the 50 requisition/month free tier limit and is bad UX.
- Flat single table → rejected because the 1:N relationship (connection → accounts) is real and the N:M relationship (accounts ↔ workspaces) requires a join table.

### 3. Extend transactions table with `bank_account_id` and new source enum value

**Decision:** Add nullable `bank_account_id BIGINT FK → bank_accounts` to `transactions`. Add `'bank'` to the `transaction_source` enum.

**Why:** This tracks provenance — which bank account a transaction came from. Nullable because manual/import transactions have no bank origin. The new enum value distinguishes bank-synced from other sources for filtering and display.

**Migration approach:** Single migration that adds the enum value and column. Non-breaking — existing rows get NULL for `bank_account_id`.

### 4. Smart date-range sync with fingerprint safety net

**Decision:** Each sync fetches transactions from `(last_synced_at - 3 days)` to `today` using GoCardless `date_from`/`date_to` query params. Deduplication uses `ON CONFLICT (workspace_id, fingerprint) DO NOTHING` where fingerprint = GoCardless `internalTransactionId`.

**Why:** Date-range filtering minimizes data transfer and respects the 4-call/day bank rate limit. The 3-day overlap buffer handles banks that backdate transactions. The fingerprint ensures idempotency even with overlapping date ranges. First sync uses `date_from = today - 3 months`.

**Alternatives considered:**
- Fetch all transactions every time → wasteful, hits rate limits, processes thousands of duplicates.
- Rely solely on fingerprint without date filtering → works for correctness but transfers unnecessary data.

### 5. Background cron via `robfig/cron/v3` in the main process

**Decision:** Run a cron scheduler in-process using `robfig/cron/v3`, started alongside the HTTP server in `cmd/api/main.go`. Schedule: every 6 hours. Stopped during graceful shutdown.

**Why:** Keeps the deployment simple (single binary). 6-hour interval = 4 calls/day/account which matches the bank rate limit floor. The cron library is mature, supports context cancellation, and integrates cleanly with the existing graceful shutdown pattern.

**Alternatives considered:**
- External cron job (Kubernetes CronJob, systemd timer) → adds deployment complexity for a simple periodic task.
- Polling loop with `time.Ticker` → works but `robfig/cron` is more expressive and battle-tested.

### 6. GoCardless OAuth flow via redirect-based requisition

**Decision:** The connection flow is:
1. Frontend calls `GET /api/v1/banks?country=PL` → backend proxies GoCardless institutions list
2. User picks a bank → Frontend calls `POST /api/v1/bank-connections` with `institution_id`
3. Backend creates GoCardless end-user agreement + requisition, returns the bank auth `link`
4. Frontend redirects user to the bank auth page
5. Bank redirects back to our frontend callback URL (e.g. `/bank/callback?ref={requisition_id}`)
6. Frontend calls `POST /api/v1/bank-connections/{id}/complete` → backend fetches requisition status, retrieves account IDs, stores them

**Why:** This follows GoCardless's standard flow. The redirect URL points to the frontend (not backend) so the SPA can show appropriate UI states. The "complete" step is separated because the bank redirect is asynchronous.

### 7. Auth expiry tracking with tiered notifications

**Decision:** Store `auth_expires_at` on `bank_connections` (calculated as `created_at + agreement.access_valid_for_days`, default 90 days). Frontend checks expiry state when loading workspace settings and on dashboard mount.

**Tiers:**
- `> 14 days remaining`: Green status in settings ("Connected, expires Mar 15")
- `<= 14 days remaining`: Yellow warning banner in sidebar
- `expired`: Red badge + modal prompting reconnection

**Why:** No backend push notifications needed — the frontend can compute state from `auth_expires_at`. The cron job also checks expiry and skips expired connections (no wasted API calls).

### 8. Transactions land with default "Uncategorized" category, no review

**Decision:** Bank-synced transactions are created immediately as single-entry transactions with the workspace's "Uncategorized" category. No pending/review state.

**Why:** Bank data is authoritative (booked = confirmed by bank). Adding a review gate creates friction that defeats the purpose of automatic sync. Users can recategorize at their leisure. This requires each workspace to have an "Uncategorized" category — either seeded on workspace creation or created on first bank sync.

## Risks / Trade-offs

**[GoCardless free tier changes]** → GoCardless could change pricing or limits. Mitigation: the integration is behind a feature boundary (env vars). The 50 requisition/month limit is generous for personal use but could be an issue if the app scales.

**[Bank rate limits vary]** → Some banks allow only 4 calls/day; others are more generous. We use the conservative floor. → Mitigation: 6-hour interval is safe. If a bank is stricter, GoCardless returns 429 and we retry next cycle.

**[90-day re-auth friction]** → Users may find quarterly re-auth annoying. → Mitigation: tiered notifications give ample warning. Re-auth is the same flow as initial connection (just a new requisition for the same institution).

**[No auto-categorization]** → All bank transactions land as "Uncategorized", which could feel incomplete. → Mitigation: this is explicitly a non-goal for v1. The existing rules system could later auto-categorize based on description patterns.

**[Single-binary cron]** → If the process restarts mid-sync, partial data could be inserted. → Mitigation: each transaction insert is idempotent (fingerprint dedup). On restart, the next sync picks up where it left off.

**[GoCardless API availability]** → External dependency could be down. → Mitigation: sync failures are logged and retried next cycle. Existing transactions are unaffected.

## Open Questions

- Should we expose a manual "Sync Now" button in the UI? It's simple to add but could lead to users hitting rate limits if they spam it. Consider adding a cooldown (e.g., 1 sync per hour per account).
- Should the "Uncategorized" category be auto-created per workspace, or should we require it exists before enabling bank sync? Auto-creation is simpler but adds implicit behavior.
