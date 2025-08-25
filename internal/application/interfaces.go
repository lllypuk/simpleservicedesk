package application

import (
	"context"

	"simpleservicedesk/internal/domain/categories"
	"simpleservicedesk/internal/domain/organizations"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context,
		email string,
		passwordHash []byte,
		createFn func() (*users.User, error)) (*users.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, updateFn func(*users.User) (bool, error)) (*users.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*users.User, error)
	ListUsers(ctx context.Context, filter queries.UserFilter) ([]*users.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	CountUsers(ctx context.Context, filter queries.UserFilter) (int64, error)
}

type TicketRepository interface {
	CreateTicket(ctx context.Context, createFn func() (*tickets.Ticket, error)) (*tickets.Ticket, error)
	UpdateTicket(
		ctx context.Context,
		id uuid.UUID,
		updateFn func(*tickets.Ticket) (bool, error),
	) (*tickets.Ticket, error)
	GetTicket(ctx context.Context, id uuid.UUID) (*tickets.Ticket, error)
	ListTickets(ctx context.Context, filter queries.TicketFilter) ([]*tickets.Ticket, error)
	DeleteTicket(ctx context.Context, id uuid.UUID) error
}

// CategoryTree represents a hierarchical category structure
type CategoryTree struct {
	Category *categories.Category `json:"category"`
	Children []*CategoryTree      `json:"children,omitempty"`
}

//nolint:dupl // CategoryRepository and OrganizationRepository have similar patterns by design
type CategoryRepository interface {
	CreateCategory(ctx context.Context, createFn func() (*categories.Category, error)) (*categories.Category, error)
	UpdateCategory(
		ctx context.Context,
		id uuid.UUID,
		updateFn func(*categories.Category) (bool, error),
	) (*categories.Category, error)
	GetCategory(ctx context.Context, id uuid.UUID) (*categories.Category, error)
	ListCategories(ctx context.Context, filter queries.CategoryFilter) ([]*categories.Category, error)
	GetCategoryHierarchy(ctx context.Context, rootID uuid.UUID) (*CategoryTree, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// OrganizationTree represents a hierarchical organization structure
type OrganizationTree struct {
	Organization *organizations.Organization `json:"organization"`
	Children     []*OrganizationTree         `json:"children,omitempty"`
}

//nolint:dupl // CategoryRepository and OrganizationRepository have similar patterns by design
type OrganizationRepository interface {
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
	GetOrganizationHierarchy(ctx context.Context, rootID uuid.UUID) (*OrganizationTree, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
}
