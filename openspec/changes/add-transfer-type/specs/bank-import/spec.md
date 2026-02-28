## MODIFIED Requirements

### Requirement: Bank sync maps transaction type with transfer detection
The bank sync service SHALL determine the transaction type by first checking whether the counterparty IBAN matches one of the user's own bank account IBANs. If it matches, the type SHALL be `transfer`. Otherwise, the existing CRDT/DBIT logic SHALL apply.

#### Scenario: Outgoing payment to external account
- **WHEN** a bank transaction has `credit_debit_indicator = "DBIT"`
- **AND** the creditor IBAN does not match any of the user's own IBANs
- **THEN** the transaction type SHALL be `expense`

#### Scenario: Incoming payment from external account
- **WHEN** a bank transaction has `credit_debit_indicator = "CRDT"`
- **AND** the debtor IBAN does not match any of the user's own IBANs
- **THEN** the transaction type SHALL be `income`

#### Scenario: Outgoing payment to own account
- **WHEN** a bank transaction has `credit_debit_indicator = "DBIT"`
- **AND** the creditor IBAN matches one of the user's own bank account IBANs
- **THEN** the transaction type SHALL be `transfer`

#### Scenario: Incoming payment from own account
- **WHEN** a bank transaction has `credit_debit_indicator = "CRDT"`
- **AND** the debtor IBAN matches one of the user's own bank account IBANs
- **THEN** the transaction type SHALL be `transfer`

### Requirement: Bank sync loads user IBANs before processing
The `SyncBankAccount` method SHALL accept a set of the user's own IBANs as a parameter. The `SyncAll` method SHALL load the user's bank accounts once per connection and build the IBAN set, then pass it to each `SyncBankAccount` call.

#### Scenario: IBAN set passed to sync
- **WHEN** `SyncAll` processes a connection belonging to user with IBANs "LT111" and "LT222"
- **THEN** `SyncBankAccount` SHALL receive a set containing both IBANs

#### Scenario: Nil IBANs excluded from set
- **WHEN** a bank account has a nil IBAN
- **THEN** it SHALL NOT be included in the IBAN set
