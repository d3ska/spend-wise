## Context

The application has transactions with entries, each entry assigned to a category and optionally a participant. Users need a dashboard view showing spending aggregates for any arbitrary date range. The `(workspace_id, date)` composite index from migration 000004 already optimizes these range queries.

Funding (contributions tracking) does not exist yet — it will be added in a future change. This design computes spent amounts only; funded/balance fields will be zero-valued until the funding change lands.

## Goals / Non-Goals

**Goals:**
- Provide a single summary endpoint that returns total spent, per-category breakdown, and per-participant breakdown for any date range
- Keep the backend date-range agnostic — the UI decides granularity (month, week, day, custom)
- Compute everything on-the-fly via SQL aggregation — no materialized state

**Non-Goals:**
- Funding/balance computation (deferred to the funding change)
- Caching or materialized views (unnecessary at household scale)
- Time-series or trend data (out of scope — this is a single-range snapshot)

## Decisions

### 1. On-the-fly SQL aggregation, no stored summaries

Run three aggregate queries (total spent, by category, by participant) on each request. At household scale (hundreds to low thousands of transactions/month), PostgreSQL computes these in single-digit milliseconds using the existing `(workspace_id, date)` index. Storing pre-computed summaries would add write-time complexity, staleness risk, and migration overhead for zero perceptible latency gain.

**Alternative considered:** Materialized view refreshed on transaction writes — rejected because it adds trigger/refresh complexity and the read performance is already excellent without it.

### 2. Three separate queries, not one mega-join

Use three focused queries (total, by-category, by-participant) instead of a single query with complex grouping sets or unions. Each query is simple, independently testable, and maps cleanly to its result type. The overhead of three round-trips to the same connection pool is negligible compared to query execution time.

**Alternative considered:** Single query with `GROUPING SETS` — rejected because the result shape is heterogeneous (different columns per grouping) and mapping it to Go structs would be awkward.

### 3. SummaryStore on pgxpool.Pool, not sqlc DBTX

Like TransactionStore, the SummaryStore takes `*pgxpool.Pool` directly. The aggregate queries return custom result shapes (category name + sum, participant ID + sum) that don't map 1:1 to any single table, so we use manual `pool.Query` with row scanning rather than sqlc-generated code.

**Alternative considered:** Define sqlc queries for aggregates — rejected because sqlc struggles with aggregate results that join multiple tables and return computed columns; manual queries are clearer here.

### 4. Summary read model returned directly, no DTO layer

The `model.SpendingSummary` struct is already a read model (never stored). The service returns it directly and the handler maps it to JSON. No intermediate DTO is needed since the read model IS the output shape.

### 5. Funded/Balance fields stub to zero until funding lands

`ParticipantSpending` includes `Funded` and `Balance` fields typed as `model.Money`, but they will be `Zero("PLN")` until the funding change adds the funding store and queries. This avoids breaking the API contract later — the JSON shape is stable from day one.

## Risks / Trade-offs

- **[Funding not yet available]** → ParticipantSpending.Funded and Balance will be zero. Mitigation: fields exist in the model and JSON response; the funding change will populate them by querying the funding store.
- **[No pagination on summary]** → If a workspace has hundreds of categories, the by-category list could be large. Mitigation: unlikely at household scale; add `LIMIT` later if needed.
- **[Currency assumption]** → Aggregation sums amounts assuming single currency per workspace (PLN default). Mitigation: all entries in a workspace use the same currency; multi-currency support is out of scope.
