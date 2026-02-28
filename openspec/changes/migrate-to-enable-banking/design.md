## Context

SpendWise has an existing bank sync feature built on GoCardless Bank Account Data API. The implementation spans `banksync/` (HTTP client), `service/` (connection, linking, sync services), `store/` (DB queries), `handler/` (HTTP endpoints), and `config/` (env vars). GoCardless has discontinued free access for new users, requiring migration to Enable Banking which offers a free personal tier.

The existing architecture uses a `banksync.Client` interface that decouples the API provider from business logic. This interface-based design makes the swap straightforward — only the client implementation and the service methods that directly reference GoCardless-specific concepts (agreement, requisition) need to change.

## Goals / Non-Goals

**Goals:**
- Replace GoCardless with Enable Banking while preserving the existing frontend API contract (zero frontend changes)
- Maintain the same user-facing flow: select bank → redirect to authenticate → accounts discovered → link to workspace → transactions sync
- Support all major Polish banks and Revolut
- Leverage Enable Banking's longer consent duration (up to 180 days per ASPSP)
- Improve dedup robustness with composite fingerprints

**Non-Goals:**
- Adding payment initiation (PIS) — only account information (AIS) is in scope
- Supporting UK banks (Enable Banking doesn't cover UK)
- Changing the sync scheduling strategy or cron interval
- Migrating existing GoCardless connections (users will need to re-authenticate)

## Decisions

### 1. JWT Auth via RSA256 with PEM Key File

**Decision**: Store the RSA private key as a PEM file on disk, referenced by `ENABLEBANKING_KEY_PATH` env var. Generate short-lived JWTs (1 hour) on each API call.

**Rationale**: Enable Banking requires JWT/RSA256 auth (no API key/secret option). The official Go example uses this approach. Short-lived tokens avoid token refresh complexity — just generate a new one per request batch.

**Alternatives considered**:
- Embedding key in env var as base64 — more complex to manage, PEM file is simpler and matches their examples
- Long-lived tokens with refresh — unnecessary complexity since JWT generation is cheap (< 1ms)

### 2. Rename `requisition_id` to `session_id` via Migration

**Decision**: Create a migration that renames the `requisition_id` column to `session_id` and drops+recreates the unique index.

**Rationale**: Enable Banking uses "sessions" where GoCardless used "requisitions". The column stores the same concept (external auth session identifier). Renaming keeps the schema self-documenting.

**Alternatives considered**:
- Keep `requisition_id` column name and just store session IDs — confusing, technical debt
- Add a new column and deprecate old one — unnecessary complexity for an unreleased feature

### 3. Institution Identification: Store ASPSP Name

**Decision**: The `institution_id` column in `bank_connections` will store the Enable Banking ASPSP name (e.g., "mBank", "ING Bank Slaski") since Enable Banking identifies ASPSPs by name+country pair, not by a separate ID.

**Rationale**: Enable Banking's `POST /auth` endpoint requires `aspsp.name` and `aspsp.country`. Storing the name in `institution_id` keeps the schema stable and avoids a migration. The `institution_name` column remains for display purposes (can be the same value or a user-friendly variant).

### 4. Composite Fingerprint for Dedup

**Decision**: Build fingerprint as `sha256(transaction_id + "|" + booking_date + "|" + amount + "|" + currency)` when `transaction_id` is present, or `sha256(booking_date + "|" + amount + "|" + currency + "|" + remittance_info)` when it's absent.

**Rationale**: Enable Banking warns that `transaction_id` is not guaranteed unique across all institutions and may be absent from some banks. A composite fingerprint is more robust. Including amount and date prevents collisions. SHA256 keeps the fingerprint a fixed-length string.

**Alternatives considered**:
- Use `transaction_id` alone — risky, not universally available
- Use all fields — over-sensitive, minor bank-side description changes would cause duplicates

### 5. Dynamic Consent Duration from ASPSP Metadata

**Decision**: When initiating auth, query the ASPSP's `maximum_consent_validity` field and set `access.valid_until` to `min(max_consent, 180 days)`. Store the actual expiry in `auth_expires_at`.

**Rationale**: Different banks have different maximum consent durations. Using the ASPSP-provided value maximizes the time before re-auth while staying within PSD2 limits.

### 6. Rewrite `banksync.Client` Interface

**Decision**: Change the `Client` interface methods to match Enable Banking's API model:
- `GetInstitutions` → `GetASPSPs(ctx, country)` returning ASPSP list
- `CreateAgreement` + `CreateRequisition` → `StartAuth(ctx, aspspName, country, redirectURL, validUntil)` returning auth URL + authorization ID
- `GetRequisition` → `CreateSession(ctx, authorizationCode)` returning session with accounts
- `GetAccountTransactions` → stays similar but with different response mapping

**Rationale**: The interface should reflect the actual provider's API model rather than forcing GoCardless concepts onto Enable Banking.

### 7. Account Discovery: Use Session Response Directly

**Decision**: When `POST /sessions` completes, the response includes the full account list with `uid` (for API access) and `account_id` (IBAN). Store `uid` as `external_id` and populate the `iban` field directly.

**Rationale**: GoCardless only returned account IDs (opaque strings) at the requisition stage, requiring separate calls for details. Enable Banking returns everything in the session response, so we can populate more fields upfront.

## Risks / Trade-offs

- **[No migration path for existing connections]** → Users with active GoCardless connections will need to re-authenticate via Enable Banking. Since the feature hasn't shipped to production yet, this is acceptable. If it had, we'd need a migration period with both providers.

- **[RSA key file management]** → Requires provisioning a PEM file on the server. In containerized deployments, this means mounting a secret volume. Mitigation: document the setup clearly, support reading from env var as fallback.

- **[Transaction ID not guaranteed]** → Some banks may not provide `transaction_id` at all. Mitigation: composite fingerprint falls back to date+amount+description, with 3-day overlap window as additional safety net.

- **[Rate limiting]** → Background sync (without PSU headers) is limited to 4 requests/day per account by banks. Our 6-hour cron (4x/day) exactly fits this limit. Mitigation: handle 429 responses gracefully, log and skip until next cycle.

- **[ASPSP name stability]** → If Enable Banking changes an ASPSP name, existing connections referencing the old name could break reconnection. Mitigation: low risk since ASPSP names are stable identifiers in their system.

## Open Questions

- Should we support reading the RSA private key from an environment variable (base64-encoded) as an alternative to a file path, for platforms where file mounts are difficult?
