package store

import (
	"context"
	"errors"
	"fmt"

	"backend/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BankConnectionStore provides bank connection persistence operations.
type BankConnectionStore struct {
	q *Queries
}

// NewBankConnectionStore creates a new BankConnectionStore.
func NewBankConnectionStore(pool *pgxpool.Pool) *BankConnectionStore {
	return &BankConnectionStore{q: New(pool)}
}

// Create inserts a new bank connection and returns it with a generated ID.
func (s *BankConnectionStore) Create(ctx context.Context, bc model.BankConnection) (model.BankConnection, error) {
	row, err := s.q.InsertBankConnection(ctx, InsertBankConnectionParams{
		UserID:          int64(bc.UserID),
		InstitutionID:   bc.InstitutionID,
		InstitutionName: bc.InstitutionName,
		SessionID:       bc.SessionID,
		Status:          string(bc.Status),
		AuthExpiresAt:   pgtype.Timestamptz{Time: bc.AuthExpiresAt, Valid: true},
	})
	if err != nil {
		return model.BankConnection{}, fmt.Errorf("inserting bank connection: %w", err)
	}
	return toModelBankConnection(row), nil
}

// GetByID returns a bank connection by ID.
func (s *BankConnectionStore) GetByID(ctx context.Context, id model.BankConnectionID) (model.BankConnection, error) {
	row, err := s.q.GetBankConnectionByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BankConnection{}, model.ErrBankConnectionNotFound
		}
		return model.BankConnection{}, fmt.Errorf("getting bank connection: %w", err)
	}
	return toModelBankConnection(row), nil
}

// ListByUser returns all bank connections for a user.
func (s *BankConnectionStore) ListByUser(ctx context.Context, userID model.UserID) ([]model.BankConnection, error) {
	rows, err := s.q.ListBankConnectionsByUser(ctx, int64(userID))
	if err != nil {
		return nil, fmt.Errorf("listing bank connections: %w", err)
	}
	result := make([]model.BankConnection, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelBankConnection(row))
	}
	return result, nil
}

// ListActive returns all bank connections with active status and valid auth.
func (s *BankConnectionStore) ListActive(ctx context.Context) ([]model.BankConnection, error) {
	rows, err := s.q.ListActiveBankConnections(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing active bank connections: %w", err)
	}
	result := make([]model.BankConnection, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelBankConnection(row))
	}
	return result, nil
}

// UpdateStatus updates a bank connection's status.
func (s *BankConnectionStore) UpdateStatus(ctx context.Context, id model.BankConnectionID, status model.BankConnectionStatus) error {
	if err := s.q.UpdateBankConnectionStatus(ctx, UpdateBankConnectionStatusParams{
		ID:     int64(id),
		Status: string(status),
	}); err != nil {
		return fmt.Errorf("updating bank connection status: %w", err)
	}
	return nil
}

// UpdateSession updates a connection's session, status, and expiry (for reconnect).
func (s *BankConnectionStore) UpdateSession(ctx context.Context, id model.BankConnectionID, sessionID string, status model.BankConnectionStatus, expiresAt pgtype.Timestamptz) error {
	if err := s.q.UpdateBankConnectionSession(ctx, UpdateBankConnectionSessionParams{
		ID:            int64(id),
		SessionID:     sessionID,
		Status:        string(status),
		AuthExpiresAt: expiresAt,
	}); err != nil {
		return fmt.Errorf("updating bank connection session: %w", err)
	}
	return nil
}

// Delete removes a bank connection owned by the given user.
func (s *BankConnectionStore) Delete(ctx context.Context, id model.BankConnectionID, userID model.UserID) error {
	result, err := s.q.DeleteBankConnection(ctx, DeleteBankConnectionParams{
		ID:     int64(id),
		UserID: int64(userID),
	})
	if err != nil {
		return fmt.Errorf("deleting bank connection: %w", err)
	}
	if result.RowsAffected() == 0 {
		return model.ErrBankConnectionNotFound
	}
	return nil
}

func toModelBankConnection(row BankConnection) model.BankConnection {
	return model.BankConnection{
		ID:              model.BankConnectionID(row.ID),
		UserID:          model.UserID(row.UserID),
		InstitutionID:   row.InstitutionID,
		InstitutionName: row.InstitutionName,
		SessionID:       row.SessionID,
		Status:          model.BankConnectionStatus(row.Status),
		AuthExpiresAt:   row.AuthExpiresAt.Time,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}
