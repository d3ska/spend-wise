## ADDED Requirements

### Requirement: Money value object is immutable
The `Money` type SHALL have unexported fields (`amount`, `currency`) enforcing immutability. All methods SHALL use value receivers.

#### Scenario: Fields are not directly accessible
- **WHEN** code outside the `model` package attempts to access `Money.amount` or `Money.currency`
- **THEN** it SHALL fail to compile
- **AND** `Amount()` and `Currency()` accessor methods SHALL be used instead

### Requirement: Money creation with defaults
`NewMoney` SHALL create a Money value. If currency is empty, it SHALL default to `"PLN"`.

#### Scenario: Default currency
- **WHEN** `NewMoney(decimal.NewFromInt(100), "")` is called
- **THEN** `Currency()` SHALL return `"PLN"`

#### Scenario: Explicit currency
- **WHEN** `NewMoney(decimal.NewFromInt(100), "EUR")` is called
- **THEN** `Currency()` SHALL return `"EUR"`

### Requirement: Money creation from string
`NewMoneyFromString` SHALL parse a string amount and return a Money or an error.

#### Scenario: Valid string amount
- **WHEN** `NewMoneyFromString("10.50", "PLN")` is called
- **THEN** it SHALL return a Money with amount `10.50` and no error

#### Scenario: Invalid string amount
- **WHEN** `NewMoneyFromString("not-a-number", "PLN")` is called
- **THEN** it SHALL return an error

### Requirement: Money arithmetic preserves precision
Money arithmetic (Add, Sub) SHALL use `shopspring/decimal` for arbitrary-precision and SHALL return new Money values without mutating the originals.

#### Scenario: Addition
- **WHEN** `Money(10.50 PLN).Add(Money(3.25 PLN))` is called
- **THEN** it SHALL return `Money(13.75 PLN)`
- **AND** the original values SHALL be unchanged

#### Scenario: Subtraction
- **WHEN** `Money(10.00 PLN).Sub(Money(3.25 PLN))` is called
- **THEN** it SHALL return `Money(6.75 PLN)`

### Requirement: Money arithmetic panics on currency mismatch
Arithmetic operations SHALL panic if currencies differ. This is a programmer error, not a runtime condition.

#### Scenario: Currency mismatch on Add
- **WHEN** `Money(10.00 PLN).Add(Money(5.00 EUR))` is called
- **THEN** it SHALL panic with a message containing both currency codes

### Requirement: Money comparison methods
Money SHALL provide `Equal`, `GreaterThan`, `LessThanOrEqual`, `IsZero`, `IsNegative`, `IsPositive` methods.

#### Scenario: Equal comparison
- **WHEN** `Money(10.00 PLN).Equal(Money(10.00 PLN))` is called
- **THEN** it SHALL return `true`

#### Scenario: Different currency not equal
- **WHEN** `Money(10.00 PLN).Equal(Money(10.00 EUR))` is called
- **THEN** it SHALL return `false`

#### Scenario: Zero check
- **WHEN** `Zero("PLN").IsZero()` is called
- **THEN** it SHALL return `true`

### Requirement: Money string representation
`Money.String()` SHALL return the amount with 2 decimal places followed by the currency code.

#### Scenario: Formatted string
- **WHEN** `Money(10.50 PLN).String()` is called
- **THEN** it SHALL return `"10.50 PLN"`

### Requirement: DateRange value object with validation
`DateRange` SHALL enforce that `from` is strictly before `to`. Fields SHALL be unexported for immutability.

#### Scenario: Valid date range
- **WHEN** `NewDateRange(2026-02-01, 2026-03-01)` is called
- **THEN** it SHALL return a valid `DateRange` and no error

#### Scenario: Invalid date range (from >= to)
- **WHEN** `NewDateRange(2026-03-01, 2026-02-01)` is called
- **THEN** it SHALL return an error

#### Scenario: Same date (zero-length range)
- **WHEN** `NewDateRange(2026-02-01, 2026-02-01)` is called
- **THEN** it SHALL return an error

### Requirement: DateRange convenience constructors
DateRange SHALL provide convenience constructors for common time periods.

#### Scenario: MonthRange
- **WHEN** `MonthRange(2026, 2)` is called
- **THEN** it SHALL return a DateRange from `2026-02-01` to `2026-03-01`

#### Scenario: WeekRange
- **WHEN** `WeekRange(2026, 7)` is called
- **THEN** it SHALL return a DateRange spanning ISO week 7 of 2026 (Monday to next Monday)

#### Scenario: DayRange
- **WHEN** `DayRange(2026-02-12)` is called
- **THEN** it SHALL return a DateRange from `2026-02-12` to `2026-02-13`

### Requirement: DateRange accessor methods
DateRange SHALL provide `From()` and `To()` methods returning `time.Time`.

#### Scenario: Accessors return correct values
- **WHEN** a DateRange is created from `2026-02-01` to `2026-03-01`
- **THEN** `From()` SHALL return `2026-02-01` and `To()` SHALL return `2026-03-01`

### Requirement: Sentinel domain errors
The `domain` package SHALL define sentinel errors for all domain areas using `errors.New`. These SHALL be compared using `errors.Is()`.

#### Scenario: Error sentinel values exist
- **WHEN** the `model` package is imported
- **THEN** the following sentinel errors SHALL be defined:
  - `ErrInvalidCredentials`, `ErrEmailAlreadyExists`, `ErrUserNotFound`
  - `ErrWorkspaceNotFound`, `ErrWorkspaceNameRequired`, `ErrWorkspaceDescriptionRequired`, `ErrInvalidWorkspaceType`, `ErrNotWorkspaceMember`, `ErrInsufficientPermission`
  - `ErrTransactionNotFound`, `ErrTransactionNoEntries`, `ErrEntrySumMismatch`, `ErrDuplicateFingerprint`
  - `ErrCategoryNotFound`

#### Scenario: Errors work with errors.Is
- **WHEN** `fmt.Errorf("context: %w", model.ErrUserNotFound)` wraps a sentinel error
- **THEN** `errors.Is(err, model.ErrUserNotFound)` SHALL return `true`
