package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"backend/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// TransactionStore provides transaction and entry persistence operations.
// It holds a pool reference (not just DBTX) because atomic creates need Begin().
type TransactionStore struct {
	pool *pgxpool.Pool
	q    *Queries
}

// NewTransactionStore creates a new TransactionStore.
func NewTransactionStore(pool *pgxpool.Pool) *TransactionStore {
	return &TransactionStore{
		pool: pool,
		q:    New(pool),
	}
}

// Create atomically inserts a transaction and its entries within a single
// database transaction.
func (s *TransactionStore) Create(ctx context.Context, t model.Transaction) (model.Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("rollback failed", "error", err)
		}
	}()

	qtx := New(tx)

	row, err := qtx.InsertTransaction(ctx, InsertTransactionParams{
		WorkspaceID:      int64(t.WorkspaceID),
		CreatedBy:        int64(t.CreatedBy),
		TotalAmount:      decimalToNumeric(t.TotalAmount.Amount()),
		Currency:         t.TotalAmount.Currency(),
		Description:      t.Description,
		Date:             timeToPgDate(t.Date),
		Fingerprint:      stringPtrToText(t.Fingerprint),
		Source:           TransactionSource(t.Source),
		TransactionType:  TransactionType(t.Type),
		Notes:            t.Notes,
		BankAccountID:    bankAccountIDToPgInt8(t.BankAccountID),
		Iban:             stringPtrToText(t.IBAN),
		CounterpartyIban: stringPtrToText(t.CounterpartyIBAN),
	})
	if err != nil {
		return model.Transaction{}, fmt.Errorf("inserting transaction: %w", err)
	}

	entries := make([]model.Entry, 0, len(t.Entries))
	for _, e := range t.Entries {
		entryRow, err := qtx.InsertEntry(ctx, InsertEntryParams{
			TransactionID: row.ID,
			CategoryID:    int64(e.CategoryID),
			ParticipantID: userIDPtrToInt8(e.ParticipantID),
			Amount:        decimalToNumeric(e.Amount.Amount()),
			Currency:      e.Amount.Currency(),
			Note:          e.Note,
		})
		if err != nil {
			return model.Transaction{}, fmt.Errorf("inserting entry: %w", err)
		}
		entries = append(entries, toModelEntry(entryRow))
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Transaction{}, fmt.Errorf("committing transaction: %w", err)
	}

	result := toModelTransactionFromInsert(row)
	result.Entries = entries
	return result, nil
}

// GetByID returns a transaction with its entries. Returns model.ErrTransactionNotFound if not found.
func (s *TransactionStore) GetByID(ctx context.Context, id model.TransactionID) (model.Transaction, error) {
	row, err := s.q.GetTransactionByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Transaction{}, model.ErrTransactionNotFound
		}
		return model.Transaction{}, fmt.Errorf("getting transaction by id: %w", err)
	}

	entryRows, err := s.q.ListEntriesByTransaction(ctx, row.ID)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("listing entries: %w", err)
	}

	t := toModelTransactionFromGet(row)
	t.Entries = make([]model.Entry, 0, len(entryRows))
	for _, er := range entryRows {
		t.Entries = append(t.Entries, toModelEntry(er))
	}
	return t, nil
}

// ListByWorkspaceFilter holds optional filters for listing transactions.
type ListByWorkspaceFilter struct {
	Type          *model.TransactionType
	Currency      *string
	CategoryID    *model.CategoryID
	BankAccountID *model.BankAccountID
	AmountMin     *decimal.Decimal
	AmountMax     *decimal.Decimal
	SortBy        string // "date", "amount", "description" (default: "date")
	SortOrder     string // "asc", "desc" (default: "desc")
}

// sortColumnWhitelist maps allowed sort keys to safe SQL column expressions.
var sortColumnWhitelist = map[string]string{
	"date":        "t.date",
	"amount":      "ABS(t.total_amount)",
	"description": "t.description",
}

