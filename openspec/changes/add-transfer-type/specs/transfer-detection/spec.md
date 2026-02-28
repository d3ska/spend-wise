## ADDED Requirements

### Requirement: Counterparty IBAN extraction from bank transactions
The `banksync.PartyIdentification` struct SHALL include an `Account *AccountIdentification` field to capture the counterparty's IBAN from Enable Banking transaction data. The `AccountIdentification` struct already exists and contains an `IBAN` field.

#### Scenario: Creditor has IBAN
- **WHEN** a bank transaction is fetched with `creditor.account.iban` populated
- **THEN** the `Creditor.Account.IBAN` field SHALL contain the creditor's IBAN

#### Scenario: Creditor has no account info
- **WHEN** a bank transaction is fetched without `creditor.account`
- **THEN** the `Creditor.Account` field SHALL be nil

### Requirement: Own-IBAN set built per sync connection
During bank sync, the system SHALL load all bank account IBANs belonging to the connection's user and build an in-memory set before processing transactions. This set SHALL be built once per connection, not per transaction.

#### Scenario: User has multiple bank accounts
- **WHEN** a user has 3 bank accounts with IBANs "LT111", "LT222", and one with nil IBAN
- **THEN** the IBAN set SHALL contain "LT111" and "LT222" (nil IBANs excluded)

### Requirement: Auto-detect internal transfers by counterparty IBAN
During bank sync, when mapping a transaction, the system SHALL check the counterparty's IBAN against the user's own IBAN set. If the counterparty IBAN matches, the transaction type SHALL be set to `transfer` instead of `expense` or `income`.

#### Scenario: Outgoing transfer to own account
- **WHEN** a bank transaction has `credit_debit_indicator = "DBIT"` (outgoing)
- **AND** the `creditor.account.iban` matches one of the user's own bank account IBANs
- **THEN** the transaction type SHALL be `transfer`

#### Scenario: Incoming transfer from own account
- **WHEN** a bank transaction has `credit_debit_indicator = "CRDT"` (incoming)
- **AND** the `debtor.account.iban` matches one of the user's own bank account IBANs
- **THEN** the transaction type SHALL be `transfer`

#### Scenario: Counterparty IBAN does not match
- **WHEN** the counterparty IBAN does not match any of the user's own bank account IBANs
- **THEN** the transaction type SHALL follow the existing logic (expense for DBIT, income for CRDT)

#### Scenario: No counterparty IBAN available
- **WHEN** the counterparty has no IBAN (Account is nil or IBAN is empty)
- **THEN** the transaction type SHALL follow the existing logic (no transfer detection)

### Requirement: Apply Rules skips transfer transactions
The Apply Rules feature SHALL skip entries belonging to transactions with `type = 'transfer'`. Internal transfers do not need categorization.

#### Scenario: Transfer entry skipped during apply rules
- **WHEN** Apply Rules is executed on a workspace
- **AND** a transaction with `type = 'transfer'` exists
- **THEN** its entries SHALL NOT be re-categorized

#### Scenario: Non-transfer entries still processed
- **WHEN** Apply Rules is executed on a workspace
- **AND** both transfer and expense transactions exist
- **THEN** only expense and income transaction entries SHALL be evaluated against rules
