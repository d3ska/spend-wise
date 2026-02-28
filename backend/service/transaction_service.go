package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/model"
	"backend/store"

	"github.com/shopspring/decimal"
)

// TransactionStoreIface defines the transaction store methods used by TransactionService.
type TransactionStoreIface interface {
	Create(ctx context.Context, t model.Transaction) (model.Transaction, error)
	GetByID(ctx context.Context, id model.TransactionID) (model.Transaction, error)
	ListByWorkspace(ctx context.Context, wsID model.WorkspaceID, from, to time.Time, limit, offset int32, filter store.ListByWorkspaceFilter) ([]model.Transaction, error)
	Update(ctx context.Context, id model.TransactionID, t model.Transaction) (model.Transaction, error)
	Delete(ctx context.Context, id model.TransactionID) error
	ExistsByFingerprint(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error)
	ListAllEntries(ctx context.Context, wsID model.WorkspaceID) ([]store.TransactionEntryRow, error)
	UpdateEntryCategory(ctx context.Context, entryID model.EntryID, catID model.CategoryID) error
	GetByIDs(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) ([]model.Transaction, error)
	BulkCategorizeFirstEntry(ctx context.Context, txIDs []model.TransactionID, catID model.CategoryID) (int64, error)
	BulkDelete(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) (int64, error)
}

// CreateTransactionInput holds input for creating a transaction.
type CreateTransactionInput struct {
	Description string
	Date        time.Time
	TotalAmount model.Money
	Type        model.TransactionType
	Notes       string
	Entries     []CreateEntryInput
}

// CreateEntryInput holds input for a single entry within a transaction.
type CreateEntryInput struct {
	CategoryID    model.CategoryID
	ParticipantID *model.UserID
	Amount        model.Money
	Note          string
}

// UpdateTransactionInput holds input for updating a transaction.
type UpdateTransactionInput struct {
	Description string
	Date        time.Time
	Type        model.TransactionType
	Notes       string
	Entries     []CreateEntryInput
}

// ImportTransactionInput holds input for importing a single transaction.
type ImportTransactionInput struct {
	CreateTransactionInput
	Fingerprint string
}

// ImportResult holds the result of a batch import.
type ImportResult struct {
	Imported     int
	Skipped      int
	Transactions []model.Transaction
}

// ListTransactionsInput holds input for listing transactions.
type ListTransactionsInput struct {
	From          time.Time
	To            time.Time
	Limit         int32
	Offset        int32
	Type          *model.TransactionType
	Currency      *string
	CategoryID    *model.CategoryID
	BankAccountID *model.BankAccountID
	AmountMin     *decimal.Decimal
	AmountMax     *decimal.Decimal
	SortBy        string
	SortOrder     string
}

// TransactionService handles transaction operations with permission checks.
type TransactionService struct {
	txStore  TransactionStoreIface
	wsStore  WorkspaceStoreIface
	catStore CategoryStoreIface
	ruleSvc  *RuleService
}

// NewTransactionService creates a new TransactionService.
func NewTransactionService(txStore TransactionStoreIface, wsStore WorkspaceStoreIface, catStore CategoryStoreIface, ruleSvc *RuleService) *TransactionService {
	return &TransactionService{
		txStore:  txStore,
		wsStore:  wsStore,
		catStore: catStore,
		ruleSvc:  ruleSvc,
	}
}

// requireMembership checks that a user is a member with at least the given role.
func (s *TransactionService) requireMembership(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, minRole model.MemberRole) error {
	member, err := s.wsStore.GetMember(ctx, wsID, userID)
	if err != nil {
		return err
	}
	if member.Role.Level() < minRole.Level() {
		return model.ErrInsufficientPermission
	}
	return nil
}