// ListByWorkspace returns transactions in a date range with pagination and optional filters.
func (s *TransactionStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID, from, to time.Time, limit, offset int32, filter ListByWorkspaceFilter) ([]model.Transaction, error) {
	var typeFilter NullTransactionType
	if filter.Type != nil {
		typeFilter = NullTransactionType{TransactionType: TransactionType(*filter.Type), Valid: true}
	}

	var currencyFilter pgtype.Text
	if filter.Currency != nil {
		currencyFilter = pgtype.Text{String: *filter.Currency, Valid: true}
	}

	var categoryFilter pgtype.Int8
	if filter.CategoryID != nil {
		categoryFilter = pgtype.Int8{Int64: int64(*filter.CategoryID), Valid: true}
	}

	var bankAccountFilter pgtype.Int8
	if filter.BankAccountID != nil {
		bankAccountFilter = pgtype.Int8{Int64: int64(*filter.BankAccountID), Valid: true}
	}

	var amountMinFilter pgtype.Numeric
	if filter.AmountMin != nil {
		amountMinFilter = pgtype.Numeric{
			Int:   filter.AmountMin.Coefficient(),
			Exp:   filter.AmountMin.Exponent(),
			Valid: true,
		}
	}

	var amountMaxFilter pgtype.Numeric
	if filter.AmountMax != nil {
		amountMaxFilter = pgtype.Numeric{
			Int:   filter.AmountMax.Coefficient(),
			Exp:   filter.AmountMax.Exponent(),
			Valid: true,
		}
	}

	// Resolve sort column from whitelist (safe against injection).
	sortCol := "t.date"
	if col, ok := sortColumnWhitelist[filter.SortBy]; ok {
		sortCol = col
	}
	sortDir := "DESC"
	if filter.SortOrder == "asc" {
		sortDir = "ASC"
	}

	query := fmt.Sprintf(`SELECT t.id, t.workspace_id, t.created_by, t.total_amount, t.currency, t.description, t.date, t.fingerprint, t.source, t.transaction_type, t.notes, t.bank_account_id, t.iban, t.counterparty_iban, t.created_at, t.updated_at,
       COALESCE(bc.institution_name, '') AS bank_name
FROM transactions t
LEFT JOIN bank_accounts ba ON ba.id = t.bank_account_id
LEFT JOIN bank_connections bc ON bc.id = ba.bank_connection_id
WHERE t.workspace_id = $1 AND t.date >= $2 AND t.date < $3
  AND ($6::transaction_type IS NULL OR t.transaction_type = $6)
  AND ($7::text IS NULL OR t.currency = $7)
  AND ($8::bigint IS NULL OR EXISTS (
    SELECT 1 FROM entries e WHERE e.transaction_id = t.id AND e.category_id = $8
  ))
  AND ($9::bigint IS NULL OR t.bank_account_id = $9)
  AND ($10::numeric IS NULL OR ABS(t.total_amount) >= $10)
  AND ($11::numeric IS NULL OR ABS(t.total_amount) <= $11)
ORDER BY %s %s, t.id DESC
LIMIT $4 OFFSET $5`, sortCol, sortDir)

	rows, err := s.pool.Query(ctx, query,
		int64(wsID),
		timeToPgDate(from),
		timeToPgDate(to),
		limit,
		offset,
		typeFilter,
		currencyFilter,
		categoryFilter,
		bankAccountFilter,
		amountMinFilter,
		amountMaxFilter,
	)
	if err != nil {
		return nil, fmt.Errorf("listing transactions: %w", err)
	}
	defer rows.Close()

	var txRows []ListTransactionsByWorkspaceRow
	for rows.Next() {
		var i ListTransactionsByWorkspaceRow
		if err := rows.Scan(
			&i.ID,
			&i.WorkspaceID,
			&i.CreatedBy,
			&i.TotalAmount,
			&i.Currency,
			&i.Description,
			&i.Date,
			&i.Fingerprint,
			&i.Source,
			&i.TransactionType,
			&i.Notes,
			&i.BankAccountID,
			&i.Iban,
			&i.CounterpartyIban,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.BankName,
		); err != nil {
			return nil, fmt.Errorf("scanning transaction row: %w", err)
		}
		txRows = append(txRows, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transaction rows: %w", err)
	}

	// Collect transaction IDs for batch entry loading.
	txIDs := make([]int64, 0, len(txRows))
	for _, row := range txRows {
		txIDs = append(txIDs, row.ID)
	}

	// Batch-load entries for all transactions.
	entryMap := make(map[int64][]model.Entry)
	if len(txIDs) > 0 {
		entryRows, err := s.q.ListEntriesByTransactionIDs(ctx, txIDs)
		if err != nil {
			return nil, fmt.Errorf("loading entries: %w", err)
		}
		for _, er := range entryRows {
			entryMap[er.TransactionID] = append(entryMap[er.TransactionID], toModelEntry(er))
		}
	}

	result := make([]model.Transaction, 0, len(txRows))
	for _, row := range txRows {
		t := toModelTransactionFromList(row)
		if entries, ok := entryMap[row.ID]; ok {
			t.Entries = entries
		} else {
			t.Entries = []model.Entry{}
		}
		result = append(result, t)
	}
	return result, nil
}

// Update atomically updates a transaction's fields and replaces its entries.
func (s *TransactionStore) Update(ctx context.Context, id model.TransactionID, t model.Transaction) (model.Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return model.Transaction{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("rollback failed", "error", err)
		}
	}()

	qtx := New(tx)

	row, err := qtx.UpdateTransaction(ctx, UpdateTransactionParams{
		ID:              int64(id),
		Description:     t.Description,
		Date:            timeToPgDate(t.Date),
		TransactionType: TransactionType(t.Type),
		Notes:           t.Notes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Transaction{}, model.ErrTransactionNotFound
		}
		return model.Transaction{}, fmt.Errorf("updating transaction: %w", err)
	}

	if err := qtx.DeleteEntriesByTransaction(ctx, int64(id)); err != nil {
		return model.Transaction{}, fmt.Errorf("deleting old entries: %w", err)
	}

	entries := make([]model.Entry, 0, len(t.Entries))
	for _, e := range t.Entries {
		entryRow, err := qtx.InsertEntry(ctx, InsertEntryParams{
			TransactionID: row.ID,
			CategoryID:    int64(e.CategoryID),
			ParticipantID: userIDPtrToInt8(e.ParticipantID),
			Amount:        decimalToNumeric(e.Amount.Amount()),
			Currency:      e.Amount.Currency(),
			Note:          e.Note,
		})
		if err != nil {
			return model.Transaction{}, fmt.Errorf("inserting entry: %w", err)
		}
		entries = append(entries, toModelEntry(entryRow))
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Transaction{}, fmt.Errorf("committing transaction: %w", err)
	}

	result := toModelTransactionFromUpdate(row)
	result.Entries = entries
	return result, nil
}

