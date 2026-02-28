package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/banksync"
	"backend/model"
	"backend/store"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// ── Mock: BankSyncTransactionStoreIface ──

type mockBankSyncTxStore struct {
	createFunc                     func(ctx context.Context, t model.Transaction) (model.Transaction, error)
	existsByFingerprintFunc        func(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error)
	relinkOrphanedTransactionsFunc func(ctx context.Context, bankAccountID model.BankAccountID, iban, currency string, wsID model.WorkspaceID) (int64, error)
}

func (m *mockBankSyncTxStore) Create(ctx context.Context, t model.Transaction) (model.Transaction, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, t)
	}
	return t, nil
}

func (m *mockBankSyncTxStore) ExistsByFingerprint(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error) {
	if m.existsByFingerprintFunc != nil {
		return m.existsByFingerprintFunc(ctx, wsID, fingerprint)
	}
	return false, nil
}

func (m *mockBankSyncTxStore) RelinkOrphanedTransactions(ctx context.Context, bankAccountID model.BankAccountID, iban, currency string, wsID model.WorkspaceID) (int64, error) {
	if m.relinkOrphanedTransactionsFunc != nil {
		return m.relinkOrphanedTransactionsFunc(ctx, bankAccountID, iban, currency, wsID)
	}
	return 0, nil
}

// ── Mock: BankConnectionStoreIface ──

type mockBankConnStore struct {
	createFunc        func(ctx context.Context, bc model.BankConnection) (model.BankConnection, error)
	getByIDFunc       func(ctx context.Context, id model.BankConnectionID) (model.BankConnection, error)
	listByUserFunc    func(ctx context.Context, userID model.UserID) ([]model.BankConnection, error)
	listActiveFunc    func(ctx context.Context) ([]model.BankConnection, error)
	updateStatusFunc  func(ctx context.Context, id model.BankConnectionID, status model.BankConnectionStatus) error
	updateSessionFunc func(ctx context.Context, id model.BankConnectionID, sessionID string, status model.BankConnectionStatus, expiresAt pgtype.Timestamptz) error
	deleteFunc        func(ctx context.Context, id model.BankConnectionID, userID model.UserID) error
}

func (m *mockBankConnStore) Create(ctx context.Context, bc model.BankConnection) (model.BankConnection, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, bc)
	}
	return bc, nil
}

func (m *mockBankConnStore) GetByID(ctx context.Context, id model.BankConnectionID) (model.BankConnection, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return model.BankConnection{}, nil
}

func (m *mockBankConnStore) ListByUser(ctx context.Context, userID model.UserID) ([]model.BankConnection, error) {
	if m.listByUserFunc != nil {
		return m.listByUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockBankConnStore) ListActive(ctx context.Context) ([]model.BankConnection, error) {
	if m.listActiveFunc != nil {
		return m.listActiveFunc(ctx)
	}
	return nil, nil
}

func (m *mockBankConnStore) UpdateStatus(ctx context.Context, id model.BankConnectionID, status model.BankConnectionStatus) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *mockBankConnStore) UpdateSession(ctx context.Context, id model.BankConnectionID, sessionID string, status model.BankConnectionStatus, expiresAt pgtype.Timestamptz) error {
	if m.updateSessionFunc != nil {
		return m.updateSessionFunc(ctx, id, sessionID, status, expiresAt)
	}
	return nil
}

func (m *mockBankConnStore) Delete(ctx context.Context, id model.BankConnectionID, userID model.UserID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id, userID)
	}
	return nil
}

// ── Mock: BankAccountStoreIface ──

type mockBankAcctStore struct {
	createFunc                      func(ctx context.Context, ba model.BankAccount) (model.BankAccount, error)
	getByIDFunc                     func(ctx context.Context, id model.BankAccountID) (model.BankAccount, error)
	listByConnectionFunc            func(ctx context.Context, connID model.BankConnectionID) ([]model.BankAccount, error)
	listByWorkspaceFunc             func(ctx context.Context, wsID model.WorkspaceID) ([]store.BankAccountWithInstitution, error)
	listByUserFunc                  func(ctx context.Context, userID model.UserID) ([]store.BankAccountWithInstitution, error)
	linkToWorkspaceFunc             func(ctx context.Context, wsID model.WorkspaceID, baID model.BankAccountID) error
	unlinkFromWorkspaceFunc         func(ctx context.Context, wsID model.WorkspaceID, baID model.BankAccountID) error
	listWorkspacesByBankAccountFunc func(ctx context.Context, baID model.BankAccountID) ([]model.WorkspaceID, error)
	updateLastSyncedAtFunc          func(ctx context.Context, id model.BankAccountID, t time.Time) error
	getByIBANAndCurrencyForUserFunc func(ctx context.Context, iban, currency string, userID model.UserID) (model.BankAccount, error)
	updateConnectionFunc            func(ctx context.Context, id model.BankAccountID, externalID string, connID model.BankConnectionID) error
}

func (m *mockBankAcctStore) Create(ctx context.Context, ba model.BankAccount) (model.BankAccount, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, ba)
	}
	return ba, nil
}

func (m *mockBankAcctStore) GetByID(ctx context.Context, id model.BankAccountID) (model.BankAccount, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return model.BankAccount{}, nil
}

func (m *mockBankAcctStore) ListByConnection(ctx context.Context, connID model.BankConnectionID) ([]model.BankAccount, error) {
	if m.listByConnectionFunc != nil {
		return m.listByConnectionFunc(ctx, connID)
	}
	return nil, nil
}

func (m *mockBankAcctStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]store.BankAccountWithInstitution, error) {
	if m.listByWorkspaceFunc != nil {
		return m.listByWorkspaceFunc(ctx, wsID)
	}
	return nil, nil
}

