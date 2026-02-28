package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BankAccountStore provides bank account and workspace linking persistence.
type BankAccountStore struct {
	q *Queries
}

// NewBankAccountStore creates a new BankAccountStore.
func NewBankAccountStore(pool *pgxpool.Pool) *BankAccountStore {
	return &BankAccountStore{q: New(pool)}
}

// Create inserts a new bank account.
func (s *BankAccountStore) Create(ctx context.Context, ba model.BankAccount) (model.BankAccount, error) {
	row, err := s.q.InsertBankAccount(ctx, InsertBankAccountParams{
		BankConnectionID: int64(ba.BankConnectionID),
		ExternalID:       ba.ExternalID,
		Iban:             stringPtrToText(ba.IBAN),
		Currency:         ba.Currency,
		Name:             ba.Name,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return model.BankAccount{}, model.ErrBankAccountAlreadyExists
		}
		return model.BankAccount{}, fmt.Errorf("inserting bank account: %w", err)
	}
	return toModelBankAccount(row), nil
}

// GetByID returns a bank account by ID.
func (s *BankAccountStore) GetByID(ctx context.Context, id model.BankAccountID) (model.BankAccount, error) {
	row, err := s.q.GetBankAccountByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BankAccount{}, model.ErrBankAccountNotFound
		}
		return model.BankAccount{}, fmt.Errorf("getting bank account: %w", err)
	}
	return toModelBankAccount(row), nil
}

// ListByConnection returns all bank accounts for a connection.
func (s *BankAccountStore) ListByConnection(ctx context.Context, connID model.BankConnectionID) ([]model.BankAccount, error) {
	rows, err := s.q.ListBankAccountsByConnection(ctx, int64(connID))
	if err != nil {
		return nil, fmt.Errorf("listing bank accounts: %w", err)
	}
	result := make([]model.BankAccount, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelBankAccount(row))
	}
	return result, nil
}

// BankAccountWithInstitution is a bank account with its parent connection's institution name.
type BankAccountWithInstitution struct {
	model.BankAccount
	InstitutionName string
}

// ListByWorkspace returns bank accounts linked to a workspace with institution names.
func (s *BankAccountStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]BankAccountWithInstitution, error) {
	rows, err := s.q.ListBankAccountsByWorkspace(ctx, int64(wsID))
	if err != nil {
		return nil, fmt.Errorf("listing bank accounts by workspace: %w", err)
	}
	result := make([]BankAccountWithInstitution, 0, len(rows))
	for _, row := range rows {
		result = append(result, BankAccountWithInstitution{
			BankAccount: model.BankAccount{
				ID:               model.BankAccountID(row.ID),
				BankConnectionID: model.BankConnectionID(row.BankConnectionID),
				ExternalID:       row.ExternalID,
				IBAN:             textToStringPtr(row.Iban),
				Currency:         row.Currency,
				Name:             row.Name,
				CustomName:       row.CustomName,
				LastSyncedAt:     timestamptzToTimePtr(row.LastSyncedAt),
				CreatedAt:        row.CreatedAt.Time,
				UpdatedAt:        row.UpdatedAt.Time,
			},
			InstitutionName: row.InstitutionName,
		})
	}
	return result, nil
}

// ListByUser returns all bank accounts for a user across all connections.
func (s *BankAccountStore) ListByUser(ctx context.Context, userID model.UserID) ([]BankAccountWithInstitution, error) {
	rows, err := s.q.ListBankAccountsByUser(ctx, int64(userID))
	if err != nil {
		return nil, fmt.Errorf("listing bank accounts by user: %w", err)
	}
	result := make([]BankAccountWithInstitution, 0, len(rows))
	for _, row := range rows {
		result = append(result, BankAccountWithInstitution{
			BankAccount: model.BankAccount{
				ID:               model.BankAccountID(row.ID),
				BankConnectionID: model.BankConnectionID(row.BankConnectionID),
				ExternalID:       row.ExternalID,
				IBAN:             textToStringPtr(row.Iban),
				Currency:         row.Currency,
				Name:             row.Name,
				CustomName:       row.CustomName,
				LastSyncedAt:     timestamptzToTimePtr(row.LastSyncedAt),
				CreatedAt:        row.CreatedAt.Time,
				UpdatedAt:        row.UpdatedAt.Time,
			},
			InstitutionName: row.InstitutionName,
		})
	}
	return result, nil
}

