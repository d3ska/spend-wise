package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"backend/banksync"
	"backend/model"
	"backend/store"

	"github.com/jackc/pgx/v5/pgtype"
)

// redactIBAN masks an IBAN, showing only the country code and last 4 characters.
func redactIBAN(iban string) string {
	if len(iban) <= 6 {
		return "****"
	}
	return iban[:2] + "****" + iban[len(iban)-4:]
}

// BankConnectionStoreIface defines the bank connection store methods used by BankConnectionService.
type BankConnectionStoreIface interface {
	Create(ctx context.Context, bc model.BankConnection) (model.BankConnection, error)
	GetByID(ctx context.Context, id model.BankConnectionID) (model.BankConnection, error)
	ListByUser(ctx context.Context, userID model.UserID) ([]model.BankConnection, error)
	ListActive(ctx context.Context) ([]model.BankConnection, error)
	UpdateStatus(ctx context.Context, id model.BankConnectionID, status model.BankConnectionStatus) error
	UpdateSession(ctx context.Context, id model.BankConnectionID, sessionID string, status model.BankConnectionStatus, expiresAt pgtype.Timestamptz) error
	Delete(ctx context.Context, id model.BankConnectionID, userID model.UserID) error
}

// BankAccountStoreIface defines the bank account store methods used by bank services.
type BankAccountStoreIface interface {
	Create(ctx context.Context, ba model.BankAccount) (model.BankAccount, error)
	GetByID(ctx context.Context, id model.BankAccountID) (model.BankAccount, error)
	ListByConnection(ctx context.Context, connID model.BankConnectionID) ([]model.BankAccount, error)
	ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]store.BankAccountWithInstitution, error)
	ListByUser(ctx context.Context, userID model.UserID) ([]store.BankAccountWithInstitution, error)
	LinkToWorkspace(ctx context.Context, wsID model.WorkspaceID, baID model.BankAccountID) error
	UnlinkFromWorkspace(ctx context.Context, wsID model.WorkspaceID, baID model.BankAccountID) error
	ListWorkspacesByBankAccount(ctx context.Context, baID model.BankAccountID) ([]model.WorkspaceID, error)
	UpdateLastSyncedAt(ctx context.Context, id model.BankAccountID, t time.Time) error
	GetByIBANAndCurrencyForUser(ctx context.Context, iban, currency string, userID model.UserID) (model.BankAccount, error)
	UpdateConnection(ctx context.Context, id model.BankAccountID, externalID string, connID model.BankConnectionID) error
	UpdateCustomName(ctx context.Context, id model.BankAccountID, customName string) (model.BankAccount, error)
}

// BankConnectionService handles bank connection operations.
type BankConnectionService struct {
	connStore BankConnectionStoreIface
	acctStore BankAccountStoreIface
	client    banksync.Client
}

// NewBankConnectionService creates a new BankConnectionService.
func NewBankConnectionService(
	connStore BankConnectionStoreIface,
	acctStore BankAccountStoreIface,
	client banksync.Client,
) *BankConnectionService {
	return &BankConnectionService{
		connStore: connStore,
		acctStore: acctStore,
		client:    client,
	}
}

// InitiateResult is the result of initiating a bank connection.
type InitiateResult struct {
	Connection model.BankConnection
	AuthLink   string
}

// ListInstitutions returns available banking institutions (ASPSPs) for a country.
func (s *BankConnectionService) ListInstitutions(ctx context.Context, countryCode string) ([]banksync.ASPSP, error) {
	aspsps, err := s.client.GetASPSPs(ctx, countryCode)
	if err != nil {
		return nil, fmt.Errorf("listing ASPSPs: %w", err)
	}
	return aspsps, nil
}