func (m *mockBankAcctStore) ListByUser(ctx context.Context, userID model.UserID) ([]store.BankAccountWithInstitution, error) {
	if m.listByUserFunc != nil {
		return m.listByUserFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockBankAcctStore) LinkToWorkspace(ctx context.Context, wsID model.WorkspaceID, baID model.BankAccountID) error {
	if m.linkToWorkspaceFunc != nil {
		return m.linkToWorkspaceFunc(ctx, wsID, baID)
	}
	return nil
}

func (m *mockBankAcctStore) UnlinkFromWorkspace(ctx context.Context, wsID model.WorkspaceID, baID model.BankAccountID) error {
	if m.unlinkFromWorkspaceFunc != nil {
		return m.unlinkFromWorkspaceFunc(ctx, wsID, baID)
	}
	return nil
}

func (m *mockBankAcctStore) ListWorkspacesByBankAccount(ctx context.Context, baID model.BankAccountID) ([]model.WorkspaceID, error) {
	if m.listWorkspacesByBankAccountFunc != nil {
		return m.listWorkspacesByBankAccountFunc(ctx, baID)
	}
	return nil, nil
}

func (m *mockBankAcctStore) UpdateLastSyncedAt(ctx context.Context, id model.BankAccountID, t time.Time) error {
	if m.updateLastSyncedAtFunc != nil {
		return m.updateLastSyncedAtFunc(ctx, id, t)
	}
	return nil
}

func (m *mockBankAcctStore) GetByIBANAndCurrencyForUser(ctx context.Context, iban, currency string, userID model.UserID) (model.BankAccount, error) {
	if m.getByIBANAndCurrencyForUserFunc != nil {
		return m.getByIBANAndCurrencyForUserFunc(ctx, iban, currency, userID)
	}
	return model.BankAccount{}, nil
}

func (m *mockBankAcctStore) UpdateConnection(ctx context.Context, id model.BankAccountID, externalID string, connID model.BankConnectionID) error {
	if m.updateConnectionFunc != nil {
		return m.updateConnectionFunc(ctx, id, externalID, connID)
	}
	return nil
}

func (m *mockBankAcctStore) UpdateCustomName(_ context.Context, id model.BankAccountID, customName string) (model.BankAccount, error) {
	return model.BankAccount{ID: id, CustomName: customName}, nil
}

// ── Mock: banksync.Client ──

type mockBankSyncClient struct {
	getASPSPsFunc              func(ctx context.Context, country string) ([]banksync.ASPSP, error)
	startAuthFunc              func(ctx context.Context, aspspName, country, redirectURL string, validUntil time.Time) (banksync.AuthResult, error)
	createSessionFunc          func(ctx context.Context, code string) (banksync.Session, error)
	getAccountTransactionsFunc func(ctx context.Context, accountUID string, dateFrom, dateTo time.Time) ([]banksync.Transaction, error)
}

func (m *mockBankSyncClient) GetASPSPs(ctx context.Context, country string) ([]banksync.ASPSP, error) {
	if m.getASPSPsFunc != nil {
		return m.getASPSPsFunc(ctx, country)
	}
	return nil, nil
}

func (m *mockBankSyncClient) StartAuth(ctx context.Context, aspspName, country, redirectURL string, validUntil time.Time) (banksync.AuthResult, error) {
	if m.startAuthFunc != nil {
		return m.startAuthFunc(ctx, aspspName, country, redirectURL, validUntil)
	}
	return banksync.AuthResult{}, nil
}

func (m *mockBankSyncClient) CreateSession(ctx context.Context, code string) (banksync.Session, error) {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, code)
	}
	return banksync.Session{}, nil
}

func (m *mockBankSyncClient) GetAccountTransactions(ctx context.Context, accountUID string, dateFrom, dateTo time.Time) ([]banksync.Transaction, error) {
	if m.getAccountTransactionsFunc != nil {
		return m.getAccountTransactionsFunc(ctx, accountUID, dateFrom, dateTo)
	}
	return nil, nil
}

// ── Helpers ──

func strPtr(s string) *string { return &s }

func newBankSyncService(
	connStore *mockBankConnStore,
	acctStore *mockBankAcctStore,
	txStore *mockBankSyncTxStore,
	catStore *mockCategoryStore,
	client *mockBankSyncClient,
) *BankSyncService {
	return NewBankSyncService(connStore, acctStore, txStore, catStore, nil, client)
}

// ── Tests: Self-Transfer Detection ──

