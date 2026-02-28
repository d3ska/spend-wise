## Context

This is the first cross-cutting change: it introduces two new entities (Funding, CategorizationRule) and modifies two existing services (SummaryService for funding integration, TransactionService for auto-categorization during import). The user requested that rules support enable/disable toggling in addition to standard CRUD.

Existing infrastructure: the `(workspace_id, date)` composite index on transactions supports the summary queries. The summary endpoint currently stubs funded/balance to zero — this change populates them with real funding data.

## Goals / Non-Goals

**Goals:**
- Track participant contributions (funding) per workspace per month with upsert semantics
- Integrate funding into the spending summary (Funded and Balance fields become real)
- Provide tiered categorization rules (workspace > user > system) with pattern matching
- Auto-categorize imported transactions using rule resolution
- Allow rules to be enabled/disabled without deleting them

**Non-Goals:**
- Regex-based pattern matching (use simple case-insensitive `LIKE`/`ILIKE` for now)
- Real-time rule re-application to existing transactions (rules apply only at import time)
- Multi-currency funding (all funding uses workspace default currency PLN)
- Rule conflict resolution UI (first match wins, ordered by scope then priority)

## Decisions

### 1. Funding uses upsert (INSERT ON CONFLICT UPDATE)

A participant's funding for a given month is a single value, not an event stream. `PUT /fundings` with `(workspace_id, user_id, year_month)` either inserts or updates. This avoids duplicate entries and simplifies the UI — just "set my contribution for February to 3000 PLN."

**Alternative considered:** Append-only funding events with SUM — rejected because monthly contribution is inherently a single value, not cumulative events.

### 2. year_month stored as TEXT in "YYYY-MM" format

Storing year_month as a TEXT column (`"2026-02"`) is simple, sortable, and needs no custom type. The funding table's unique constraint `(workspace_id, user_id, year_month)` ensures one entry per user per month.

**Alternative considered:** Separate `year INT` + `month INT` columns — rejected because it complicates the unique constraint and sorting.

### 3. SummaryService gets a FundingStore dependency

The summary service already takes a SummaryStore and WorkspaceStore. Adding FundingStore as a third dependency keeps the pattern consistent. The service queries funding by date range (converting from/to into year_month strings) and merges funded amounts into ParticipantSpending.

**Alternative considered:** Move funding aggregation into SummaryStore SQL — rejected because funding is keyed by year_month while spending is keyed by date, mixing them in one query would be complex and fragile.

### 4. Rule `enabled` boolean field with default true

Rules have an `enabled BOOLEAN NOT NULL DEFAULT TRUE` column. Disabled rules are skipped during resolution but preserved in the database for re-enabling later. The `PATCH /rules/{ruleID}/toggle` endpoint flips the flag. This is simpler than soft-delete and makes the intent explicit.

**Alternative considered:** Soft-delete with `deleted_at` — rejected because the user wants to explicitly toggle, and "disabled" is semantically different from "deleted."

### 5. Pattern matching uses ILIKE (case-insensitive)

Rule resolution uses `description ILIKE '%' || pattern || '%'` for simplicity. At household scale with dozens of rules, this is fast enough. If needed later, `pg_trgm` with GIN indexes can be added.

**Alternative considered:** Regex (`~*`) — rejected as overpowered for household use and harder for users to write.

### 6. Rule resolution: scope precedence then priority DESC

Rules are fetched for a workspace ordered by: scope level (workspace=3, user=2, system=1) DESC, then priority DESC. First matching enabled rule wins. This gives workspace-specific rules highest precedence, personal rules next, and system defaults last.

**Alternative considered:** Single flat priority number — rejected because tiered scoping is the whole point of the rule system. A workspace rule at priority 1 should still beat a system rule at priority 100.

### 7. FundingStore and RuleStore use sqlc (not manual queries)

Unlike SummaryStore (which uses manual queries for aggregate shapes), Funding and Rule entities map cleanly to single-table CRUD. sqlc-generated code works well here.

**Alternative considered:** Manual queries like SummaryStore — rejected because sqlc handles these straightforward queries perfectly.

## Risks / Trade-offs

- **[Cross-service modification]** → SummaryService and TransactionService are modified. Mitigation: changes are additive (new dependency, new call) not behavioral changes to existing logic.
- **[Rule pattern injection]** → `ILIKE` with user-provided patterns could be slow if patterns contain `%` or `_`. Mitigation: at household scale this is negligible; can add pattern sanitization later if needed.
- **[Funding date-range overlap]** → If a summary date range spans partial months, the full month's funding is included. Mitigation: this is acceptable — funding is monthly granularity by design. The UI will typically query full months.
- **[Rule ordering opacity]** → Users may not understand why one rule matches over another. Mitigation: rules endpoint returns the priority and scope, and the UI can display them sorted by resolution order.
