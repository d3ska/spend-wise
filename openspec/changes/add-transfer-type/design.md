## Context

Users with multiple bank accounts see internal transfers counted as both an expense (source) and income (destination), inflating summaries. The Enable Banking API returns `Creditor` and `Debtor` fields on each transaction, but our `PartyIdentification` struct currently only captures `Name` — not the counterparty's IBAN. We need to extend it to extract the IBAN, then compare it against the user's own bank accounts during sync.

The spending summary queries already filter on `transaction_type = 'expense'`, so adding a `transfer` type automatically excludes transfers from summaries with no query changes needed.

## Goals / Non-Goals

**Goals:**
- Add `transfer` as a third `TransactionType` (alongside `expense`/`income`)
- Auto-detect internal transfers during bank sync by matching counterparty IBAN against user's own bank account IBANs
- Transfers excluded from spending summaries (already handled by existing `WHERE transaction_type = 'expense'`)
- Users can view transfers via the type filter on the transactions page
- Apply Rules skips `transfer` transactions

**Non-Goals:**
- Refund detection (phase 2 — separate change)
- Matching transfers across workspaces (transfers are detected per-user, applied per-workspace)
- Retroactive detection of transfers already synced as expense/income (user can re-sync or we can add a one-time migration script later)

## Decisions

### 1. Extend `PartyIdentification` with IBAN

**Decision:** Add `Account *AccountIdentification` to `banksync.PartyIdentification`.

Enable Banking's transaction response includes `creditor.account.iban` and `debtor.account.iban`. Our struct currently only captures `Name`. We extend it to include the counterparty's account info.

```go
type PartyIdentification struct {
    Name    string                  `json:"name"`
    Account *AccountIdentification `json:"account"`
}
```

**Why not a separate field?** The Enable Banking API nests account under creditor/debtor. Matching the API shape avoids extra mapping.

### 2. IBAN lookup set built per-sync-cycle

**Decision:** Before syncing each connection, load all IBANs from the user's bank accounts into a `map[string]bool`. Pass this set to `mapTransaction`.

The `SyncAll` method already iterates connections by user. We load the user's bank accounts once per connection (not per transaction) and build an IBAN set. In `mapTransaction`, if the counterparty IBAN is in the set, the transaction type becomes `transfer`.

**Alternative considered:** Query the DB for each transaction's counterparty IBAN. Rejected — N+1 queries per sync.

### 3. Which IBAN to check — creditor vs debtor

**Decision:** For outgoing transactions (DBIT / negative amount), check `creditor.account.iban`. For incoming (CRDT / positive), check `debtor.account.iban`. The counterparty is always the "other side".

If the counterparty IBAN matches one of the user's own IBANs → it's an internal transfer.

### 4. DB migration: add value to enum

**Decision:** `ALTER TYPE transaction_type ADD VALUE 'transfer'`. This is a forward-only migration — PostgreSQL does not support removing enum values, so the down migration is a no-op with a comment.

### 5. Summary queries need no changes

The existing summary queries (`TotalSpent`, `SpentByCategory`, `SpentByParticipant`) all filter `WHERE t.transaction_type = 'expense'`. Since `transfer` is a distinct value, transfers are already excluded. No SQL changes needed.

### 6. Apply Rules skips transfers

**Decision:** In `ApplyRules` service method, skip entries whose transaction type is `transfer`. Internal transfers don't need categorization.

This requires the `ListAllTransactionEntries` query to also return `transaction_type`, or we filter in Go after loading.

### 7. Frontend type filter

**Decision:** Add "Transfers" as a fourth option in the type toggle button group: `Expenses | Income | Transfers | All`. Transfer rows get a distinct badge (e.g. "Transfer" with an arrow icon).

## Risks / Trade-offs

**[Risk] Not all banks provide counterparty IBAN** → Some banks may only return a name, not an account identifier. In that case, the transfer won't be auto-detected and will remain as expense/income. Mitigation: users can manually change the type to `transfer` via the edit dialog. This is acceptable for phase 1.

**[Risk] Existing transactions won't be reclassified** → Transactions already synced as expense/income before this change stay as-is. Mitigation: could add a one-time "detect transfers" action or document that a re-sync after clearing `last_synced_at` would pick them up. Not in scope for this change.

**[Risk] PostgreSQL enum values can't be removed** → The `ALTER TYPE ADD VALUE` migration is irreversible. Mitigation: `transfer` is a permanent concept, not experimental. The down migration is a no-op.

**[Risk] Same IBAN used for different currencies (Revolut)** → A single IBAN may back multiple currency sub-accounts. This is fine — if the IBAN matches, it's still the same user's account regardless of currency.
