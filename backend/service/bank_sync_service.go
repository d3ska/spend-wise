package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"backend/banksync"
	"backend/model"

	"github.com/shopspring/decimal"
)

// BankSyncTransactionStoreIface defines the transaction store methods used by BankSyncService.
type BankSyncTransactionStoreIface interface {
	Create(ctx context.Context, t model.Transaction) (model.Transaction, error)
	ExistsByFingerprint(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error)
	RelinkOrphanedTransactions(ctx context.Context, bankAccountID model.BankAccountID, iban, currency string, wsID model.WorkspaceID) (int64, error)
}

// BankSyncService handles background transaction syncing from Enable Banking.
type BankSyncService struct {
	connStore BankConnectionStoreIface
	acctStore BankAccountStoreIface
	txStore   BankSyncTransactionStoreIface
	catStore  CategoryStoreIface
	ruleStore RuleStoreIface
	client    banksync.Client
}

// NewBankSyncService creates a new BankSyncService.
func NewBankSyncService(
	connStore BankConnectionStoreIface,
	acctStore BankAccountStoreIface,
	txStore BankSyncTransactionStoreIface,
	catStore CategoryStoreIface,
	ruleStore RuleStoreIface,
	client banksync.Client,
) *BankSyncService {
	return &BankSyncService{
		connStore: connStore,
		acctStore: acctStore,
		txStore:   txStore,
		catStore:  catStore,
		ruleStore: ruleStore,
		client:    client,
	}
}

// SyncAll syncs transactions for all active bank connections.
func (s *BankSyncService) SyncAll(ctx context.Context) {
	connections, err := s.connStore.ListActive(ctx)
	if err != nil {
		slog.Error("failed to list active connections", "error", err)
		return
	}

	for _, conn := range connections {
		s.SyncConnection(ctx, conn)
	}
}

// SyncConnection syncs transactions for a single bank connection.
func (s *BankSyncService) SyncConnection(ctx context.Context, conn model.BankConnection) {
	if time.Now().After(conn.AuthExpiresAt) {
		if err := s.connStore.UpdateStatus(ctx, conn.ID, model.BankConnectionExpired); err != nil {
			slog.Error("failed to mark connection expired", "connection_id", conn.ID, "error", err)
		}
		return
	}

	ownIBANs, err := s.buildUserIBANSet(ctx, conn.UserID)
	if err != nil {
		slog.Error("failed to build IBAN set", "user_id", conn.UserID, "error", err)
		return
	}

	accounts, err := s.acctStore.ListByConnection(ctx, conn.ID)
	if err != nil {
		slog.Error("failed to list accounts for connection", "connection_id", conn.ID, "error", err)
		return
	}

	for _, acct := range accounts {
		if err := s.SyncBankAccount(ctx, conn, acct, ownIBANs); err != nil {
			slog.Error("failed to sync bank account",
				"account_id", acct.ID,
				"external_id", acct.ExternalID,
				"error", err,
			)
			continue
		}
	}
}

// buildUserIBANSet loads all bank accounts for a user and returns a set of their IBANs.
func (s *BankSyncService) buildUserIBANSet(ctx context.Context, userID model.UserID) (map[string]bool, error) {
	conns, err := s.connStore.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing user connections: %w", err)
	}
	ibanSet := make(map[string]bool)
	for _, c := range conns {
		accts, err := s.acctStore.ListByConnection(ctx, c.ID)
		if err != nil {
			return nil, fmt.Errorf("listing accounts for connection %d: %w", c.ID, err)
		}
		for _, a := range accts {
			if a.IBAN != nil && *a.IBAN != "" {
				ibanSet[*a.IBAN] = true
			}
		}
	}
	return ibanSet, nil
}

