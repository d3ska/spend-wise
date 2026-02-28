package service

import (
	"context"
	"errors"
	"testing"

	"backend/model"
	"backend/store"
)

// ── Mock: LinkingTransactionStoreIface ──

type mockLinkingTxStore struct {
	reclassifyFunc func(ctx context.Context, wsID model.WorkspaceID, ibans []string) (int64, error)
}

func (m *mockLinkingTxStore) ReclassifyTransfersByCounterpartyIBAN(ctx context.Context, wsID model.WorkspaceID, ibans []string) (int64, error) {
	if m.reclassifyFunc != nil {
		return m.reclassifyFunc(ctx, wsID, ibans)
	}
	return 0, nil
}

// ── Tests: LinkToWorkspace reclassification ──

func TestBankAccountLinkingService_LinkToWorkspace_Reclassify(t *testing.T) {
	callerID := model.UserID(10)
	wsID := model.WorkspaceID(1)

	ownerMember := model.WorkspaceMember{
		WorkspaceID: wsID,
		UserID:      callerID,
		Role:        model.RoleOwner,
	}

	tests := map[string]struct {
		acctIBAN            *string
		linkErr             error
		reclassifyReturn    int64
		reclassifyErr       error
		wantReclassifyCalls int
		wantReclassifyIBAN  string
		wantErr             bool
	}{
		"reclassifies transactions after linking account with IBAN": {
			acctIBAN:            strPtr("PL11111111111111111111111111"),
			reclassifyReturn:    3,
			wantReclassifyCalls: 1,
			wantReclassifyIBAN:  "PL11111111111111111111111111",
		},
		"skips reclassification when IBAN is nil": {
			acctIBAN:            nil,
			wantReclassifyCalls: 0,
		},
		"skips reclassification when IBAN is empty": {
			acctIBAN:            strPtr(""),
			wantReclassifyCalls: 0,
		},
		"reclassify error does not fail linking": {
			acctIBAN:            strPtr("PL11111111111111111111111111"),
			reclassifyErr:       errors.New("db error"),
			wantReclassifyCalls: 1,
		},
		"link error prevents reclassification": {
			acctIBAN:            strPtr("PL11111111111111111111111111"),
			linkErr:             errors.New("link failed"),
			wantReclassifyCalls: 0,
			wantErr:             true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reclassifyCalls := 0
			var lastIBAN string

			txStore := &mockLinkingTxStore{
				reclassifyFunc: func(ctx context.Context, wsID model.WorkspaceID, ibans []string) (int64, error) {
					reclassifyCalls++
					if len(ibans) > 0 {
						lastIBAN = ibans[0]
					}
					return tt.reclassifyReturn, tt.reclassifyErr
				},
			}

			acctStore := &mockBankAcctStore{
				getByIDFunc: func(ctx context.Context, id model.BankAccountID) (model.BankAccount, error) {
					return model.BankAccount{
						ID:               id,
						BankConnectionID: 1,
						IBAN:             tt.acctIBAN,
						Currency:         "PLN",
					}, nil
				},
				linkToWorkspaceFunc: func(ctx context.Context, wsID model.WorkspaceID, baID model.BankAccountID) error {
					return tt.linkErr
				},
				listByWorkspaceFunc: func(ctx context.Context, wsID model.WorkspaceID) ([]store.BankAccountWithInstitution, error) {
					return nil, nil
				},
				listByUserFunc: func(ctx context.Context, userID model.UserID) ([]store.BankAccountWithInstitution, error) {
					return nil, nil
				},
			}

			connStore := &mockBankConnStore{
				getByIDFunc: func(ctx context.Context, id model.BankConnectionID) (model.BankConnection, error) {
					return model.BankConnection{
						ID:     id,
						UserID: callerID,
					}, nil
				},
			}

			wsStore := &mockWorkspaceStore{
				GetMemberFunc: func(ctx context.Context, ws model.WorkspaceID, uid model.UserID) (model.WorkspaceMember, error) {
					return ownerMember, nil
				},
			}

			svc := NewBankAccountLinkingService(acctStore, connStore, wsStore, txStore)

			err := svc.LinkToWorkspace(context.Background(), callerID, wsID, model.BankAccountID(100))
			if (err != nil) != tt.wantErr {
				t.Fatalf("want err=%v, got %v", tt.wantErr, err)
			}

			if reclassifyCalls != tt.wantReclassifyCalls {
				t.Errorf("want %d reclassify calls, got %d", tt.wantReclassifyCalls, reclassifyCalls)
			}

			if tt.wantReclassifyIBAN != "" && lastIBAN != tt.wantReclassifyIBAN {
				t.Errorf("want reclassify IBAN %q, got %q", tt.wantReclassifyIBAN, lastIBAN)
			}
		})
	}
}