func TestBankSyncService_SelfTransferDetection(t *testing.T) {
	tests := map[string]struct {
		ebTx                 banksync.Transaction
		ownIBANs             map[string]bool
		wantType             model.TransactionType
		wantCounterpartyIBAN *string // nil means expect nil, non-nil means expect that value
	}{
		"expense with counterparty IBAN matching own IBAN becomes transfer": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-1",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "-50.00", Currency: "PLN"},
				CreditDebitIndicator: "DBIT",
				Creditor:             banksync.PartyIdentification{Name: "My Savings"},
				CreditorAccount:      &banksync.AccountIdentification{IBAN: "PL11111111111111111111111111"},
			},
			ownIBANs: map[string]bool{
				"PL11111111111111111111111111": true,
			},
			wantType:             model.TypeTransfer,
			wantCounterpartyIBAN: strPtr("PL11111111111111111111111111"),
		},
		"income with debtor IBAN matching own IBAN becomes transfer": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-2",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "100.00", Currency: "PLN"},
				CreditDebitIndicator: "CRDT",
				Debtor:               banksync.PartyIdentification{Name: "My Checking"},
				DebtorAccount:        &banksync.AccountIdentification{IBAN: "PL22222222222222222222222222"},
			},
			ownIBANs: map[string]bool{
				"PL22222222222222222222222222": true,
			},
			wantType:             model.TypeTransfer,
			wantCounterpartyIBAN: strPtr("PL22222222222222222222222222"),
		},
		"expense with counterparty IBAN not in own set stays expense": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-3",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "-30.00", Currency: "PLN"},
				CreditDebitIndicator: "DBIT",
				Creditor:             banksync.PartyIdentification{Name: "Grocery Store"},
				CreditorAccount:      &banksync.AccountIdentification{IBAN: "PL99999999999999999999999999"},
			},
			ownIBANs: map[string]bool{
				"PL11111111111111111111111111": true,
			},
			wantType:             model.TypeExpense,
			wantCounterpartyIBAN: strPtr("PL99999999999999999999999999"),
		},
		"income with debtor IBAN not in own set stays income": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-4",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "2000.00", Currency: "PLN"},
				CreditDebitIndicator: "CRDT",
				Debtor:               banksync.PartyIdentification{Name: "Employer Inc."},
				DebtorAccount:        &banksync.AccountIdentification{IBAN: "PL88888888888888888888888888"},
			},
			ownIBANs: map[string]bool{
				"PL11111111111111111111111111": true,
			},
			wantType:             model.TypeIncome,
			wantCounterpartyIBAN: strPtr("PL88888888888888888888888888"),
		},
		"expense with no creditor account stays expense": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-5",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "-15.00", Currency: "PLN"},
				CreditDebitIndicator: "DBIT",
				Creditor:             banksync.PartyIdentification{Name: "ATM Withdrawal"},
			},
			ownIBANs: map[string]bool{
				"PL11111111111111111111111111": true,
			},
			wantType:             model.TypeExpense,
			wantCounterpartyIBAN: nil,
		},
		"income with no debtor account stays income": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-6",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "50.00", Currency: "PLN"},
				CreditDebitIndicator: "CRDT",
				Debtor:               banksync.PartyIdentification{Name: "Cash Deposit"},
			},
			ownIBANs: map[string]bool{
				"PL11111111111111111111111111": true,
			},
			wantType:             model.TypeIncome,
			wantCounterpartyIBAN: nil,
		},
		"expense with empty creditor IBAN stays expense": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-7",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "-10.00", Currency: "PLN"},
				CreditDebitIndicator: "DBIT",
				Creditor:             banksync.PartyIdentification{Name: "Some Shop"},
				CreditorAccount:      &banksync.AccountIdentification{IBAN: ""},
			},
			ownIBANs: map[string]bool{
				"PL11111111111111111111111111": true,
			},
			wantType:             model.TypeExpense,
			wantCounterpartyIBAN: nil,
		},
		"empty own IBAN set never matches transfer": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-8",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "-50.00", Currency: "PLN"},
				CreditDebitIndicator: "DBIT",
				Creditor:             banksync.PartyIdentification{Name: "Some Account"},
				CreditorAccount:      &banksync.AccountIdentification{IBAN: "PL11111111111111111111111111"},
			},
			ownIBANs:             map[string]bool{},
			wantType:             model.TypeExpense,
			wantCounterpartyIBAN: strPtr("PL11111111111111111111111111"),
		},
		"positive amount without indicator treated as income and detects self-transfer": {
			ebTx: banksync.Transaction{
				TransactionID:     "tx-9",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "200.00", Currency: "PLN"},
				Debtor:            banksync.PartyIdentification{Name: "My Other Account"},
				DebtorAccount:     &banksync.AccountIdentification{IBAN: "PL33333333333333333333333333"},
			},
			ownIBANs: map[string]bool{
				"PL33333333333333333333333333": true,
			},
			wantType:             model.TypeTransfer,
			wantCounterpartyIBAN: strPtr("PL33333333333333333333333333"),
		},
		"multiple own IBANs - matches any of them": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-10",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "-100.00", Currency: "PLN"},
				CreditDebitIndicator: "DBIT",
				Creditor:             banksync.PartyIdentification{Name: "My Third Account"},
				CreditorAccount:      &banksync.AccountIdentification{IBAN: "PL44444444444444444444444444"},
			},
			ownIBANs: map[string]bool{
				"PL11111111111111111111111111": true,
				"PL22222222222222222222222222": true,
				"PL33333333333333333333333333": true,
				"PL44444444444444444444444444": true,
			},
			wantType:             model.TypeTransfer,
			wantCounterpartyIBAN: strPtr("PL44444444444444444444444444"),
		},
	}

	svc := newBankSyncService(
		&mockBankConnStore{},
		&mockBankAcctStore{},
		&mockBankSyncTxStore{},
		&mockCategoryStore{},
		&mockBankSyncClient{},
	)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			bankAcctID := model.BankAccountID(1)
			bankIBAN := strPtr("PL00000000000000000000000000")
			catID := model.CategoryID(99)

			tx, err := svc.mapTransaction(tt.ebTx, 1, 1, bankAcctID, bankIBAN, catID, tt.ownIBANs)
			if err != nil {
				t.Fatalf("mapTransaction returned unexpected error: %v", err)
			}

			if tx.Type != tt.wantType {
				t.Errorf("want type %q, got %q", tt.wantType, tx.Type)
			}

			if tt.wantCounterpartyIBAN == nil {
				if tx.CounterpartyIBAN != nil {
					t.Errorf("want nil CounterpartyIBAN, got %q", *tx.CounterpartyIBAN)
				}
			} else {
				if tx.CounterpartyIBAN == nil {
					t.Errorf("want CounterpartyIBAN %q, got nil", *tt.wantCounterpartyIBAN)
				} else if *tx.CounterpartyIBAN != *tt.wantCounterpartyIBAN {
					t.Errorf("want CounterpartyIBAN %q, got %q", *tt.wantCounterpartyIBAN, *tx.CounterpartyIBAN)
				}
			}
		})
	}
}

// ── Tests: extractCounterpartyIBAN ──

func TestBankSyncService_ExtractCounterpartyIBAN(t *testing.T) {
	svc := newBankSyncService(
		&mockBankConnStore{},
		&mockBankAcctStore{},
		&mockBankSyncTxStore{},
		&mockCategoryStore{},
		&mockBankSyncClient{},
	)

	tests := map[string]struct {
		ebTx     banksync.Transaction
		txType   model.TransactionType
		wantIBAN string
	}{
		"expense returns creditor IBAN": {
			ebTx: banksync.Transaction{
				CreditorAccount: &banksync.AccountIdentification{IBAN: "PL11111111111111111111111111"},
			},
			txType:   model.TypeExpense,
			wantIBAN: "PL11111111111111111111111111",
		},
		"income returns debtor IBAN": {
			ebTx: banksync.Transaction{
				DebtorAccount: &banksync.AccountIdentification{IBAN: "PL22222222222222222222222222"},
			},
			txType:   model.TypeIncome,
			wantIBAN: "PL22222222222222222222222222",
		},
		"expense with nil creditor account returns empty": {
			ebTx: banksync.Transaction{
				Creditor: banksync.PartyIdentification{Name: "No Account"},
			},
			txType:   model.TypeExpense,
			wantIBAN: "",
		},
		"income with nil debtor account returns empty": {
			ebTx: banksync.Transaction{
				Debtor: banksync.PartyIdentification{Name: "No Account"},
			},
			txType:   model.TypeIncome,
			wantIBAN: "",
		},
		"transfer type returns empty": {
			ebTx: banksync.Transaction{
				CreditorAccount: &banksync.AccountIdentification{IBAN: "PL11111111111111111111111111"},
				DebtorAccount:   &banksync.AccountIdentification{IBAN: "PL22222222222222222222222222"},
			},
			txType:   model.TypeTransfer,
			wantIBAN: "",
		},
		"expense with empty creditor IBAN returns empty": {
			ebTx: banksync.Transaction{
				CreditorAccount: &banksync.AccountIdentification{IBAN: ""},
			},
			txType:   model.TypeExpense,
			wantIBAN: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := svc.extractCounterpartyIBAN(tt.ebTx, tt.txType)
			if got != tt.wantIBAN {
				t.Errorf("want IBAN %q, got %q", tt.wantIBAN, got)
			}
		})
	}
}