// ensureUncategorized returns the "Uncategorized" category ID for a workspace,
// creating it if it doesn't exist.
func (s *TransactionService) ensureUncategorized(ctx context.Context, wsID model.WorkspaceID) (model.CategoryID, error) {
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

// resolveEntryCategories fills in Uncategorized for any entry with CategoryID == 0.
func (s *TransactionService) resolveEntryCategories(ctx context.Context, wsID model.WorkspaceID, entries []CreateEntryInput) ([]CreateEntryInput, error) {
	var uncatID model.CategoryID
	for i, e := range entries {
		if e.CategoryID == 0 {
			if uncatID == 0 {
				var err error
				uncatID, err = s.ensureUncategorized(ctx, wsID)
				if err != nil {
					return nil, err
				}
			}
			entries[i].CategoryID = uncatID
		}
	}
	return entries, nil
}

// Create validates entries, checks editor+ permission, and atomically persists
// the transaction with its entries.
func (s *TransactionService) Create(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, input CreateTransactionInput) (model.Transaction, error) {
	if err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return model.Transaction{}, err
	}

	resolved, err := s.resolveEntryCategories(ctx, wsID, input.Entries)
	if err != nil {
		return model.Transaction{}, err
	}

	entries := make([]model.Entry, 0, len(resolved))
	for _, e := range resolved {
		entries = append(entries, model.Entry{
			CategoryID:    e.CategoryID,
			ParticipantID: e.ParticipantID,
			Amount:        e.Amount,
			Note:          e.Note,
		})
	}

	txType := input.Type
	if txType == "" {
		txType = model.TypeExpense
	}

	t := model.Transaction{
		WorkspaceID: wsID,
		CreatedBy:   userID,
		TotalAmount: input.TotalAmount,
		Description: input.Description,
		Date:        input.Date,
		Source:      model.SourceManual,
		Type:        txType,
		Notes:       input.Notes,
		Entries:     entries,
	}

	if err := t.ValidateEntries(); err != nil {
		return model.Transaction{}, err
	}

	return s.txStore.Create(ctx, t)
}

// GetTransaction returns a transaction with entries. Requires viewer+ role.
func (s *TransactionService) GetTransaction(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, txID model.TransactionID) (model.Transaction, error) {
	if err := s.requireMembership(ctx, wsID, userID, model.RoleViewer); err != nil {
		return model.Transaction{}, err
	}
	tx, err := s.txStore.GetByID(ctx, txID)
	if err != nil {
		return model.Transaction{}, err
	}
	if tx.WorkspaceID != wsID {
		return model.Transaction{}, model.ErrTransactionNotFound
	}
	return tx, nil
}

// ListTransactions returns transactions in a date range. Requires viewer+ role.
func (s *TransactionService) ListTransactions(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, input ListTransactionsInput) ([]model.Transaction, error) {
	if err := s.requireMembership(ctx, wsID, userID, model.RoleViewer); err != nil {
		return nil, err
	}
	return s.txStore.ListByWorkspace(ctx, wsID, input.From, input.To, input.Limit, input.Offset, store.ListByWorkspaceFilter{
		Type:          input.Type,
		Currency:      input.Currency,
		CategoryID:    input.CategoryID,
		BankAccountID: input.BankAccountID,
		AmountMin:     input.AmountMin,
		AmountMax:     input.AmountMax,
		SortBy:        input.SortBy,
		SortOrder:     input.SortOrder,
	})
}

// DeleteTransaction deletes a transaction. Requires editor+ role.
func (s *TransactionService) DeleteTransaction(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, txID model.TransactionID) error {
	if err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return err
	}
	tx, err := s.txStore.GetByID(ctx, txID)
	if err != nil {
		return err
	}
	if tx.WorkspaceID != wsID {
		return model.ErrTransactionNotFound
	}
	return s.txStore.Delete(ctx, txID)
}

