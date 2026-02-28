package service

import (
	"context"
	"fmt"
	"log/slog"

	"backend/model"
)

// LinkingTransactionStoreIface defines the transaction store methods used by BankAccountLinkingService.
type LinkingTransactionStoreIface interface {
	ReclassifyTransfersByCounterpartyIBAN(ctx context.Context, wsID model.WorkspaceID, ibans []string) (int64, error)
}

// BankAccountLinkingService handles linking bank accounts to workspaces.
type BankAccountLinkingService struct {
	acctStore BankAccountStoreIface
	connStore BankConnectionStoreIface
	wsStore   WorkspaceStoreIface
	txStore   LinkingTransactionStoreIface
}

// NewBankAccountLinkingService creates a new BankAccountLinkingService.
func NewBankAccountLinkingService(
	acctStore BankAccountStoreIface,
	connStore BankConnectionStoreIface,
	wsStore WorkspaceStoreIface,
	txStore LinkingTransactionStoreIface,
) *BankAccountLinkingService {
	return &BankAccountLinkingService{
		acctStore: acctStore,
		connStore: connStore,
		wsStore:   wsStore,
		txStore:   txStore,
	}
}

// LinkToWorkspace links a bank account to a workspace. Requires owner role.
func (s *BankAccountLinkingService) LinkToWorkspace(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID, bankAccountID model.BankAccountID) error {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return err
	}
	if member.Role != model.RoleOwner {
		return model.ErrInsufficientPermission
	}

	// Verify the caller owns the bank account via its parent connection.
	acct, err := s.acctStore.GetByID(ctx, bankAccountID)
	if err != nil {
		return err
	}
	conn, err := s.connStore.GetByID(ctx, acct.BankConnectionID)
	if err != nil {
		return err
	}
	if conn.UserID != callerID {
		return model.ErrBankConnectionNotFound
	}

	if err := s.acctStore.LinkToWorkspace(ctx, wsID, bankAccountID); err != nil {
		return err
	}

	// Reclassify existing transactions whose counterparty_iban matches this
	// newly-linked account's IBAN. This handles the case where bank A was
	// synced before bank B was linked — transfers between them would have been
	// classified as income/expense because the counterparty IBAN wasn't known yet.
	if acct.IBAN != nil && *acct.IBAN != "" {
		reclassified, err := s.txStore.ReclassifyTransfersByCounterpartyIBAN(ctx, wsID, []string{*acct.IBAN})
		if err != nil {
			slog.Error("failed to reclassify transfers after bank link",
				"workspace_id", wsID,
				"bank_account_id", bankAccountID,
				"iban", *acct.IBAN,
				"error", err,
			)
		} else if reclassified > 0 {
			slog.Info("reclassified transfers after bank link",
				"workspace_id", wsID,
				"bank_account_id", bankAccountID,
				"count", reclassified,
			)
		}
	}

	return nil
}

// UnlinkFromWorkspace removes a bank account from a workspace. Requires owner role.
func (s *BankAccountLinkingService) UnlinkFromWorkspace(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID, bankAccountID model.BankAccountID) error {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return err
	}
	if member.Role != model.RoleOwner {
		return model.ErrInsufficientPermission
	}

	return s.acctStore.UnlinkFromWorkspace(ctx, wsID, bankAccountID)
}

// UpdateCustomName sets a user-defined alias for a bank account.
// Verifies the caller owns the bank account via its parent connection.
func (s *BankAccountLinkingService) UpdateCustomName(ctx context.Context, callerID model.UserID, bankAccountID model.BankAccountID, customName string) (model.BankAccount, error) {
	acct, err := s.acctStore.GetByID(ctx, bankAccountID)
	if err != nil {
		return model.BankAccount{}, err
	}
	conn, err := s.connStore.GetByID(ctx, acct.BankConnectionID)
	if err != nil {
		return model.BankAccount{}, err
	}
	if conn.UserID != callerID {
		return model.BankAccount{}, model.ErrBankConnectionNotFound
	}

	return s.acctStore.UpdateCustomName(ctx, bankAccountID, customName)
}

// LinkedBankAccount is a bank account with its institution name for display.
type LinkedBankAccount struct {
	model.BankAccount
	InstitutionName string
}

// ListByUser returns all bank accounts for the user across all connections.
func (s *BankAccountLinkingService) ListByUser(ctx context.Context, callerID model.UserID) ([]LinkedBankAccount, error) {
	rows, err := s.acctStore.ListByUser(ctx, callerID)
	if err != nil {
		return nil, fmt.Errorf("listing bank accounts by user: %w", err)
	}
	result := make([]LinkedBankAccount, 0, len(rows))
	for _, row := range rows {
		result = append(result, LinkedBankAccount{
			BankAccount:     row.BankAccount,
			InstitutionName: row.InstitutionName,
		})
	}
	return result, nil
}

// ListByWorkspace returns bank accounts linked to a workspace. Requires viewer role.
func (s *BankAccountLinkingService) ListByWorkspace(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID) ([]LinkedBankAccount, error) {
	_, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return nil, err
	}

	rows, err := s.acctStore.ListByWorkspace(ctx, wsID)
	if err != nil {
		return nil, fmt.Errorf("listing linked bank accounts: %w", err)
	}

	result := make([]LinkedBankAccount, 0, len(rows))
	for _, row := range rows {
		result = append(result, LinkedBankAccount{
			BankAccount:     row.BankAccount,
			InstitutionName: row.InstitutionName,
		})
	}
	return result, nil
}