// ── Tests: buildUserIBANSet ──

func TestBankSyncService_BuildUserIBANSet(t *testing.T) {
	tests := map[string]struct {
		connsByUser []model.BankConnection
		acctsByConn map[model.BankConnectionID][]model.BankAccount
		wantIBANs   map[string]bool
		wantErr     bool
	}{
		"collects IBANs from multiple connections": {
			connsByUser: []model.BankConnection{
				{ID: 1, UserID: 10},
				{ID: 2, UserID: 10},
			},
			acctsByConn: map[model.BankConnectionID][]model.BankAccount{
				1: {
					{ID: 100, IBAN: strPtr("PL11111111111111111111111111"), Currency: "PLN"},
					{ID: 101, IBAN: strPtr("PL22222222222222222222222222"), Currency: "EUR"},
				},
				2: {
					{ID: 200, IBAN: strPtr("PL33333333333333333333333333"), Currency: "PLN"},
				},
			},
			wantIBANs: map[string]bool{
				"PL11111111111111111111111111": true,
				"PL22222222222222222222222222": true,
				"PL33333333333333333333333333": true,
			},
		},
		"skips accounts with nil IBAN": {
			connsByUser: []model.BankConnection{
				{ID: 1, UserID: 10},
			},
			acctsByConn: map[model.BankConnectionID][]model.BankAccount{
				1: {
					{ID: 100, IBAN: nil, Currency: "PLN"},
					{ID: 101, IBAN: strPtr("PL11111111111111111111111111"), Currency: "PLN"},
				},
			},
			wantIBANs: map[string]bool{
				"PL11111111111111111111111111": true,
			},
		},
		"skips accounts with empty IBAN": {
			connsByUser: []model.BankConnection{
				{ID: 1, UserID: 10},
			},
			acctsByConn: map[model.BankConnectionID][]model.BankAccount{
				1: {
					{ID: 100, IBAN: strPtr(""), Currency: "PLN"},
				},
			},
			wantIBANs: map[string]bool{},
		},
		"no connections returns empty set": {
			connsByUser: nil,
			acctsByConn: map[model.BankConnectionID][]model.BankAccount{},
			wantIBANs:   map[string]bool{},
		},
		"connection with no accounts returns empty set": {
			connsByUser: []model.BankConnection{
				{ID: 1, UserID: 10},
			},
			acctsByConn: map[model.BankConnectionID][]model.BankAccount{
				1: {},
			},
			wantIBANs: map[string]bool{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			connStore := &mockBankConnStore{
				listByUserFunc: func(ctx context.Context, userID model.UserID) ([]model.BankConnection, error) {
					return tt.connsByUser, nil
				},
			}
			acctStore := &mockBankAcctStore{
				listByConnectionFunc: func(ctx context.Context, connID model.BankConnectionID) ([]model.BankAccount, error) {
					return tt.acctsByConn[connID], nil
				},
			}

			svc := newBankSyncService(connStore, acctStore, &mockBankSyncTxStore{}, &mockCategoryStore{}, &mockBankSyncClient{})

			got, err := svc.buildUserIBANSet(context.Background(), model.UserID(10))
			if (err != nil) != tt.wantErr {
				t.Fatalf("want err=%v, got %v", tt.wantErr, err)
			}
			if err != nil {
				return
			}

			if len(got) != len(tt.wantIBANs) {
				t.Errorf("want %d IBANs, got %d", len(tt.wantIBANs), len(got))
			}
			for iban := range tt.wantIBANs {
				if !got[iban] {
					t.Errorf("want IBAN %q in set, but missing", iban)
				}
			}
		})
	}
}

func TestBankSyncService_BuildUserIBANSet_Errors(t *testing.T) {
	t.Run("ListByUser error propagates", func(t *testing.T) {
		connStore := &mockBankConnStore{
			listByUserFunc: func(ctx context.Context, userID model.UserID) ([]model.BankConnection, error) {
				return nil, errors.New("db error")
			},
		}
		svc := newBankSyncService(connStore, &mockBankAcctStore{}, &mockBankSyncTxStore{}, &mockCategoryStore{}, &mockBankSyncClient{})

		_, err := svc.buildUserIBANSet(context.Background(), model.UserID(10))
		if err == nil {
			t.Fatal("want error, got nil")
		}
	})

	t.Run("ListByConnection error propagates", func(t *testing.T) {
		connStore := &mockBankConnStore{
			listByUserFunc: func(ctx context.Context, userID model.UserID) ([]model.BankConnection, error) {
				return []model.BankConnection{{ID: 1, UserID: 10}}, nil
			},
		}
		acctStore := &mockBankAcctStore{
			listByConnectionFunc: func(ctx context.Context, connID model.BankConnectionID) ([]model.BankAccount, error) {
				return nil, errors.New("db error")
			},
		}
		svc := newBankSyncService(connStore, acctStore, &mockBankSyncTxStore{}, &mockCategoryStore{}, &mockBankSyncClient{})

		_, err := svc.buildUserIBANSet(context.Background(), model.UserID(10))
		if err == nil {
			t.Fatal("want error, got nil")
		}
	})
}

// ── Tests: Orphaned Transaction Relinking ──

