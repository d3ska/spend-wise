package model

import (
	"strings"
	"time"
)

// CategoryID is a typed wrapper for category identifiers.
type CategoryID int64

// Category field limits.
const MaxCategoryNameLen = 100

// Category represents a spending category scoped to a workspace.
type Category struct {
	ID          CategoryID
	WorkspaceID WorkspaceID
	Name        string
	Icon        string
	Slug        *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate returns an error if the category has invalid fields.
func (c Category) Validate() error {
	name := strings.TrimSpace(c.Name)
	if name == "" {
		return ErrCategoryNameRequired
	}
	if len(name) > MaxCategoryNameLen {
		return ErrCategoryNameTooLong
	}
	return nil
}