// Delete removes a transaction by ID. Entries are removed by CASCADE.
func (s *TransactionStore) Delete(ctx context.Context, id model.TransactionID) error {
	if err := s.q.DeleteTransaction(ctx, int64(id)); err != nil {
		return fmt.Errorf("deleting transaction: %w", err)
	}
	return nil
}

// BulkDelete removes transactions by IDs within a workspace. Entries are removed by CASCADE.
func (s *TransactionStore) BulkDelete(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) (int64, error) {
	int64IDs := make([]int64, len(ids))
	for i, id := range ids {
		int64IDs[i] = int64(id)
	}
	result, err := s.q.BulkDeleteTransactions(ctx, BulkDeleteTransactionsParams{
		WorkspaceID: int64(wsID),
		Column2:     int64IDs,
	})
	if err != nil {
		return 0, fmt.Errorf("bulk deleting transactions: %w", err)
	}
	return result.RowsAffected(), nil
}

// ExistsByFingerprint checks if a transaction with the given fingerprint
// already exists in the workspace.
func (s *TransactionStore) ExistsByFingerprint(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error) {
	exists, err := s.q.ExistsByFingerprint(ctx, ExistsByFingerprintParams{
		WorkspaceID: int64(wsID),
		Fingerprint: pgtype.Text{String: fingerprint, Valid: true},
	})
	if err != nil {
		return false, fmt.Errorf("checking fingerprint: %w", err)
	}
	return exists, nil
}

// RelinkOrphanedTransactions re-links orphaned transactions (bank_account_id IS NULL) to a bank account
// by matching iban + currency + workspace.
func (s *TransactionStore) RelinkOrphanedTransactions(ctx context.Context, bankAccountID model.BankAccountID, iban, currency string, wsID model.WorkspaceID) (int64, error) {
	result, err := s.q.RelinkOrphanedTransactions(ctx, RelinkOrphanedTransactionsParams{
		BankAccountID: pgtype.Int8{Int64: int64(bankAccountID), Valid: true},
		Iban:          pgtype.Text{String: iban, Valid: true},
		Currency:      currency,
		WorkspaceID:   int64(wsID),
	})
	if err != nil {
		return 0, fmt.Errorf("relinking orphaned transactions: %w", err)
	}
	return result.RowsAffected(), nil
}

