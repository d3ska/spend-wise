# Recurring Expenses & Subscription Tracker

## Problem

Households have two kinds of spending — *committed* (rent, utilities, insurance, subscriptions, phone bills) and *discretionary* (groceries, dining, shopping). Most people don't know how much of their income is already spoken for before the month starts. Forgotten or unwanted subscriptions waste significant money.

## Solution

Auto-detect recurring patterns from imported bank transactions and surface them as a dedicated view. Turns SpendWise from a ledger (records what happened) into an advisor (tells you what matters).

## Why It Fits SpendWise

- Bank import pipeline and counterparty IBAN tracking already exist — detection piggybacks on this data
- Categorization rule engine already matches patterns — recurring detection is a natural extension
- Dashboard already shows period-over-period comparison — committed vs. discretionary is the next logical insight
- Collaborative — household members jointly review "do we still need this?" in the shared workspace

## Implementation Plan

### Phase 1 — Detection (backend only, no UI)

SQL query grouping transactions by counterparty IBAN + similar amounts over 6 months, looking for ~monthly interval patterns. No new tables — just a new query, service method, and endpoint.

**Output:** JSON endpoint returning detected recurring transactions. Validate detection quality by eyeballing API response against real bank data.

**Effort:** A weekend.

### Phase 2 — Read-only page

Simple `RecurringPage` — a table listing Phase 1 results. Columns: counterparty name, amount, category, frequency, last/next date. No editing, no alerts. Just visibility.

Already useful on its own — open it, see all recurring charges, notice the gym membership you forgot about.

**Effort:** An evening or two.

### Phase 3 — User corrections

Add `recurring_expenses` table so users can confirm, dismiss, or adjust auto-detected results. Mark false positives as "not recurring," manually add ones the algorithm missed. Turns detection from "best guess" into "curated list."

**Effort:** A weekend.

### Phase 4 — Dashboard integration

Single card on existing dashboard: "Monthly committed: X PLN" next to total spending. That one number delivers most of the feature's value.

**Effort:** A couple hours.

### Reassess after Phase 4

Phases 1-4 deliver ~80% of value with ~20% of effort. Alerts, anomaly detection, subscription audit workflow are polish. Ship the core, use it with real data for a few months, let real usage reveal what's actually missing.

**Principle:** Every phase is independently useful. If motivation drops after Phase 2, there's still something valuable. No phase that only pays off "once the whole thing is done."

## Future Enhancements (post-reassessment)

- Change & anomaly alerts ("electricity bill went from 180 to 260 PLN")
- Missing charge detection ("Netflix didn't appear this month")
- Subscription audit workflow (keep / review / cancel)
- Monthly committed vs. discretionary budget breakdown
- Yearly projection of recurring costs
