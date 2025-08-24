package organizations

import (
	"context"

	"simpleservicedesk/internal/domain/organizations"

	"github.com/google/uuid"
)

const DefaultPageLimit = 20

type Repository interface {
	CreateOrganization(
		ctx context.Context,
		createFn func() (*organizations.Organization, error),
	) (*organizations.Organization, error)
	UpdateOrganization(
		ctx context.Context,
		id uuid.UUID,
		updateFn func(*organizations.Organization) (bool, error),
	) (*organizations.Organization, error)
	GetOrganization(ctx context.Context, id uuid.UUID) (*organizations.Organization, error)
	ListOrganizations(ctx context.Context, filter OrganizationFilter) ([]*organizations.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
}

type OrganizationFilter struct {
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
	Name       *string    `json:"name,omitempty"`
	Domain     *string    `json:"domain,omitempty"`
	IsRootOnly bool       `json:"is_root_only,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
	SortBy     string     `json:"sort_by,omitempty"`
	SortOrder  string     `json:"sort_order,omitempty"`
}

type OrganizationHandlers struct {
	repo Repository
}

func SetupHandlers(repo Repository) OrganizationHandlers {
	return OrganizationHandlers{
		repo: repo,
	}
}
