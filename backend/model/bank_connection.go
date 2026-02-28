package model

import "time"

// BankConnectionID is a typed wrapper for bank connection identifiers.
type BankConnectionID int64

// BankAccountID is a typed wrapper for bank account identifiers.
type BankAccountID int64

// BankConnectionStatus represents the state of a bank connection.
type BankConnectionStatus string

const (
	BankConnectionActive   BankConnectionStatus = "active"
	BankConnectionExpired  BankConnectionStatus = "expired"
	BankConnectionRevoked  BankConnectionStatus = "revoked"
	BankConnectionInactive BankConnectionStatus = "inactive"
)

// BankConnection represents a user's authenticated connection to a bank via Enable Banking.
type BankConnection struct {
	ID              BankConnectionID
	UserID          UserID
	InstitutionID   string
	InstitutionName string
	SessionID       string
	Status          BankConnectionStatus
	AuthExpiresAt   time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// BankAccount represents a single bank account discovered via a bank connection.
type BankAccount struct {
	ID               BankAccountID
	BankConnectionID BankConnectionID
	ExternalID       string
	IBAN             *string
	Currency         string
	Name             string
	CustomName       string
	LastSyncedAt     *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// DisplayName returns CustomName if non-empty, otherwise Name.
func (ba BankAccount) DisplayName() string {
	if ba.CustomName != "" {
		return ba.CustomName
	}
	return ba.Name
}
