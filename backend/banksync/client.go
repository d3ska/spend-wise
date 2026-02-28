package banksync

import (
	"context"
	"time"
)

// Client defines the interface for interacting with the Enable Banking API.
type Client interface {
	GetASPSPs(ctx context.Context, country string) ([]ASPSP, error)
	StartAuth(ctx context.Context, aspspName, country, redirectURL string, validUntil time.Time) (AuthResult, error)
	CreateSession(ctx context.Context, code string) (Session, error)
	GetAccountTransactions(ctx context.Context, accountUID string, dateFrom, dateTo time.Time) ([]Transaction, error)
}

// ASPSP represents a banking institution available via Enable Banking.
type ASPSP struct {
	Name               string `json:"name"`
	Country            string `json:"country"`
	Logo               string `json:"logo"`
	BIC                string `json:"bic"`
	MaxConsentValidity int    `json:"maximum_consent_validity"`
}

// AuthResult is the result of starting an authorization flow.
type AuthResult struct {
	URL             string `json:"url"`
	AuthorizationID string `json:"authorization_id"`
}

// Session represents an authorized Enable Banking session.
type Session struct {
	SessionID string           `json:"session_id"`
	Accounts  []SessionAccount `json:"accounts"`
}

// GenericIdentification represents an alternative account identifier.
type GenericIdentification struct {
	Identification string `json:"identification"`
	SchemeName     string `json:"scheme_name"`
}

// AccountIdentification represents a bank account identifier.
type AccountIdentification struct {
	IBAN  string                 `json:"iban"`
	Other *GenericIdentification `json:"other"`
}

// SessionAccount represents a bank account discovered during session creation.
type SessionAccount struct {
	UID       string                `json:"uid"`
	AccountID AccountIdentification `json:"account_id"`
	Name      string                `json:"name"`
	Currency  string                `json:"currency"`
}

// PartyIdentification represents a creditor or debtor.
type PartyIdentification struct {
	Name string `json:"name"`
}

// Transaction represents a single bank transaction from Enable Banking.
type Transaction struct {
	TransactionID         string                 `json:"transaction_id"`
	BookingDate           string                 `json:"booking_date"`
	TransactionAmount     Amount                 `json:"transaction_amount"`
	Creditor              PartyIdentification    `json:"creditor"`
	Debtor                PartyIdentification    `json:"debtor"`
	CreditorAccount       *AccountIdentification `json:"creditor_account"`
	DebtorAccount         *AccountIdentification `json:"debtor_account"`
	RemittanceInformation []string               `json:"remittance_information"`
	CreditDebitIndicator  string                 `json:"credit_debit_indicator"`
	Status                string                 `json:"status"`
}

// Amount represents a monetary amount with currency from Enable Banking.
type Amount struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}