// SyncBankAccount fetches transactions from Enable Banking for a single bank account
// and inserts them into all linked workspaces.
func (s *BankSyncService) SyncBankAccount(ctx context.Context, conn model.BankConnection, acct model.BankAccount, ownIBANs map[string]bool) error {
	dateFrom := s.calcDateFrom(acct.LastSyncedAt)
	dateTo := time.Now()

	transactions, err := s.client.GetAccountTransactions(ctx, acct.ExternalID, dateFrom, dateTo)
	if err != nil {
		return fmt.Errorf("fetching transactions: %w", err)
	}

	workspaceIDs, err := s.acctStore.ListWorkspacesByBankAccount(ctx, acct.ID)
	if err != nil {
		return fmt.Errorf("listing workspaces: %w", err)
	}

	if len(workspaceIDs) == 0 {
		return nil
	}

	for _, wsID := range workspaceIDs {
		// Re-link orphaned transactions from previous deletions.
		if acct.IBAN != nil && *acct.IBAN != "" {
			relinked, err := s.txStore.RelinkOrphanedTransactions(ctx, acct.ID, *acct.IBAN, acct.Currency, wsID)
			if err != nil {
				slog.Error("failed to relink orphaned transactions",
					"bank_account_id", acct.ID,
					"iban", *acct.IBAN,
					"workspace_id", wsID,
					"error", err,
				)
			} else if relinked > 0 {
				slog.Info("relinked orphaned transactions",
					"bank_account_id", acct.ID,
					"workspace_id", wsID,
					"count", relinked,
				)
			}
		}

		catID, err := s.ensureUncategorizedCategory(ctx, wsID)
		if err != nil {
			return fmt.Errorf("ensuring uncategorized category for workspace %d: %w", wsID, err)
		}

		for _, ebTx := range transactions {
			tx, err := s.mapTransaction(ebTx, wsID, conn.UserID, acct.ID, acct.IBAN, catID, ownIBANs)
			if err != nil {
				slog.Warn("skipping unmappable transaction",
					"transaction_id", ebTx.TransactionID,
					"error", err,
				)
				continue
			}

			exists, err := s.txStore.ExistsByFingerprint(ctx, wsID, *tx.Fingerprint)
			if err != nil {
				return fmt.Errorf("checking fingerprint: %w", err)
			}
			if exists {
				continue
			}

			// Apply categorization rules to resolve a category from the description.
			bankAcctID := acct.ID
			if s.ruleStore != nil {
				resolvedCatID, err := s.ruleStore.ResolveForTransaction(ctx, wsID, conn.UserID, tx.Description, tx.TotalAmount.Amount(), &bankAcctID, tx.CounterpartyIBAN)
				if err != nil {
					slog.Warn("rule resolution failed, using Uncategorized",
						"description", tx.Description,
						"error", err,
					)
				} else if resolvedCatID != nil {
					tx.Entries[0].CategoryID = *resolvedCatID
				}
			}

			if _, err := s.txStore.Create(ctx, tx); err != nil {
				return fmt.Errorf("creating transaction: %w", err)
			}
		}
	}

	return s.acctStore.UpdateLastSyncedAt(ctx, acct.ID, dateTo)
}

// calcDateFrom determines the start date for fetching transactions.
func (s *BankSyncService) calcDateFrom(lastSynced *time.Time) time.Time {
	if lastSynced == nil {
		return time.Now().AddDate(0, -3, 0)
	}
	return lastSynced.AddDate(0, 0, -3)
}

// ensureUncategorizedCategory returns the "Uncategorized" category ID for a workspace,
// creating it if it doesn't exist.
func (s *BankSyncService) ensureUncategorizedCategory(ctx context.Context, wsID model.WorkspaceID) (model.CategoryID, error) {
	cat, err := s.catStore.GetBySlug(ctx, wsID, "uncategorized")
	if err == nil {
		return cat.ID, nil
	}
	if !errors.Is(err, model.ErrCategoryNotFound) {
		return 0, fmt.Errorf("checking uncategorized category: %w", err)
	}

	slug := "uncategorized"
	cat, err = s.catStore.Create(ctx, model.Category{
		WorkspaceID: wsID,
		Name:        "Uncategorized",
		Icon:        "📂",
		Slug:        &slug,
	})
	if err != nil {
		return 0, fmt.Errorf("creating uncategorized category: %w", err)
	}
	return cat.ID, nil
}

// extractCounterpartyIBAN returns the IBAN of the counterparty based on transaction direction.
// For outgoing (expense): the counterparty is the creditor.
// For incoming (income): the counterparty is the debtor.
func (s *BankSyncService) extractCounterpartyIBAN(ebTx banksync.Transaction, txType model.TransactionType) string {
	switch txType {
	case model.TypeExpense:
		return ibanFromAccount(ebTx.CreditorAccount)
	case model.TypeIncome:
		return ibanFromAccount(ebTx.DebtorAccount)
	}
	return ""
}

