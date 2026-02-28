package store

import (
	"context"
	"errors"
	"fmt"

	"backend/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// CategoryStore provides category persistence operations.
type CategoryStore struct {
	q *Queries
}

// NewCategoryStore creates a new CategoryStore.
func NewCategoryStore(db DBTX) *CategoryStore {
	return &CategoryStore{q: New(db)}
}

// Create persists a new category and returns it with a generated ID.
func (s *CategoryStore) Create(ctx context.Context, cat model.Category) (model.Category, error) {
	var slug pgtype.Text
	if cat.Slug != nil {
		slug = pgtype.Text{String: *cat.Slug, Valid: true}
	}
	row, err := s.q.InsertCategory(ctx, InsertCategoryParams{
		WorkspaceID: int64(cat.WorkspaceID),
		Name:        cat.Name,
		Icon:        cat.Icon,
		Slug:        slug,
	})
	if err != nil {
		return model.Category{}, fmt.Errorf("inserting category: %w", err)
	}
	return toModelCategory(row), nil
}

// GetByID returns a category by ID. Returns model.ErrCategoryNotFound if not found.
func (s *CategoryStore) GetByID(ctx context.Context, id model.CategoryID) (model.Category, error) {
	row, err := s.q.GetCategoryByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, model.ErrCategoryNotFound
		}
		return model.Category{}, fmt.Errorf("getting category by id: %w", err)
	}
	return toModelCategory(row), nil
}

// ListByWorkspace returns all categories belonging to a workspace.
func (s *CategoryStore) ListByWorkspace(ctx context.Context, wsID model.WorkspaceID) ([]model.Category, error) {
	rows, err := s.q.ListCategoriesByWorkspace(ctx, int64(wsID))
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	result := make([]model.Category, 0, len(rows))
	for _, row := range rows {
		result = append(result, toModelCategory(row))
	}
	return result, nil
}

// Update updates a category's name and icon.
func (s *CategoryStore) Update(ctx context.Context, id model.CategoryID, name, icon string) (model.Category, error) {
	row, err := s.q.UpdateCategory(ctx, UpdateCategoryParams{
		ID:   int64(id),
		Name: name,
		Icon: icon,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, model.ErrCategoryNotFound
		}
		return model.Category{}, fmt.Errorf("updating category: %w", err)
	}
	return toModelCategory(row), nil
}

// GetByName returns a category by workspace ID and name. Returns model.ErrCategoryNotFound if not found.
func (s *CategoryStore) GetByName(ctx context.Context, wsID model.WorkspaceID, name string) (model.Category, error) {
	row, err := s.q.GetCategoryByName(ctx, GetCategoryByNameParams{
		WorkspaceID: int64(wsID),
		Name:        name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, model.ErrCategoryNotFound
		}
		return model.Category{}, fmt.Errorf("getting category by name: %w", err)
	}
	return toModelCategory(row), nil
}

// GetBySlug returns a category by workspace ID and slug. Returns model.ErrCategoryNotFound if not found.
func (s *CategoryStore) GetBySlug(ctx context.Context, wsID model.WorkspaceID, slug string) (model.Category, error) {
	row, err := s.q.GetCategoryBySlug(ctx, GetCategoryBySlugParams{
		WorkspaceID: int64(wsID),
		Slug:        pgtype.Text{String: slug, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Category{}, model.ErrCategoryNotFound
		}
		return model.Category{}, fmt.Errorf("getting category by slug: %w", err)
	}
	return toModelCategory(row), nil
}

// Delete removes a category by ID.
func (s *CategoryStore) Delete(ctx context.Context, id model.CategoryID) error {
	if err := s.q.DeleteCategory(ctx, int64(id)); err != nil {
		return fmt.Errorf("deleting category: %w", err)
	}
	return nil
}

func toModelCategory(row Category) model.Category {
	var slug *string
	if row.Slug.Valid {
		slug = &row.Slug.String
	}
	return model.Category{
		ID:          model.CategoryID(row.ID),
		WorkspaceID: model.WorkspaceID(row.WorkspaceID),
		Name:        row.Name,
		Icon:        row.Icon,
		Slug:        slug,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
