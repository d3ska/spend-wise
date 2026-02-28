package model

// Handler-level error codes.
const (
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeInvalidRequest = "INVALID_REQUEST"
	CodeInternalError  = "INTERNAL_ERROR"
)

// Auth errors.
var (
	ErrInvalidCredentials = NewAppError("INVALID_CREDENTIALS", "invalid credentials")
	ErrEmailAlreadyExists = NewAppError("EMAIL_ALREADY_EXISTS", "email already exists")
	ErrUserNotFound       = NewAppError("USER_NOT_FOUND", "user not found")
)

// Workspace errors.
var (
	ErrWorkspaceNotFound            = NewAppError("WORKSPACE_NOT_FOUND", "workspace not found")
	ErrWorkspaceNameRequired        = NewAppError("WORKSPACE_NAME_REQUIRED", "workspace name is required")
	ErrWorkspaceNameTooLong         = NewAppError("WORKSPACE_NAME_TOO_LONG", "workspace name exceeds maximum length")
	ErrWorkspaceDescriptionRequired = NewAppError("WORKSPACE_DESCRIPTION_REQUIRED", "workspace description is required")
	ErrWorkspaceDescriptionTooLong  = NewAppError("WORKSPACE_DESCRIPTION_TOO_LONG", "workspace description exceeds maximum length")
	ErrNotWorkspaceMember           = NewAppError("NOT_WORKSPACE_MEMBER", "not a workspace member")
	ErrInsufficientPermission       = NewAppError("INSUFFICIENT_PERMISSION", "insufficient permission")
	ErrCannotRemoveSelf             = NewAppError("CANNOT_REMOVE_SELF", "cannot remove yourself from the workspace")
)

// Transaction errors.
var (
	ErrTransactionNotFound           = NewAppError("TRANSACTION_NOT_FOUND", "transaction not found")
	ErrTransactionNoEntries          = NewAppError("TRANSACTION_NO_ENTRIES", "transaction must have at least one entry")
	ErrEntrySumMismatch              = NewAppError("ENTRY_SUM_MISMATCH", "entry amounts do not sum to transaction total")
	ErrDuplicateFingerprint          = NewAppError("DUPLICATE_FINGERPRINT", "duplicate transaction fingerprint")
	ErrTransactionAmountTooLarge     = NewAppError("TRANSACTION_AMOUNT_TOO_LARGE", "transaction amount exceeds maximum allowed value")
	ErrTransactionDateTooFarInPast   = NewAppError("TRANSACTION_DATE_TOO_FAR_IN_PAST", "transaction date is unreasonably far in the past")
	ErrTransactionDateTooFarInFuture = NewAppError("TRANSACTION_DATE_TOO_FAR_IN_FUTURE", "transaction date is too far in the future")
)

// Category errors.
var (
	ErrCategoryNotFound     = NewAppError("CATEGORY_NOT_FOUND", "category not found")
	ErrCategoryNameRequired = NewAppError("CATEGORY_NAME_REQUIRED", "category name is required")
	ErrCategoryNameTooLong  = NewAppError("CATEGORY_NAME_TOO_LONG", "category name exceeds maximum length")
	ErrCategoryUndeletable  = NewAppError("CATEGORY_UNDELETABLE", "the Uncategorized category cannot be deleted")
)

// Funding errors.
var (
	ErrFundingNotFound       = NewAppError("FUNDING_NOT_FOUND", "funding not found")
	ErrCategoryBudgetsExceed = NewAppError("CATEGORY_BUDGETS_EXCEED", "sum of category budgets exceeds overall budget")
)

// Invite errors.
var (
	ErrInviteNotFound      = NewAppError("INVITE_NOT_FOUND", "invite not found")
	ErrInviteExpired       = NewAppError("INVITE_EXPIRED", "invite has expired")
	ErrInviteUsed          = NewAppError("INVITE_USED", "invite has already been used")
	ErrAlreadyMember       = NewAppError("ALREADY_MEMBER", "already a member of this workspace")
	ErrCannotChangeOwnRole = NewAppError("CANNOT_CHANGE_OWN_ROLE", "cannot change own role")
	ErrInvalidInviteRole   = NewAppError("INVALID_INVITE_ROLE", "invalid invite role: must be editor or viewer")
)

// Rule errors.
var (
	ErrRuleNotFound           = NewAppError("RULE_NOT_FOUND", "rule not found")
	ErrRuleInvalidPattern     = NewAppError("RULE_INVALID_PATTERN", "invalid regex pattern")
	ErrRulePatternTooLong     = NewAppError("RULE_PATTERN_TOO_LONG", "regex pattern exceeds maximum length")
	ErrRuleInvalidAmountRange = NewAppError("RULE_INVALID_AMOUNT_RANGE", "invalid amount range")
)

// Bank connection errors.
var (
	ErrBankConnectionNotFound   = NewAppError("BANK_CONNECTION_NOT_FOUND", "bank connection not found")
	ErrBankAccountNotFound      = NewAppError("BANK_ACCOUNT_NOT_FOUND", "bank account not found")
	ErrBankAccountAlreadyExists = NewAppError("BANK_ACCOUNT_ALREADY_EXISTS", "bank account already exists")
	ErrBankConnectionExpired    = NewAppError("BANK_CONNECTION_EXPIRED", "bank connection has expired")
)

// Language errors.
var (
	ErrUnsupportedLanguage = NewAppError("UNSUPPORTED_LANGUAGE", "unsupported language")
)