// UpdateTransaction updates a transaction's fields and entries. Requires editor+ role.
func (s *TransactionService) UpdateTransaction(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, txID model.TransactionID, input UpdateTransactionInput) (model.Transaction, error) {
	if err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return model.Transaction{}, err
	}

	existing, err := s.txStore.GetByID(ctx, txID)
	if err != nil {
		return model.Transaction{}, err
	}
	if existing.WorkspaceID != wsID {
		return model.Transaction{}, model.ErrTransactionNotFound
	}

	resolved, err := s.resolveEntryCategories(ctx, wsID, input.Entries)
	if err != nil {
		return model.Transaction{}, err
	}

	entries := make([]model.Entry, 0, len(resolved))
	for _, e := range resolved {
		entries = append(entries, model.Entry{
			CategoryID:    e.CategoryID,
			ParticipantID: e.ParticipantID,
			Amount:        e.Amount,
			Note:          e.Note,
		})
	}

	txType := input.Type
	if txType == "" {
		txType = model.TypeExpense
	}

	t := model.Transaction{
		TotalAmount: existing.TotalAmount,
		Description: input.Description,
		Date:        input.Date,
		Type:        txType,
		Notes:       input.Notes,
		Entries:     entries,
	}

	if err := t.ValidateEntries(); err != nil {
		return model.Transaction{}, err
	}

	return s.txStore.Update(ctx, txID, t)
}

// ImportTransactions imports a batch of transactions, skipping duplicates
// based on fingerprint. Requires editor+ role.
func (s *TransactionService) ImportTransactions(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, inputs []ImportTransactionInput) (ImportResult, error) {
	if err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return ImportResult{}, err
	}

	var result ImportResult

	for _, input := range inputs {
		exists, err := s.txStore.ExistsByFingerprint(ctx, wsID, input.Fingerprint)
		if err != nil {
			return ImportResult{}, fmt.Errorf("checking fingerprint: %w", err)
		}
		if exists {
			result.Skipped++
			continue
		}

		// Auto-categorize using rule engine if available (skip income).
		importType := input.Type
		if importType == "" {
			importType = model.TypeExpense
		}

		var resolvedCatID *model.CategoryID
		if s.ruleSvc != nil && importType != model.TypeIncome {
			catID, err := s.ruleSvc.ResolveCategory(ctx, wsID, userID, input.Description, input.TotalAmount.Amount(), nil, nil)
			if err != nil {
				return ImportResult{}, fmt.Errorf("resolving category: %w", err)
			}
			resolvedCatID = catID
		}

		entries := make([]model.Entry, 0, len(input.Entries))
		for _, e := range input.Entries {
			catID := e.CategoryID
			if resolvedCatID != nil && catID == 0 {
				catID = *resolvedCatID
			}
			entries = append(entries, model.Entry{
				CategoryID:    catID,
				ParticipantID: e.ParticipantID,
				Amount:        e.Amount,
				Note:          e.Note,
			})
		}

		fp := input.Fingerprint
		t := model.Transaction{
			WorkspaceID: wsID,
			CreatedBy:   userID,
			TotalAmount: input.TotalAmount,
			Description: input.Description,
			Date:        input.Date,
			Fingerprint: &fp,
			Source:      model.SourceImport,
			Type:        importType,
			Notes:       input.Notes,
			Entries:     entries,
		}

		if err := t.ValidateEntries(); err != nil {
			return ImportResult{}, err
		}

		created, err := s.txStore.Create(ctx, t)
		if err != nil {
			return ImportResult{}, fmt.Errorf("creating imported transaction: %w", err)
		}

		result.Imported++
		result.Transactions = append(result.Transactions, created)
	}

	if result.Transactions == nil {
		result.Transactions = []model.Transaction{}
	}

	return result, nil
}

// maxBulkCategorizeIDs is the maximum number of transaction IDs per bulk categorize request.
const maxBulkCategorizeIDs = 200