func TestBankSyncService_OrphanedTransactionRelinking(t *testing.T) {
	tests := map[string]struct {
		iban             *string
		currency         string
		workspaceIDs     []model.WorkspaceID
		relinkReturn     int64
		relinkErr        error
		wantRelinkCalls  int
		wantRelinkIBAN   string
		wantRelinkAcctID model.BankAccountID
		wantRelinkWSID   model.WorkspaceID
	}{
		"relinks orphaned transactions with matching IBAN and currency": {
			iban:             strPtr("PL11111111111111111111111111"),
			currency:         "PLN",
			workspaceIDs:     []model.WorkspaceID{1},
			relinkReturn:     3,
			wantRelinkCalls:  1,
			wantRelinkIBAN:   "PL11111111111111111111111111",
			wantRelinkAcctID: 100,
			wantRelinkWSID:   1,
		},
		"calls relink for each linked workspace": {
			iban:            strPtr("PL11111111111111111111111111"),
			currency:        "PLN",
			workspaceIDs:    []model.WorkspaceID{1, 2, 3},
			relinkReturn:    1,
			wantRelinkCalls: 3,
		},
		"skips relink when IBAN is nil": {
			iban:            nil,
			currency:        "PLN",
			workspaceIDs:    []model.WorkspaceID{1},
			wantRelinkCalls: 0,
		},
		"skips relink when IBAN is empty": {
			iban:            strPtr(""),
			currency:        "PLN",
			workspaceIDs:    []model.WorkspaceID{1},
			wantRelinkCalls: 0,
		},
		"relink error does not stop sync": {
			iban:            strPtr("PL11111111111111111111111111"),
			currency:        "PLN",
			workspaceIDs:    []model.WorkspaceID{1},
			relinkErr:       errors.New("db error"),
			wantRelinkCalls: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			relinkCalls := 0
			var lastRelinkIBAN string
			var lastRelinkAcctID model.BankAccountID
			var lastRelinkWSID model.WorkspaceID

			txStore := &mockBankSyncTxStore{
				relinkOrphanedTransactionsFunc: func(ctx context.Context, bankAccountID model.BankAccountID, iban, currency string, wsID model.WorkspaceID) (int64, error) {
					relinkCalls++
					lastRelinkIBAN = iban
					lastRelinkAcctID = bankAccountID
					lastRelinkWSID = wsID
					return tt.relinkReturn, tt.relinkErr
				},
				existsByFingerprintFunc: func(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error) {
					return true, nil // skip all transactions to keep test focused on relink
				},
			}

			acctStore := &mockBankAcctStore{
				listWorkspacesByBankAccountFunc: func(ctx context.Context, baID model.BankAccountID) ([]model.WorkspaceID, error) {
					return tt.workspaceIDs, nil
				},
			}

			catStore := &mockCategoryStore{
				GetBySlugFunc: func(ctx context.Context, wsID model.WorkspaceID, slug string) (model.Category, error) {
					return model.Category{ID: 99, Name: "Uncategorized"}, nil
				},
			}

			client := &mockBankSyncClient{
				getAccountTransactionsFunc: func(ctx context.Context, accountUID string, dateFrom, dateTo time.Time) ([]banksync.Transaction, error) {
					return []banksync.Transaction{}, nil
				},
			}

			svc := newBankSyncService(&mockBankConnStore{}, acctStore, txStore, catStore, client)

			conn := model.BankConnection{
				ID:            1,
				UserID:        10,
				AuthExpiresAt: time.Now().Add(24 * time.Hour),
			}
			acct := model.BankAccount{
				ID:         100,
				ExternalID: "ext-1",
				IBAN:       tt.iban,
				Currency:   tt.currency,
			}

			err := svc.SyncBankAccount(context.Background(), conn, acct, map[string]bool{})
			if err != nil {
				t.Fatalf("SyncBankAccount returned unexpected error: %v", err)
			}

			if relinkCalls != tt.wantRelinkCalls {
				t.Errorf("want %d relink calls, got %d", tt.wantRelinkCalls, relinkCalls)
			}

			if tt.wantRelinkCalls > 0 && tt.wantRelinkIBAN != "" {
				if lastRelinkIBAN != tt.wantRelinkIBAN {
					t.Errorf("want relink IBAN %q, got %q", tt.wantRelinkIBAN, lastRelinkIBAN)
				}
				if lastRelinkAcctID != tt.wantRelinkAcctID {
					t.Errorf("want relink account ID %d, got %d", tt.wantRelinkAcctID, lastRelinkAcctID)
				}
				if lastRelinkWSID != tt.wantRelinkWSID {
					t.Errorf("want relink workspace ID %d, got %d", tt.wantRelinkWSID, lastRelinkWSID)
				}
			}
		})
	}
}

// ── Tests: mapTransaction ──