// Initiate starts an Enable Banking auth flow, persists the connection,
// and returns the auth link for user redirect.
func (s *BankConnectionService) Initiate(ctx context.Context, userID model.UserID, institutionID, institutionName, redirectURL, country string) (InitiateResult, error) {
	// Query ASPSP metadata for max consent validity.
	aspsps, err := s.client.GetASPSPs(ctx, country)
	var maxConsentDays int
	if err == nil {
		for _, a := range aspsps {
			if a.Name == institutionID {
				maxConsentDays = a.MaxConsentValidity
				break
			}
		}
	}

	// Compute validUntil = min(maxConsent, 180 days).
	validUntil := time.Now().AddDate(0, 0, 180)
	if maxConsentDays > 0 && maxConsentDays < 180 {
		validUntil = time.Now().AddDate(0, 0, maxConsentDays)
	}

	authResult, err := s.client.StartAuth(ctx, institutionID, country, redirectURL, validUntil)
	if err != nil {
		return InitiateResult{}, fmt.Errorf("starting auth: %w", err)
	}

	conn, err := s.connStore.Create(ctx, model.BankConnection{
		UserID:          userID,
		InstitutionID:   institutionID,
		InstitutionName: institutionName,
		SessionID:       authResult.AuthorizationID,
		Status:          model.BankConnectionActive,
		AuthExpiresAt:   validUntil,
	})
	if err != nil {
		return InitiateResult{}, fmt.Errorf("persisting connection: %w", err)
	}

	return InitiateResult{Connection: conn, AuthLink: authResult.URL}, nil
}

// Complete finalizes a bank connection by exchanging the authorization code for a session.
func (s *BankConnectionService) Complete(ctx context.Context, userID model.UserID, connectionID model.BankConnectionID, authorizationCode string) ([]model.BankAccount, error) {
	slog.Info("BankComplete: looking up connection", "connection_id", connectionID, "user_id", userID)

	conn, err := s.connStore.GetByID(ctx, connectionID)
	if err != nil {
		slog.Error("BankComplete: failed to get connection", "error", err)
		return nil, err
	}
	if conn.UserID != userID {
		slog.Warn("BankComplete: user mismatch", "conn_user", conn.UserID, "caller", userID)
		return nil, model.ErrBankConnectionNotFound
	}

	slog.Info("BankComplete: calling CreateSession", "code_len", len(authorizationCode))
	session, err := s.client.CreateSession(ctx, authorizationCode)
	if err != nil {
		slog.Error("BankComplete: CreateSession failed", "error", err)
		return nil, fmt.Errorf("creating session: %w", err)
	}
	slog.Info("BankComplete: session created", "accounts_count", len(session.Accounts))

	// Update connection's session_id to the new session ID.
	if err := s.connStore.UpdateSession(ctx, connectionID, session.SessionID, model.BankConnectionActive, pgtype.Timestamptz{Time: conn.AuthExpiresAt, Valid: true}); err != nil {
		slog.Error("BankComplete: failed to update session", "error", err)
		return nil, fmt.Errorf("updating connection session: %w", err)
	}

	accounts := make([]model.BankAccount, 0, len(session.Accounts))
	for _, sa := range session.Accounts {
		slog.Info("BankComplete: processing bank account", "name", sa.Name, "iban", redactIBAN(sa.AccountID.IBAN), "currency", sa.Currency)
		var iban *string
		if sa.AccountID.IBAN != "" {
			iban = &sa.AccountID.IBAN
		}
		name := sa.Name
		if name == "" {
			name = sa.UID
		}

		// Try to reuse existing bank account by IBAN + currency + user.
		if iban != nil {
			existing, err := s.acctStore.GetByIBANAndCurrencyForUser(ctx, *iban, sa.Currency, userID)
			if err == nil {
				slog.Info("BankComplete: reusing existing bank account", "existing_id", existing.ID)
				if err := s.acctStore.UpdateConnection(ctx, existing.ID, sa.UID, connectionID); err != nil {
					return nil, fmt.Errorf("updating bank account %s: %w", sa.UID, err)
				}
				existing.ExternalID = sa.UID
				existing.BankConnectionID = connectionID
				existing.Name = name
				accounts = append(accounts, existing)
				continue
			}
			if !errors.Is(err, model.ErrBankAccountNotFound) {
				return nil, fmt.Errorf("looking up bank account by IBAN: %w", err)
			}
		}

		acct, err := s.acctStore.Create(ctx, model.BankAccount{
			BankConnectionID: connectionID,
			ExternalID:       sa.UID,
			IBAN:             iban,
			Currency:         sa.Currency,
			Name:             name,
		})
		if err != nil {
			slog.Error("BankComplete: failed to create bank account", "error", err)
			return nil, fmt.Errorf("creating bank account %s: %w", sa.UID, err)
		}
		accounts = append(accounts, acct)
	}

	slog.Info("BankComplete: done", "total_accounts", len(accounts))
	return accounts, nil
}

// GetConnection returns a bank connection by ID after verifying ownership.
func (s *BankConnectionService) GetConnection(ctx context.Context, userID model.UserID, connectionID model.BankConnectionID) (model.BankConnection, error) {
	conn, err := s.connStore.GetByID(ctx, connectionID)
	if err != nil {
		return model.BankConnection{}, err
	}
	if conn.UserID != userID {
		return model.BankConnection{}, model.ErrBankConnectionNotFound
	}
	return conn, nil
}

