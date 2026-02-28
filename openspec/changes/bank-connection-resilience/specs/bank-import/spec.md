## ADDED Requirements

### Requirement: IBAN tracking on synced transactions
The bank sync flow SHALL populate the `iban` field on each transaction from the source bank account's IBAN when syncing transactions from Enable Banking.

#### Scenario: Sync populates IBAN
- **WHEN** a transaction is synced from a bank account that has an IBAN
- **THEN** the transaction's `iban` field SHALL be set to the bank account's IBAN

#### Scenario: Sync with no IBAN on bank account
- **WHEN** a transaction is synced from a bank account that has no IBAN (NULL)
- **THEN** the transaction's `iban` field SHALL be NULL