// LinkToWorkspace links a bank account to a workspace. Idempotent.
func (s *BankAccountStore) LinkToWorkspace(ctx context.Context, wsID model.WorkspaceID, baID model.BankAccountID) error {
	if err := s.q.LinkBankAccountToWorkspace(ctx, LinkBankAccountToWorkspaceParams{
		WorkspaceID:   int64(wsID),
		BankAccountID: int64(baID),
	}); err != nil {
		return fmt.Errorf("linking bank account to workspace: %w", err)
	}
	return nil
}

// UnlinkFromWorkspace removes a bank account from a workspace.
func (s *BankAccountStore) UnlinkFromWorkspace(ctx context.Context, wsID model.WorkspaceID, baID model.BankAccountID) error {
	_, err := s.q.UnlinkBankAccountFromWorkspace(ctx, UnlinkBankAccountFromWorkspaceParams{
		WorkspaceID:   int64(wsID),
		BankAccountID: int64(baID),
	})
	if err != nil {
		return fmt.Errorf("unlinking bank account from workspace: %w", err)
	}
	return nil
}

// ListWorkspacesByBankAccount returns workspace IDs linked to a bank account.
func (s *BankAccountStore) ListWorkspacesByBankAccount(ctx context.Context, baID model.BankAccountID) ([]model.WorkspaceID, error) {
	rows, err := s.q.ListWorkspacesByBankAccount(ctx, int64(baID))
	if err != nil {
		return nil, fmt.Errorf("listing workspaces for bank account: %w", err)
	}
	result := make([]model.WorkspaceID, 0, len(rows))
	for _, id := range rows {
		result = append(result, model.WorkspaceID(id))
	}
	return result, nil
}

// UpdateLastSyncedAt updates the last synced timestamp.
func (s *BankAccountStore) UpdateLastSyncedAt(ctx context.Context, id model.BankAccountID, t time.Time) error {
	if err := s.q.UpdateBankAccountLastSyncedAt(ctx, UpdateBankAccountLastSyncedAtParams{
		ID:           int64(id),
		LastSyncedAt: pgtype.Timestamptz{Time: t, Valid: true},
	}); err != nil {
		return fmt.Errorf("updating last synced at: %w", err)
	}
	return nil
}

// GetByIBANAndCurrencyForUser looks up a bank account by IBAN + currency scoped to a user.
func (s *BankAccountStore) GetByIBANAndCurrencyForUser(ctx context.Context, iban, currency string, userID model.UserID) (model.BankAccount, error) {
	row, err := s.q.GetBankAccountByIBANAndCurrencyForUser(ctx, GetBankAccountByIBANAndCurrencyForUserParams{
		Iban:     pgtype.Text{String: iban, Valid: true},
		Currency: currency,
		UserID:   int64(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BankAccount{}, model.ErrBankAccountNotFound
		}
		return model.BankAccount{}, fmt.Errorf("getting bank account by IBAN: %w", err)
	}
	return toModelBankAccount(row), nil
}

// UpdateConnection re-points a bank account to a new connection and external ID.
func (s *BankAccountStore) UpdateConnection(ctx context.Context, id model.BankAccountID, externalID string, connID model.BankConnectionID) error {
	if err := s.q.UpdateBankAccountConnection(ctx, UpdateBankAccountConnectionParams{
		ID:               int64(id),
		ExternalID:       externalID,
		BankConnectionID: int64(connID),
	}); err != nil {
		return fmt.Errorf("updating bank account connection: %w", err)
	}
	return nil
}

// UpdateCustomName sets a user-defined custom name for a bank account.
func (s *BankAccountStore) UpdateCustomName(ctx context.Context, id model.BankAccountID, customName string) (model.BankAccount, error) {
	row, err := s.q.UpdateBankAccountCustomName(ctx, UpdateBankAccountCustomNameParams{
		ID:         int64(id),
		CustomName: customName,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BankAccount{}, model.ErrBankAccountNotFound
		}
		return model.BankAccount{}, fmt.Errorf("updating bank account custom name: %w", err)
	}
	return toModelBankAccount(row), nil
}

func toModelBankAccount(row BankAccount) model.BankAccount {
	return model.BankAccount{
		ID:               model.BankAccountID(row.ID),
		BankConnectionID: model.BankConnectionID(row.BankConnectionID),
		ExternalID:       row.ExternalID,
		IBAN:             textToStringPtr(row.Iban),
		Currency:         row.Currency,
		Name:             row.Name,
		CustomName:       row.CustomName,
		LastSyncedAt:     timestamptzToTimePtr(row.LastSyncedAt),
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func timestamptzToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func isUniqueViolation(err error) bool {
	// pgx wraps constraint violations; check the error code.
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