func TestBankSyncService_MapTransaction(t *testing.T) {
	svc := newBankSyncService(
		&mockBankConnStore{},
		&mockBankAcctStore{},
		&mockBankSyncTxStore{},
		&mockCategoryStore{},
		&mockBankSyncClient{},
	)

	bankAcctID := model.BankAccountID(100)
	bankIBAN := strPtr("PL00000000000000000000000000")
	catID := model.CategoryID(99)
	wsID := model.WorkspaceID(1)
	userID := model.UserID(10)
	ownIBANs := map[string]bool{}

	tests := map[string]struct {
		ebTx       banksync.Transaction
		wantErr    bool
		wantType   model.TransactionType
		wantDesc   string
		wantAmount string
		wantCurr   string
		wantSource model.TransactionSource
	}{
		"DBIT indicator maps to expense": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-1",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "-50.00", Currency: "PLN"},
				CreditDebitIndicator: "DBIT",
				Creditor:             banksync.PartyIdentification{Name: "Grocery Store"},
			},
			wantType:   model.TypeExpense,
			wantDesc:   "Grocery Store",
			wantAmount: "-50",
			wantCurr:   "PLN",
			wantSource: model.SourceBank,
		},
		"CRDT indicator maps to income": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-2",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "2000.00", Currency: "PLN"},
				CreditDebitIndicator: "CRDT",
				Debtor:               banksync.PartyIdentification{Name: "Employer Inc."},
			},
			wantType:   model.TypeIncome,
			wantDesc:   "Employer Inc.",
			wantAmount: "2000",
			wantCurr:   "PLN",
			wantSource: model.SourceBank,
		},
		"positive amount without indicator maps to income": {
			ebTx: banksync.Transaction{
				TransactionID:     "tx-3",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "150.00", Currency: "EUR"},
				Debtor:            banksync.PartyIdentification{Name: "Friend"},
			},
			wantType:   model.TypeIncome,
			wantDesc:   "Friend",
			wantAmount: "150",
			wantCurr:   "EUR",
		},
		"negative amount without indicator maps to expense": {
			ebTx: banksync.Transaction{
				TransactionID:     "tx-4",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "-20.00", Currency: "PLN"},
				Creditor:          banksync.PartyIdentification{Name: "Coffee Shop"},
			},
			wantType: model.TypeExpense,
			wantDesc: "Coffee Shop",
		},
		"description falls back to debtor name": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-5",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "100.00", Currency: "PLN"},
				CreditDebitIndicator: "CRDT",
				Debtor:               banksync.PartyIdentification{Name: "Sender Name"},
			},
			wantDesc: "Sender Name",
			wantType: model.TypeIncome,
		},
		"description falls back to remittance info": {
			ebTx: banksync.Transaction{
				TransactionID:         "tx-6",
				BookingDate:           "2025-01-15",
				TransactionAmount:     banksync.Amount{Amount: "-5.00", Currency: "PLN"},
				CreditDebitIndicator:  "DBIT",
				RemittanceInformation: []string{"Payment for invoice #123"},
			},
			wantDesc: "Payment for invoice #123",
			wantType: model.TypeExpense,
		},
		"description falls back to default when all empty": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-7",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "-5.00", Currency: "PLN"},
				CreditDebitIndicator: "DBIT",
			},
			wantDesc: "Bank transaction",
			wantType: model.TypeExpense,
		},
		"empty currency defaults to PLN": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-8",
				BookingDate:          "2025-01-15",
				TransactionAmount:    banksync.Amount{Amount: "-10.00", Currency: ""},
				CreditDebitIndicator: "DBIT",
				Creditor:             banksync.PartyIdentification{Name: "Shop"},
			},
			wantCurr: "PLN",
			wantType: model.TypeExpense,
		},
		"invalid amount returns error": {
			ebTx: banksync.Transaction{
				TransactionID:     "tx-err-1",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "not-a-number", Currency: "PLN"},
			},
			wantErr: true,
		},
		"invalid date returns error": {
			ebTx: banksync.Transaction{
				TransactionID:     "tx-err-2",
				BookingDate:       "not-a-date",
				TransactionAmount: banksync.Amount{Amount: "10.00", Currency: "PLN"},
			},
			wantErr: true,
		},
		"fingerprint is set and source is bank": {
			ebTx: banksync.Transaction{
				TransactionID:        "tx-fp",
				BookingDate:          "2025-06-01",
				TransactionAmount:    banksync.Amount{Amount: "-100.00", Currency: "PLN"},
				CreditDebitIndicator: "DBIT",
				Creditor:             banksync.PartyIdentification{Name: "Test"},
			},
			wantSource: model.SourceBank,
			wantType:   model.TypeExpense,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tx, err := svc.mapTransaction(tt.ebTx, wsID, userID, bankAcctID, bankIBAN, catID, ownIBANs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("want err=%v, got %v", tt.wantErr, err)
			}
			if err != nil {
				return
			}

			if tx.Type != tt.wantType {
				t.Errorf("want type %q, got %q", tt.wantType, tx.Type)
			}
			if tt.wantDesc != "" && tx.Description != tt.wantDesc {
				t.Errorf("want description %q, got %q", tt.wantDesc, tx.Description)
			}
			if tt.wantAmount != "" {
				wantDec, _ := decimal.NewFromString(tt.wantAmount)
				if !tx.TotalAmount.Amount().Equal(wantDec) {
					t.Errorf("want amount %s, got %s", wantDec, tx.TotalAmount.Amount())
				}
			}
			if tt.wantCurr != "" && tx.TotalAmount.Currency() != tt.wantCurr {
				t.Errorf("want currency %q, got %q", tt.wantCurr, tx.TotalAmount.Currency())
			}
			if tt.wantSource != "" && tx.Source != tt.wantSource {
				t.Errorf("want source %q, got %q", tt.wantSource, tx.Source)
			}
			if tx.Fingerprint == nil {
				t.Error("want non-nil fingerprint")
			}
			if tx.BankAccountID == nil || *tx.BankAccountID != bankAcctID {
				t.Errorf("want bank account ID %d, got %v", bankAcctID, tx.BankAccountID)
			}
			if len(tx.Entries) != 1 {
				t.Fatalf("want 1 entry, got %d", len(tx.Entries))
			}
			if tx.Entries[0].CategoryID != catID {
				t.Errorf("want category ID %d, got %d", catID, tx.Entries[0].CategoryID)
			}
		})
	}
}

// ── Tests: buildFingerprint ──

func TestBuildFingerprint(t *testing.T) {
	tests := map[string]struct {
		tx1  banksync.Transaction
		tx2  banksync.Transaction
		same bool
	}{
		"same transaction_id produces same fingerprint": {
			tx1: banksync.Transaction{
				TransactionID:     "abc123",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "100.00", Currency: "PLN"},
			},
			tx2: banksync.Transaction{
				TransactionID:     "abc123",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "100.00", Currency: "PLN"},
			},
			same: true,
		},
		"different transaction_id produces different fingerprint": {
			tx1: banksync.Transaction{
				TransactionID:     "abc123",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "100.00", Currency: "PLN"},
			},
			tx2: banksync.Transaction{
				TransactionID:     "def456",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "100.00", Currency: "PLN"},
			},
			same: false,
		},
		"no transaction_id uses remittance info": {
			tx1: banksync.Transaction{
				BookingDate:           "2025-01-15",
				TransactionAmount:     banksync.Amount{Amount: "50.00", Currency: "PLN"},
				RemittanceInformation: []string{"Invoice #001"},
			},
			tx2: banksync.Transaction{
				BookingDate:           "2025-01-15",
				TransactionAmount:     banksync.Amount{Amount: "50.00", Currency: "PLN"},
				RemittanceInformation: []string{"Invoice #001"},
			},
			same: true,
		},
		"different remittance info produces different fingerprint": {
			tx1: banksync.Transaction{
				BookingDate:           "2025-01-15",
				TransactionAmount:     banksync.Amount{Amount: "50.00", Currency: "PLN"},
				RemittanceInformation: []string{"Invoice #001"},
			},
			tx2: banksync.Transaction{
				BookingDate:           "2025-01-15",
				TransactionAmount:     banksync.Amount{Amount: "50.00", Currency: "PLN"},
				RemittanceInformation: []string{"Invoice #002"},
			},
			same: false,
		},
		"different amount produces different fingerprint": {
			tx1: banksync.Transaction{
				TransactionID:     "abc",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "100.00", Currency: "PLN"},
			},
			tx2: banksync.Transaction{
				TransactionID:     "abc",
				BookingDate:       "2025-01-15",
				TransactionAmount: banksync.Amount{Amount: "200.00", Currency: "PLN"},
			},
			same: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fp1 := buildFingerprint(tt.tx1)
			fp2 := buildFingerprint(tt.tx2)

			if fp1 == "" {
				t.Fatal("want non-empty fingerprint")
			}

			if tt.same && fp1 != fp2 {
				t.Errorf("want same fingerprint, got %q and %q", fp1, fp2)
			}
			if !tt.same && fp1 == fp2 {
				t.Errorf("want different fingerprints, got same: %q", fp1)
			}
		})
	}
}

