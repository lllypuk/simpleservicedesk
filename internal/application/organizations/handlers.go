package organizations

import (
	"context"

	"simpleservicedesk/internal/domain/organizations"
	"simpleservicedesk/internal/queries"

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
	ListOrganizations(ctx context.Context, filter queries.OrganizationFilter) ([]*organizations.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
}

type OrganizationHandlers struct {
	repo Repository
}

func SetupHandlers(repo Repository) OrganizationHandlers {
	return OrganizationHandlers{
		repo: repo,
	}
}
