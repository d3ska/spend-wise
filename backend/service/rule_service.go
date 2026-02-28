package service

import (
	"context"
	"fmt"
	"regexp"

	"backend/model"

	"github.com/shopspring/decimal"
)

// maxPatternLength is the maximum allowed length for a regex match pattern.
const maxPatternLength = 500

// RuleStoreIface defines the rule store methods used by RuleService.
type RuleStoreIface interface {
	GetByID(ctx context.Context, id model.RuleID) (model.CategorizationRule, error)
	Create(ctx context.Context, r model.CategorizationRule) (model.CategorizationRule, error)
	Update(ctx context.Context, id model.RuleID, pattern string, targetCatID model.CategoryID, priority int, amountMin, amountMax *decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (model.CategorizationRule, error)
	Delete(ctx context.Context, id model.RuleID) error
	ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]model.CategorizationRule, error)
	ToggleEnabled(ctx context.Context, id model.RuleID, enabled bool) (model.CategorizationRule, error)
	ResolveForTransaction(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, description string, amount decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (*model.CategoryID, error)
}

// CreateRuleInput holds input for creating a categorization rule.
type CreateRuleInput struct {
	MatchPattern     string
	TargetCategoryID model.CategoryID
	Priority         int
	AmountMin        *decimal.Decimal
	AmountMax        *decimal.Decimal
	BankAccountID    *model.BankAccountID
	CounterpartyIBAN *string
}

// UpdateRuleInput holds input for updating a categorization rule.
type UpdateRuleInput struct {
	MatchPattern     string
	TargetCategoryID model.CategoryID
	Priority         int
	AmountMin        *decimal.Decimal
	AmountMax        *decimal.Decimal
	BankAccountID    *model.BankAccountID
	CounterpartyIBAN *string
}

// RuleService handles categorization rule operations with permission checks.
type RuleService struct {
	ruleStore RuleStoreIface
	wsStore   WorkspaceStoreIface
}

// NewRuleService creates a new RuleService.
func NewRuleService(ruleStore RuleStoreIface, wsStore WorkspaceStoreIface) *RuleService {
	return &RuleService{
		ruleStore: ruleStore,
		wsStore:   wsStore,
	}
}

// validatePattern checks that a regex pattern is valid and within length limits.
func validatePattern(pattern string) error {
	if len(pattern) > maxPatternLength {
		return fmt.Errorf("%w: must be at most %d characters", model.ErrRulePatternTooLong, maxPatternLength)
	}
	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("%w: %v", model.ErrRuleInvalidPattern, err)
	}
	return nil
}

// validateAmountRange checks that amount bounds are non-negative and min <= max.
func validateAmountRange(min, max *decimal.Decimal) error {
	if min != nil && min.IsNegative() {
		return fmt.Errorf("%w: min must be non-negative", model.ErrRuleInvalidAmountRange)
	}
	if max != nil && max.IsNegative() {
		return fmt.Errorf("%w: max must be non-negative", model.ErrRuleInvalidAmountRange)
	}
	if min != nil && max != nil && min.GreaterThan(*max) {
		return fmt.Errorf("%w: min must be less than or equal to max", model.ErrRuleInvalidAmountRange)
	}
	return nil
}

// requireRuleInWorkspace verifies that the rule belongs to the given workspace.
func (s *RuleService) requireRuleInWorkspace(ctx context.Context, ruleID model.RuleID, wsID model.WorkspaceID) error {
	rule, err := s.ruleStore.GetByID(ctx, ruleID)
	if err != nil {
		return err
	}
	if rule.WorkspaceID == nil || *rule.WorkspaceID != wsID {
		return model.ErrRuleNotFound
	}
	return nil
}

