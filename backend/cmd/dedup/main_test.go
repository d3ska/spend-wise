package main

import (
	"context"
	"errors"
	"testing"
)

// ── Mock ──

type mockDedupFinder struct {
	findFunc   func(ctx context.Context) ([]duplicate, error)
	deleteFunc func(ctx context.Context, ids []int64) (int64, error)

	deletedIDs []int64
}

func (m *mockDedupFinder) FindDuplicates(ctx context.Context) ([]duplicate, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx)
	}
	return nil, nil
}

func (m *mockDedupFinder) DeleteByIDs(ctx context.Context, ids []int64) (int64, error) {
	m.deletedIDs = ids
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, ids)
	}
	return int64(len(ids)), nil
}

// ── Tests ──

func TestDedup(t *testing.T) {
	tests := map[string]struct {
		finder      *mockDedupFinder
		apply       bool
		wantErr     bool
		wantDeleted []int64
	}{
		"no duplicates found": {
			finder: &mockDedupFinder{
				findFunc: func(ctx context.Context) ([]duplicate, error) {
					return nil, nil
				},
			},
			apply:       true,
			wantDeleted: nil,
		},
		"dry run does not delete": {
			finder: &mockDedupFinder{
				findFunc: func(ctx context.Context) ([]duplicate, error) {
					return []duplicate{
						{KeepID: 1, DeleteID: 2, WorkspaceID: 10, Description: "AUTOPAY S.A.", Amount: "500.34", Currency: "PLN", KeepDate: "2026-02-20", DeleteDate: "2026-02-21"},
					}, nil
				},
			},
			apply:       false,
			wantDeleted: nil,
		},
		"apply deletes duplicates": {
			finder: &mockDedupFinder{
				findFunc: func(ctx context.Context) ([]duplicate, error) {
					return []duplicate{
						{KeepID: 1, DeleteID: 2, WorkspaceID: 10, Description: "AUTOPAY S.A.", Amount: "500.34", Currency: "PLN", KeepDate: "2026-02-20", DeleteDate: "2026-02-21"},
					}, nil
				},
			},
			apply:       true,
			wantDeleted: []int64{2},
		},
		"apply deletes multiple duplicates": {
			finder: &mockDedupFinder{
				findFunc: func(ctx context.Context) ([]duplicate, error) {
					return []duplicate{
						{KeepID: 1, DeleteID: 2, WorkspaceID: 10, Description: "AUTOPAY S.A.", Amount: "500.34", Currency: "PLN", KeepDate: "2026-02-20", DeleteDate: "2026-02-21"},
						{KeepID: 5, DeleteID: 6, WorkspaceID: 10, Description: "Grocery Store", Amount: "123.00", Currency: "PLN", KeepDate: "2026-02-18", DeleteDate: "2026-02-19"},
						{KeepID: 10, DeleteID: 11, WorkspaceID: 20, Description: "Netflix", Amount: "49.99", Currency: "PLN", KeepDate: "2026-02-15", DeleteDate: "2026-02-16"},
					}, nil
				},
			},
			apply:       true,
			wantDeleted: []int64{2, 6, 11},
		},
		"find error propagates": {
			finder: &mockDedupFinder{
				findFunc: func(ctx context.Context) ([]duplicate, error) {
					return nil, errors.New("db connection failed")
				},
			},
			apply:   false,
			wantErr: true,
		},
		"delete error propagates": {
			finder: &mockDedupFinder{
				findFunc: func(ctx context.Context) ([]duplicate, error) {
					return []duplicate{
						{KeepID: 1, DeleteID: 2, WorkspaceID: 10, Description: "Test", Amount: "10.00", Currency: "PLN", KeepDate: "2026-01-01", DeleteDate: "2026-01-02"},
					}, nil
				},
				deleteFunc: func(ctx context.Context, ids []int64) (int64, error) {
					return 0, errors.New("delete failed")
				},
			},
			apply:   true,
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := dedup(context.Background(), tt.finder, tt.apply)

			if (err != nil) != tt.wantErr {
				t.Fatalf("want err=%v, got %v", tt.wantErr, err)
			}
			if err != nil {
				return
			}

			if tt.wantDeleted == nil {
				if tt.finder.deletedIDs != nil {
					t.Errorf("want no delete call, got IDs %v", tt.finder.deletedIDs)
				}
			} else {
				if len(tt.finder.deletedIDs) != len(tt.wantDeleted) {
					t.Fatalf("want %d deleted IDs, got %d", len(tt.wantDeleted), len(tt.finder.deletedIDs))
				}
				for i, id := range tt.wantDeleted {
					if tt.finder.deletedIDs[i] != id {
						t.Errorf("want deleted ID[%d]=%d, got %d", i, id, tt.finder.deletedIDs[i])
					}
				}
			}
		})
	}
}
