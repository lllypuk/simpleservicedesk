package categories

import (
	"context"

	"simpleservicedesk/internal/domain/categories"

	"github.com/google/uuid"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, createFn func() (*categories.Category, error)) (*categories.Category, error)
	UpdateCategory(
		ctx context.Context,
		id uuid.UUID,
		updateFn func(*categories.Category) (bool, error),
	) (*categories.Category, error)
	GetCategory(ctx context.Context, id uuid.UUID) (*categories.Category, error)
	ListCategories(ctx context.Context, filter CategoryFilter) ([]*categories.Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

type CategoryFilter struct {
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	Name           *string    `json:"name,omitempty"`
	IsRootOnly     bool       `json:"is_root_only,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
	SortBy         string     `json:"sort_by,omitempty"`
	SortOrder      string     `json:"sort_order,omitempty"`
}

type CategoryHandlers struct {
	repo CategoryRepository
}

func SetupHandlers(repo CategoryRepository) CategoryHandlers {
	return CategoryHandlers{
		repo: repo,
	}
}