// ReclassifyTransfersByCounterpartyIBAN updates transactions whose counterparty_iban
// matches one of the given IBANs, changing their type to 'transfer'.
// Only non-transfer transactions in the workspace are affected.
func (s *TransactionStore) ReclassifyTransfersByCounterpartyIBAN(ctx context.Context, wsID model.WorkspaceID, ibans []string) (int64, error) {
	if len(ibans) == 0 {
		return 0, nil
	}
	query := `UPDATE transactions
SET transaction_type = 'transfer', updated_at = NOW()
WHERE workspace_id = $1
  AND counterparty_iban = ANY($2)
  AND transaction_type != 'transfer'`
	result, err := s.pool.Exec(ctx, query, int64(wsID), ibans)
	if err != nil {
		return 0, fmt.Errorf("reclassifying transfers by counterparty IBAN: %w", err)
	}
	return result.RowsAffected(), nil
}

// ── pgtype conversion helpers ──

func decimalToNumeric(d decimal.Decimal) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(d.String())
	return n
}

func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	// pgtype.Numeric stores Int (*big.Int) and Exp (int32),
	// which maps directly to shopspring/decimal.
	return decimal.NewFromBigInt(n.Int, n.Exp)
}

func timeToPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func stringPtrToText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func textToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func userIDPtrToInt8(uid *model.UserID) pgtype.Int8 {
	if uid == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: int64(*uid), Valid: true}
}

func int8ToUserIDPtr(i pgtype.Int8) *model.UserID {
	if !i.Valid {
		return nil
	}
	uid := model.UserID(i.Int64)
	return &uid
}