// CompleteByCode finalizes a bank connection using just the authorization code.
// It finds the user's most recent connection and completes it.
// Returns the discovered accounts and the institution name.
func (s *BankConnectionService) CompleteByCode(ctx context.Context, userID model.UserID, authorizationCode string) ([]model.BankAccount, string, error) {
	slog.Info("CompleteByCode: listing connections", "user_id", userID)

	connections, err := s.connStore.ListByUser(ctx, userID)
	if err != nil {
		slog.Error("CompleteByCode: failed to list connections", "error", err)
		return nil, "", fmt.Errorf("listing connections: %w", err)
	}

	slog.Info("CompleteByCode: found connections", "count", len(connections))

	if len(connections) == 0 {
		slog.Warn("CompleteByCode: no connections found for user")
		return nil, "", model.ErrBankConnectionNotFound
	}

	conn := connections[0]
	slog.Info("CompleteByCode: using most recent connection", "connection_id", conn.ID, "institution", conn.InstitutionName, "status", conn.Status)
	accounts, err := s.Complete(ctx, userID, conn.ID, authorizationCode)
	if err != nil {
		return nil, "", err
	}
	return accounts, conn.InstitutionName, nil
}

// ListByUser returns all bank connections for a user.
func (s *BankConnectionService) ListByUser(ctx context.Context, userID model.UserID) ([]model.BankConnection, error) {
	return s.connStore.ListByUser(ctx, userID)
}

// Delete removes a bank connection owned by the user.
func (s *BankConnectionService) Delete(ctx context.Context, userID model.UserID, connectionID model.BankConnectionID) error {
	return s.connStore.Delete(ctx, connectionID, userID)
}

// Deactivate sets a bank connection to inactive status, skipping it in future syncs.
func (s *BankConnectionService) Deactivate(ctx context.Context, userID model.UserID, connectionID model.BankConnectionID) error {
	conn, err := s.connStore.GetByID(ctx, connectionID)
	if err != nil {
		return err
	}
	if conn.UserID != userID {
		return model.ErrBankConnectionNotFound
	}
	return s.connStore.UpdateStatus(ctx, connectionID, model.BankConnectionInactive)
}

// Reactivate sets an inactive bank connection back to active status, resuming syncs.
func (s *BankConnectionService) Reactivate(ctx context.Context, userID model.UserID, connectionID model.BankConnectionID) error {
	conn, err := s.connStore.GetByID(ctx, connectionID)
	if err != nil {
		return err
	}
	if conn.UserID != userID {
		return model.ErrBankConnectionNotFound
	}
	if conn.Status != model.BankConnectionInactive {
		return fmt.Errorf("connection is not inactive")
	}
	return s.connStore.UpdateStatus(ctx, connectionID, model.BankConnectionActive)
}

// Reconnect creates a new auth session for an expired connection.
func (s *BankConnectionService) Reconnect(ctx context.Context, userID model.UserID, connectionID model.BankConnectionID, redirectURL string) (string, error) {
	conn, err := s.connStore.GetByID(ctx, connectionID)
	if err != nil {
		return "", err
	}
	if conn.UserID != userID {
		return "", model.ErrBankConnectionNotFound
	}

	// Query ASPSP metadata for country and consent duration.
	aspsps, err := s.client.GetASPSPs(ctx, "")
	var maxConsentDays int
	var country string
	if err == nil {
		for _, a := range aspsps {
			if a.Name == conn.InstitutionID {
				maxConsentDays = a.MaxConsentValidity
				country = a.Country
				break
			}
		}
	}

	validUntil := time.Now().AddDate(0, 0, 180)
	if maxConsentDays > 0 && maxConsentDays < 180 {
		validUntil = time.Now().AddDate(0, 0, maxConsentDays)
	}

	authResult, err := s.client.StartAuth(ctx, conn.InstitutionID, country, redirectURL, validUntil)
	if err != nil {
		return "", fmt.Errorf("starting auth: %w", err)
	}

	if err := s.connStore.UpdateSession(ctx, connectionID, authResult.AuthorizationID, model.BankConnectionActive, pgtype.Timestamptz{Time: validUntil, Valid: true}); err != nil {
		return "", fmt.Errorf("updating connection: %w", err)
	}

	return authResult.URL, nil
}