// Create creates a workspace-scoped rule. Requires editor+ permission.
func (s *RuleService) Create(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID, input CreateRuleInput) (model.CategorizationRule, error) {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return model.CategorizationRule{}, err
	}
	if member.Role.Level() < model.RoleEditor.Level() {
		return model.CategorizationRule{}, model.ErrInsufficientPermission
	}

	if err := validatePattern(input.MatchPattern); err != nil {
		return model.CategorizationRule{}, err
	}

	if err := validateAmountRange(input.AmountMin, input.AmountMax); err != nil {
		return model.CategorizationRule{}, err
	}

	return s.ruleStore.Create(ctx, model.CategorizationRule{
		Scope:            model.ScopeWorkspace,
		WorkspaceID:      &wsID,
		MatchPattern:     input.MatchPattern,
		TargetCategoryID: input.TargetCategoryID,
		Priority:         input.Priority,
		Enabled:          true,
		AmountMin:        input.AmountMin,
		AmountMax:        input.AmountMax,
		BankAccountID:    input.BankAccountID,
		CounterpartyIBAN: input.CounterpartyIBAN,
	})
}

// Update modifies a rule. Requires editor+ permission.
func (s *RuleService) Update(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID, ruleID model.RuleID, input UpdateRuleInput) (model.CategorizationRule, error) {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return model.CategorizationRule{}, err
	}
	if member.Role.Level() < model.RoleEditor.Level() {
		return model.CategorizationRule{}, model.ErrInsufficientPermission
	}

	if err := s.requireRuleInWorkspace(ctx, ruleID, wsID); err != nil {
		return model.CategorizationRule{}, err
	}

	if err := validatePattern(input.MatchPattern); err != nil {
		return model.CategorizationRule{}, err
	}

	if err := validateAmountRange(input.AmountMin, input.AmountMax); err != nil {
		return model.CategorizationRule{}, err
	}

	return s.ruleStore.Update(ctx, ruleID, input.MatchPattern, input.TargetCategoryID, input.Priority, input.AmountMin, input.AmountMax, input.BankAccountID, input.CounterpartyIBAN)
}

// Delete removes a rule. Requires editor+ permission.
func (s *RuleService) Delete(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID, ruleID model.RuleID) error {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return err
	}
	if member.Role.Level() < model.RoleEditor.Level() {
		return model.ErrInsufficientPermission
	}

	if err := s.requireRuleInWorkspace(ctx, ruleID, wsID); err != nil {
		return err
	}

	return s.ruleStore.Delete(ctx, ruleID)
}

// Toggle enables or disables a rule. Requires editor+ permission.
func (s *RuleService) Toggle(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID, ruleID model.RuleID, enabled bool) (model.CategorizationRule, error) {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return model.CategorizationRule{}, err
	}
	if member.Role.Level() < model.RoleEditor.Level() {
		return model.CategorizationRule{}, model.ErrInsufficientPermission
	}

	if err := s.requireRuleInWorkspace(ctx, ruleID, wsID); err != nil {
		return model.CategorizationRule{}, err
	}

	return s.ruleStore.ToggleEnabled(ctx, ruleID, enabled)
}

// ListByWorkspace returns all rules for a workspace. Requires viewer+ permission.
func (s *RuleService) ListByWorkspace(ctx context.Context, callerID model.UserID, wsID model.WorkspaceID) ([]model.CategorizationRule, error) {
	member, err := s.wsStore.GetMember(ctx, wsID, callerID)
	if err != nil {
		return nil, err
	}
	if member.Role.Level() < model.RoleViewer.Level() {
		return nil, model.ErrInsufficientPermission
	}

	return s.ruleStore.ListByWorkspace(ctx, wsID)
}

// ResolveCategory finds the best matching category for a transaction description, amount,
// and optionally a bank account and counterparty IBAN. Returns nil if no rule matches.
func (s *RuleService) ResolveCategory(ctx context.Context, wsID model.WorkspaceID, userID model.UserID, description string, amount decimal.Decimal, bankAccountID *model.BankAccountID, counterpartyIBAN *string) (*model.CategoryID, error) {
	return s.ruleStore.ResolveForTransaction(ctx, wsID, userID, description, amount, bankAccountID, counterpartyIBAN)
}