func toModelTransactionFromInsert(row InsertTransactionRow) model.Transaction {
	return model.Transaction{
		ID:               model.TransactionID(row.ID),
		WorkspaceID:      model.WorkspaceID(row.WorkspaceID),
		CreatedBy:        model.UserID(row.CreatedBy),
		TotalAmount:      model.NewMoney(numericToDecimal(row.TotalAmount), row.Currency),
		Description:      row.Description,
		Date:             row.Date.Time,
		Fingerprint:      textToStringPtr(row.Fingerprint),
		Source:           model.TransactionSource(row.Source),
		Type:             model.TransactionType(row.TransactionType),
		BankAccountID:    int8ToBankAccountIDPtr(row.BankAccountID),
		IBAN:             textToStringPtr(row.Iban),
		CounterpartyIBAN: textToStringPtr(row.CounterpartyIban),
		Notes:            row.Notes,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func toModelTransactionFromGet(row GetTransactionByIDRow) model.Transaction {
	return model.Transaction{
		ID:               model.TransactionID(row.ID),
		WorkspaceID:      model.WorkspaceID(row.WorkspaceID),
		CreatedBy:        model.UserID(row.CreatedBy),
		TotalAmount:      model.NewMoney(numericToDecimal(row.TotalAmount), row.Currency),
		Description:      row.Description,
		Date:             row.Date.Time,
		Fingerprint:      textToStringPtr(row.Fingerprint),
		Source:           model.TransactionSource(row.Source),
		Type:             model.TransactionType(row.TransactionType),
		BankAccountID:    int8ToBankAccountIDPtr(row.BankAccountID),
		IBAN:             textToStringPtr(row.Iban),
		CounterpartyIBAN: textToStringPtr(row.CounterpartyIban),
		Notes:            row.Notes,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func toModelTransactionFromList(row ListTransactionsByWorkspaceRow) model.Transaction {
	return model.Transaction{
		ID:               model.TransactionID(row.ID),
		WorkspaceID:      model.WorkspaceID(row.WorkspaceID),
		CreatedBy:        model.UserID(row.CreatedBy),
		TotalAmount:      model.NewMoney(numericToDecimal(row.TotalAmount), row.Currency),
		Description:      row.Description,
		Date:             row.Date.Time,
		Fingerprint:      textToStringPtr(row.Fingerprint),
		Source:           model.TransactionSource(row.Source),
		Type:             model.TransactionType(row.TransactionType),
		BankAccountID:    int8ToBankAccountIDPtr(row.BankAccountID),
		IBAN:             textToStringPtr(row.Iban),
		CounterpartyIBAN: textToStringPtr(row.CounterpartyIban),
		BankName:         row.BankName,
		Notes:            row.Notes,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func toModelTransactionFromUpdate(row UpdateTransactionRow) model.Transaction {
	return model.Transaction{
		ID:               model.TransactionID(row.ID),
		WorkspaceID:      model.WorkspaceID(row.WorkspaceID),
		CreatedBy:        model.UserID(row.CreatedBy),
		TotalAmount:      model.NewMoney(numericToDecimal(row.TotalAmount), row.Currency),
		Description:      row.Description,
		Date:             row.Date.Time,
		Fingerprint:      textToStringPtr(row.Fingerprint),
		Source:           model.TransactionSource(row.Source),
		Type:             model.TransactionType(row.TransactionType),
		BankAccountID:    int8ToBankAccountIDPtr(row.BankAccountID),
		IBAN:             textToStringPtr(row.Iban),
		CounterpartyIBAN: textToStringPtr(row.CounterpartyIban),
		Notes:            row.Notes,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func bankAccountIDToPgInt8(id *model.BankAccountID) pgtype.Int8 {
	if id == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: int64(*id), Valid: true}
}

func int8ToBankAccountIDPtr(i pgtype.Int8) *model.BankAccountID {
	if !i.Valid {
		return nil
	}
	id := model.BankAccountID(i.Int64)
	return &id
}

// TransactionEntryRow represents a transaction entry with its parent transaction's description.
type TransactionEntryRow struct {
	TransactionID    model.TransactionID
	Description      string
	CreatedBy        model.UserID
	EntryID          model.EntryID
	CategoryID       model.CategoryID
	Type             model.TransactionType
	TotalAmount      decimal.Decimal
	BankAccountID    *model.BankAccountID
	CounterpartyIBAN *string
}

// ListAllEntries returns all entries for a workspace, joined with their transaction descriptions.
func (s *TransactionStore) ListAllEntries(ctx context.Context, wsID model.WorkspaceID) ([]TransactionEntryRow, error) {
	rows, err := s.q.ListAllTransactionEntries(ctx, int64(wsID))
	if err != nil {
		return nil, fmt.Errorf("listing transaction entries: %w", err)
	}
	result := make([]TransactionEntryRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, TransactionEntryRow{
			TransactionID:    model.TransactionID(row.ID),
			Description:      row.Description,
			CreatedBy:        model.UserID(row.CreatedBy),
			EntryID:          model.EntryID(row.EntryID),
			CategoryID:       model.CategoryID(row.CategoryID),
			Type:             model.TransactionType(row.TransactionType),
			TotalAmount:      numericToDecimal(row.TotalAmount),
			BankAccountID:    int8ToBankAccountIDPtr(row.BankAccountID),
			CounterpartyIBAN: textToStringPtr(row.CounterpartyIban),
		})
	}
	return result, nil
}

// GetByIDs returns transactions matching the given IDs within a workspace.
func (s *TransactionStore) GetByIDs(ctx context.Context, wsID model.WorkspaceID, ids []model.TransactionID) ([]model.Transaction, error) {
	int64IDs := make([]int64, len(ids))
	for i, id := range ids {
		int64IDs[i] = int64(id)
	}
	rows, err := s.q.GetTransactionsByIDs(ctx, GetTransactionsByIDsParams{
		WorkspaceID: int64(wsID),
		Column2:     int64IDs,
	})
	if err != nil {
		return nil, fmt.Errorf("getting transactions by ids: %w", err)
	}
	result := make([]model.Transaction, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelTransactionFromGet(GetTransactionByIDRow(row)))
	}
	return result, nil
}

// BulkCategorizeFirstEntry updates the first entry's category for each given transaction.
func (s *TransactionStore) BulkCategorizeFirstEntry(ctx context.Context, txIDs []model.TransactionID, catID model.CategoryID) (int64, error) {
	int64IDs := make([]int64, len(txIDs))
	for i, id := range txIDs {
		int64IDs[i] = int64(id)
	}
	result, err := s.q.BulkUpdateFirstEntryCategory(ctx, BulkUpdateFirstEntryCategoryParams{
		CategoryID: int64(catID),
		Column2:    int64IDs,
	})
	if err != nil {
		return 0, fmt.Errorf("bulk updating entry categories: %w", err)
	}
	return result.RowsAffected(), nil
}

// UpdateEntryCategory updates a single entry's category.
func (s *TransactionStore) UpdateEntryCategory(ctx context.Context, entryID model.EntryID, catID model.CategoryID) error {
	return s.q.UpdateEntryCategory(ctx, UpdateEntryCategoryParams{
		ID:         int64(entryID),
		CategoryID: int64(catID),
	})
}

func toModelEntry(row Entry) model.Entry {
	return model.Entry{
		ID:            model.EntryID(row.ID),
		TransactionID: model.TransactionID(row.TransactionID),
		CategoryID:    model.CategoryID(row.CategoryID),
		ParticipantID: int8ToUserIDPtr(row.ParticipantID),
		Amount:        model.NewMoney(numericToDecimal(row.Amount), row.Currency),
		Note:          row.Note,
	}
}