// BulkCategorizeTransactions updates the first entry's category for multiple transactions.
// Requires editor+ role. Validates category ownership and that all IDs belong to the workspace.
func (s *TransactionService) BulkCategorizeTransactions(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, ids []model.TransactionID, categoryID model.CategoryID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	if len(ids) > maxBulkCategorizeIDs {
		return 0, fmt.Errorf("too many IDs: maximum %d per request", maxBulkCategorizeIDs)
	}

	if err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return 0, err
	}

	// Validate category exists and belongs to workspace.
	cat, err := s.catStore.GetByID(ctx, categoryID)
	if err != nil {
		return 0, err
	}
	if cat.WorkspaceID != wsID {
		return 0, model.ErrCategoryNotFound
	}

	// Verify all transaction IDs belong to the workspace (IDOR prevention).
	txs, err := s.txStore.GetByIDs(ctx, wsID, ids)
	if err != nil {
		return 0, fmt.Errorf("verifying transaction ownership: %w", err)
	}
	if len(txs) != len(ids) {
		return 0, model.ErrTransactionNotFound
	}

	updated, err := s.txStore.BulkCategorizeFirstEntry(ctx, ids, categoryID)
	if err != nil {
		return 0, fmt.Errorf("bulk categorizing: %w", err)
	}
	return updated, nil
}

// BulkDeleteTransactions deletes multiple transactions by ID. Requires editor+ role.
// The SQL query is scoped to workspace_id so IDs from other workspaces are silently ignored.
// Returns an error if the count of deleted rows doesn't match the requested count (IDOR).
func (s *TransactionService) BulkDeleteTransactions(ctx context.Context, userID model.UserID, wsID model.WorkspaceID, ids []model.TransactionID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	if len(ids) > maxBulkCategorizeIDs {
		return 0, fmt.Errorf("too many IDs: maximum %d per request", maxBulkCategorizeIDs)
	}

	if err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return 0, err
	}

	// Verify all IDs belong to this workspace before deleting.
	txs, err := s.txStore.GetByIDs(ctx, wsID, ids)
	if err != nil {
		return 0, fmt.Errorf("verifying transaction ownership: %w", err)
	}
	if len(txs) != len(ids) {
		return 0, model.ErrTransactionNotFound
	}

	deleted, err := s.txStore.BulkDelete(ctx, wsID, ids)
	if err != nil {
		return 0, fmt.Errorf("bulk deleting: %w", err)
	}
	return deleted, nil
}

// ApplyRulesResult is the result of applying categorization rules.
type ApplyRulesResult struct {
	Updated int `json:"updated"`
	Total   int `json:"total"`
}

// ApplyRules re-categorizes all transactions by matching rules against descriptions.
// Entries that match a rule get that category; entries that don't match any rule
// get set to "Uncategorized". Requires editor+ role.
func (s *TransactionService) ApplyRules(ctx context.Context, userID model.UserID, wsID model.WorkspaceID) (ApplyRulesResult, error) {
	if err := s.requireMembership(ctx, wsID, userID, model.RoleEditor); err != nil {
		return ApplyRulesResult{}, err
	}

	entries, err := s.txStore.ListAllEntries(ctx, wsID)
	if err != nil {
		return ApplyRulesResult{}, fmt.Errorf("listing entries: %w", err)
	}

	uncatID, err := s.ensureUncategorized(ctx, wsID)
	if err != nil {
		return ApplyRulesResult{}, err
	}

	var updated int
	for _, e := range entries {
		if e.Type == model.TypeTransfer || e.Type == model.TypeIncome {
			continue
		}

		catID, err := s.ruleSvc.ResolveCategory(ctx, wsID, e.CreatedBy, e.Description, e.TotalAmount, e.BankAccountID, e.CounterpartyIBAN)
		if err != nil {
			return ApplyRulesResult{}, fmt.Errorf("resolving category: %w", err)
		}

		newCatID := uncatID
		if catID != nil {
			newCatID = *catID
		}

		if newCatID == e.CategoryID {
			continue
		}

		if err := s.txStore.UpdateEntryCategory(ctx, e.EntryID, newCatID); err != nil {
			return ApplyRulesResult{}, fmt.Errorf("updating entry category: %w", err)
		}
		updated++
	}

	return ApplyRulesResult{Total: len(entries), Updated: updated}, nil
}