// ── Tests: SyncBankAccount full flow ──

func TestBankSyncService_SyncBankAccount_FullFlow(t *testing.T) {
	t.Run("creates new transactions and skips duplicates", func(t *testing.T) {
		var createdTxs []model.Transaction
		existingFingerprints := map[string]bool{}

		txStore := &mockBankSyncTxStore{
			createFunc: func(ctx context.Context, tx model.Transaction) (model.Transaction, error) {
				createdTxs = append(createdTxs, tx)
				return tx, nil
			},
			existsByFingerprintFunc: func(ctx context.Context, wsID model.WorkspaceID, fingerprint string) (bool, error) {
				return existingFingerprints[fingerprint], nil
			},
		}

		acctStore := &mockBankAcctStore{
			listWorkspacesByBankAccountFunc: func(ctx context.Context, baID model.BankAccountID) ([]model.WorkspaceID, error) {
				return []model.WorkspaceID{1}, nil
			},
		}

		catStore := &mockCategoryStore{
			GetByNameFunc: func(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
				return model.Category{ID: 99, Name: "Uncategorized"}, nil
			},
		}

		// First transaction is new, second already has its fingerprint
		tx1 := banksync.Transaction{
			TransactionID:        "new-tx",
			BookingDate:          "2025-01-15",
			TransactionAmount:    banksync.Amount{Amount: "-50.00", Currency: "PLN"},
			CreditDebitIndicator: "DBIT",
			Creditor:             banksync.PartyIdentification{Name: "Store"},
		}
		tx2 := banksync.Transaction{
			TransactionID:        "existing-tx",
			BookingDate:          "2025-01-14",
			TransactionAmount:    banksync.Amount{Amount: "-30.00", Currency: "PLN"},
			CreditDebitIndicator: "DBIT",
			Creditor:             banksync.PartyIdentification{Name: "Other Store"},
		}
		existingFingerprints[buildFingerprint(tx2)] = true

		client := &mockBankSyncClient{
			getAccountTransactionsFunc: func(ctx context.Context, accountUID string, dateFrom, dateTo time.Time) ([]banksync.Transaction, error) {
				return []banksync.Transaction{tx1, tx2}, nil
			},
		}

		svc := newBankSyncService(&mockBankConnStore{}, acctStore, txStore, catStore, client)

		conn := model.BankConnection{ID: 1, UserID: 10, AuthExpiresAt: time.Now().Add(24 * time.Hour)}
		iban := strPtr("PL00000000000000000000000000")
		acct := model.BankAccount{ID: 100, ExternalID: "ext-1", IBAN: iban, Currency: "PLN"}

		err := svc.SyncBankAccount(context.Background(), conn, acct, map[string]bool{})
		if err != nil {
			t.Fatalf("want nil error, got %v", err)
		}

		if len(createdTxs) != 1 {
			t.Fatalf("want 1 created transaction, got %d", len(createdTxs))
		}
		if createdTxs[0].Description != "Store" {
			t.Errorf("want description %q, got %q", "Store", createdTxs[0].Description)
		}
	})

	t.Run("self-transfer detected during full sync", func(t *testing.T) {
		var createdTxs []model.Transaction

		txStore := &mockBankSyncTxStore{
			createFunc: func(ctx context.Context, tx model.Transaction) (model.Transaction, error) {
				createdTxs = append(createdTxs, tx)
				return tx, nil
			},
		}

		acctStore := &mockBankAcctStore{
			listWorkspacesByBankAccountFunc: func(ctx context.Context, baID model.BankAccountID) ([]model.WorkspaceID, error) {
				return []model.WorkspaceID{1}, nil
			},
		}

		catStore := &mockCategoryStore{
			GetByNameFunc: func(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
				return model.Category{ID: 99, Name: "Uncategorized"}, nil
			},
		}

		// Transaction where creditor IBAN matches one of user's own IBANs
		client := &mockBankSyncClient{
			getAccountTransactionsFunc: func(ctx context.Context, accountUID string, dateFrom, dateTo time.Time) ([]banksync.Transaction, error) {
				return []banksync.Transaction{
					{
						TransactionID:        "transfer-tx",
						BookingDate:          "2025-01-15",
						TransactionAmount:    banksync.Amount{Amount: "-500.00", Currency: "PLN"},
						CreditDebitIndicator: "DBIT",
						Creditor:             banksync.PartyIdentification{Name: "My Savings Account"},
						CreditorAccount:      &banksync.AccountIdentification{IBAN: "PL11111111111111111111111111"},
					},
				}, nil
			},
		}

		svc := newBankSyncService(&mockBankConnStore{}, acctStore, txStore, catStore, client)

		conn := model.BankConnection{ID: 1, UserID: 10, AuthExpiresAt: time.Now().Add(24 * time.Hour)}
		iban := strPtr("PL00000000000000000000000000")
		acct := model.BankAccount{ID: 100, ExternalID: "ext-1", IBAN: iban, Currency: "PLN"}

		ownIBANs := map[string]bool{
			"PL11111111111111111111111111": true,
			"PL00000000000000000000000000": true,
		}

		err := svc.SyncBankAccount(context.Background(), conn, acct, ownIBANs)
		if err != nil {
			t.Fatalf("want nil error, got %v", err)
		}

		if len(createdTxs) != 1 {
			t.Fatalf("want 1 created transaction, got %d", len(createdTxs))
		}
		if createdTxs[0].Type != model.TypeTransfer {
			t.Errorf("want type %q, got %q", model.TypeTransfer, createdTxs[0].Type)
		}
		if createdTxs[0].CounterpartyIBAN == nil {
			t.Error("want non-nil CounterpartyIBAN on created transaction")
		} else if *createdTxs[0].CounterpartyIBAN != "PL11111111111111111111111111" {
			t.Errorf("want CounterpartyIBAN %q, got %q", "PL11111111111111111111111111", *createdTxs[0].CounterpartyIBAN)
		}
	})

	t.Run("relinks orphaned then syncs new transactions", func(t *testing.T) {
		relinkCalled := false
		var createdTxs []model.Transaction

		txStore := &mockBankSyncTxStore{
			relinkOrphanedTransactionsFunc: func(ctx context.Context, bankAccountID model.BankAccountID, iban, currency string, wsID model.WorkspaceID) (int64, error) {
				relinkCalled = true
				if iban != "PL11111111111111111111111111" {
					t.Errorf("want relink IBAN %q, got %q", "PL11111111111111111111111111", iban)
				}
				if currency != "PLN" {
					t.Errorf("want relink currency %q, got %q", "PLN", currency)
				}
				return 2, nil
			},
			createFunc: func(ctx context.Context, tx model.Transaction) (model.Transaction, error) {
				createdTxs = append(createdTxs, tx)
				return tx, nil
			},
		}

		acctStore := &mockBankAcctStore{
			listWorkspacesByBankAccountFunc: func(ctx context.Context, baID model.BankAccountID) ([]model.WorkspaceID, error) {
				return []model.WorkspaceID{1}, nil
			},
		}

		catStore := &mockCategoryStore{
			GetByNameFunc: func(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
				return model.Category{ID: 99, Name: "Uncategorized"}, nil
			},
		}

		client := &mockBankSyncClient{
			getAccountTransactionsFunc: func(ctx context.Context, accountUID string, dateFrom, dateTo time.Time) ([]banksync.Transaction, error) {
				return []banksync.Transaction{
					{
						TransactionID:        "new-after-reconnect",
						BookingDate:          "2025-02-01",
						TransactionAmount:    banksync.Amount{Amount: "-25.00", Currency: "PLN"},
						CreditDebitIndicator: "DBIT",
						Creditor:             banksync.PartyIdentification{Name: "Cafe"},
					},
				}, nil
			},
		}

		svc := newBankSyncService(&mockBankConnStore{}, acctStore, txStore, catStore, client)

		conn := model.BankConnection{ID: 1, UserID: 10, AuthExpiresAt: time.Now().Add(24 * time.Hour)}
		iban := strPtr("PL11111111111111111111111111")
		acct := model.BankAccount{ID: 100, ExternalID: "ext-1", IBAN: iban, Currency: "PLN"}

		err := svc.SyncBankAccount(context.Background(), conn, acct, map[string]bool{})
		if err != nil {
			t.Fatalf("want nil error, got %v", err)
		}

		if !relinkCalled {
			t.Error("want relink to be called, but it was not")
		}
		if len(createdTxs) != 1 {
			t.Fatalf("want 1 created transaction, got %d", len(createdTxs))
		}
		if createdTxs[0].Description != "Cafe" {
			t.Errorf("want description %q, got %q", "Cafe", createdTxs[0].Description)
		}
	})

	t.Run("no workspaces linked means no relink and no creates", func(t *testing.T) {
		relinkCalls := 0
		createCalls := 0

		txStore := &mockBankSyncTxStore{
			relinkOrphanedTransactionsFunc: func(ctx context.Context, bankAccountID model.BankAccountID, iban, currency string, wsID model.WorkspaceID) (int64, error) {
				relinkCalls++
				return 0, nil
			},
			createFunc: func(ctx context.Context, tx model.Transaction) (model.Transaction, error) {
				createCalls++
				return tx, nil
			},
		}

		acctStore := &mockBankAcctStore{
			listWorkspacesByBankAccountFunc: func(ctx context.Context, baID model.BankAccountID) ([]model.WorkspaceID, error) {
				return nil, nil // no linked workspaces
			},
		}

		client := &mockBankSyncClient{
			getAccountTransactionsFunc: func(ctx context.Context, accountUID string, dateFrom, dateTo time.Time) ([]banksync.Transaction, error) {
				return []banksync.Transaction{
					{
						TransactionID:     "tx-1",
						BookingDate:       "2025-01-15",
						TransactionAmount: banksync.Amount{Amount: "-10.00", Currency: "PLN"},
					},
				}, nil
			},
		}

		svc := newBankSyncService(&mockBankConnStore{}, acctStore, txStore, &mockCategoryStore{}, client)

		conn := model.BankConnection{ID: 1, UserID: 10, AuthExpiresAt: time.Now().Add(24 * time.Hour)}
		iban := strPtr("PL11111111111111111111111111")
		acct := model.BankAccount{ID: 100, ExternalID: "ext-1", IBAN: iban, Currency: "PLN"}

		err := svc.SyncBankAccount(context.Background(), conn, acct, map[string]bool{})
		if err != nil {
			t.Fatalf("want nil error, got %v", err)
		}

		if relinkCalls != 0 {
			t.Errorf("want 0 relink calls, got %d", relinkCalls)
		}
		if createCalls != 0 {
			t.Errorf("want 0 create calls, got %d", createCalls)
		}
	})
}

// ── Tests: calcDateFrom ──

func TestBankSyncService_CalcDateFrom(t *testing.T) {
	svc := newBankSyncService(
		&mockBankConnStore{},
		&mockBankAcctStore{},
		&mockBankSyncTxStore{},
		&mockCategoryStore{},
		&mockBankSyncClient{},
	)

	t.Run("nil lastSynced returns 3 months ago", func(t *testing.T) {
		got := svc.calcDateFrom(nil)
		expected := time.Now().AddDate(0, -3, 0)
		diff := got.Sub(expected)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("want ~%v, got %v", expected, got)
		}
	})

	t.Run("non-nil lastSynced returns 3 days before", func(t *testing.T) {
		lastSynced := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		got := svc.calcDateFrom(&lastSynced)
		expected := time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC)
		if !got.Equal(expected) {
			t.Errorf("want %v, got %v", expected, got)
		}
	})
}