// ibanFromAccount extracts an IBAN from an AccountIdentification.
// The Enable Banking API returns {"iban": "...", "other": {...}} for account fields.
func ibanFromAccount(acct *banksync.AccountIdentification) string {
	if acct == nil {
		return ""
	}
	if acct.IBAN != "" {
		return acct.IBAN
	}
	// Fallback: check "other" field for IBAN scheme.
	if acct.Other != nil && acct.Other.Identification != "" {
		if acct.Other.SchemeName == "IBAN" || acct.Other.SchemeName == "" {
			return acct.Other.Identification
		}
	}
	return ""
}

// buildFingerprint creates a composite fingerprint for dedup.
// When transaction_id is present: sha256(transaction_id|amount|currency)
//
//	— booking_date is excluded because banks may shift dates between syncs.
//
// When absent: sha256(booking_date|amount|currency|first_remittance_info)
func buildFingerprint(tx banksync.Transaction) string {
	var input string
	if tx.TransactionID != "" {
		input = fmt.Sprintf("%s|%s|%s",
			tx.TransactionID,
			tx.TransactionAmount.Amount,
			tx.TransactionAmount.Currency,
		)
	} else {
		remittance := ""
		if len(tx.RemittanceInformation) > 0 {
			remittance = tx.RemittanceInformation[0]
		}
		input = fmt.Sprintf("%s|%s|%s|%s",
			tx.BookingDate,
			tx.TransactionAmount.Amount,
			tx.TransactionAmount.Currency,
			remittance,
		)
	}
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

// mapTransaction converts an Enable Banking transaction to a SpendWise transaction.
func (s *BankSyncService) mapTransaction(
	ebTx banksync.Transaction,
	wsID model.WorkspaceID,
	userID model.UserID,
	bankAccountID model.BankAccountID,
	bankIBAN *string,
	categoryID model.CategoryID,
	ownIBANs map[string]bool,
) (model.Transaction, error) {
	amount, err := decimal.NewFromString(ebTx.TransactionAmount.Amount)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("parsing amount %q: %w", ebTx.TransactionAmount.Amount, err)
	}

	date, err := time.Parse("2006-01-02", ebTx.BookingDate)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("parsing date %q: %w", ebTx.BookingDate, err)
	}

	description := ebTx.Creditor.Name
	if description == "" {
		description = ebTx.Debtor.Name
	}
	if description == "" && len(ebTx.RemittanceInformation) > 0 {
		description = ebTx.RemittanceInformation[0]
	}
	if description == "" {
		description = "Bank transaction"
	}

	currency := ebTx.TransactionAmount.Currency
	if currency == "" {
		currency = "PLN"
	}

	txType := model.TypeExpense
	if ebTx.CreditDebitIndicator == "CRDT" {
		txType = model.TypeIncome
	} else if ebTx.CreditDebitIndicator == "" && amount.IsPositive() {
		txType = model.TypeIncome
	}

	// Extract and store counterparty IBAN for future reclassification.
	var counterpartyIBANPtr *string
	if counterpartyIBAN := s.extractCounterpartyIBAN(ebTx, txType); counterpartyIBAN != "" {
		counterpartyIBANPtr = &counterpartyIBAN
		if ownIBANs[counterpartyIBAN] {
			txType = model.TypeTransfer
		}
	}

	money := model.NewMoney(amount, currency)
	fingerprint := buildFingerprint(ebTx)

	return model.Transaction{
		WorkspaceID:      wsID,
		CreatedBy:        userID,
		TotalAmount:      money,
		Description:      description,
		Date:             date,
		Fingerprint:      &fingerprint,
		Source:           model.SourceBank,
		Type:             txType,
		BankAccountID:    &bankAccountID,
		IBAN:             bankIBAN,
		CounterpartyIBAN: counterpartyIBANPtr,
		Entries: []model.Entry{
			{
				CategoryID: categoryID,
				Amount:     money,
			},
		},
	}, nil
}
