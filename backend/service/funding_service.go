package service

import (
	"context"
	"fmt"

	"backend/model"

	"github.com/shopspring/decimal"
)

// FundingStoreIface defines the funding store methods used by FundingService.
type FundingStoreIface interface {
	Upsert(ctx context.Context, f model.Funding) (model.Funding, error)
	ListByWorkspace(ctx context.Context, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error)
	Delete(ctx context.Context, id model.FundingID, wsID model.WorkspaceID) error
}

// RecordFundingInput holds input for recording a funding/budget entry.
type RecordFundingInput struct {
	UserID     model.UserID
	CategoryID *model.CategoryID // nil = overall budget
	YearMonth  string
	Amount     model.Money
}

// FundingService handles funding operations with permission checks.
type FundingService struct {
	fundingStore FundingStoreIface
	wsStore      WorkspaceStoreIface
}

// NewFundingService creates a new FundingService.
func NewFundingService(fundingStore FundingStoreIface, wsStore WorkspaceStoreIface) *FundingService {
	return &FundingService{
		fundingStore: fundingStore,
		wsStore:      wsStore,
	}
}

// Record upserts a funding/budget entry. Requires editor+ permission.
func (s *FundingService) Record(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID, input RecordFundingInput) (model.Funding, error) {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return model.Funding{}, err
	}
	if member.Role.Level() < model.RoleEditor.Level() {
		return model.Funding{}, model.ErrInsufficientPermission
	}

	// Validate that category budgets don't exceed overall budget.
	if err := s.validateBudgetLimit(ctx, wsID, input); err != nil {
		return model.Funding{}, err
	}

	return s.fundingStore.Upsert(ctx, model.Funding{
		WorkspaceID: wsID,
		UserID:      input.UserID,
		CategoryID:  input.CategoryID,
		YearMonth:   input.YearMonth,
		Amount:      input.Amount,
	})
}

// validateBudgetLimit ensures category budgets don't exceed the overall budget for the month.
func (s *FundingService) validateBudgetLimit(ctx context.Context, wsID model.WorkspaceID, input RecordFundingInput) error {
	existing, err := s.fundingStore.ListByWorkspace(ctx, wsID, input.YearMonth, input.YearMonth)
	if err != nil {
		return fmt.Errorf("fetching existing budgets: %w", err)
	}

	var overallAmount decimal.Decimal
	hasOverall := false
	categorySum := decimal.Zero

	for _, f := range existing {
		if f.CategoryID == nil {
			overallAmount = f.Amount.Amount()
			hasOverall = true
		} else {
			// Skip the category being updated — we'll use the new amount instead.
			if input.CategoryID != nil && *f.CategoryID == *input.CategoryID {
				continue
			}
			categorySum = categorySum.Add(f.Amount.Amount())
		}
	}

	if input.CategoryID == nil {
		// Setting overall budget: check that new overall >= sum of existing category budgets.
		if categorySum.GreaterThan(input.Amount.Amount()) {
			return model.ErrCategoryBudgetsExceed
		}
	} else {
		// Setting a category budget: check sum of categories + this one <= overall.
		if !hasOverall {
			return nil // No overall budget set — no constraint to enforce.
		}
		newCategorySum := categorySum.Add(input.Amount.Amount())
		if newCategorySum.GreaterThan(overallAmount) {
			return model.ErrCategoryBudgetsExceed
		}
	}

	return nil
}

// ListByRange returns fundings for a workspace within a year_month range. Requires viewer+.
func (s *FundingService) ListByRange(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID, fromMonth, toMonth string) ([]model.Funding, error) {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return nil, err
	}
	if member.Role.Level() < model.RoleViewer.Level() {
		return nil, model.ErrInsufficientPermission
	}

	return s.fundingStore.ListByWorkspace(ctx, wsID, fromMonth, toMonth)
}

// Delete removes a funding/budget entry by ID. Requires editor+ permission.
func (s *FundingService) Delete(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID, fundingID model.FundingID) error {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return err
	}
	if member.Role.Level() < model.RoleEditor.Level() {
		return model.ErrInsufficientPermission
	}

	return s.fundingStore.Delete(ctx, fundingID, wsID)
}
