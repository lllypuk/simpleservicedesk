package application

import (
	"context"
	"time"

	"simpleservicedesk/internal/domain/categories"
	"simpleservicedesk/internal/domain/organizations"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context,
		email string,
		passwordHash []byte,
		createFn func() (*users.User, error)) (*users.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, updateFn func(*users.User) (bool, error)) (*users.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*users.User, error)
}

// TicketFilter defines filtering options for ticket queries
type TicketFilter struct {
	Status         *tickets.Status   `json:"status,omitempty"`
	Priority       *tickets.Priority `json:"priority,omitempty"`
	AssigneeID     *uuid.UUID        `json:"assignee_id,omitempty"`
	AuthorID       *uuid.UUID        `json:"author_id,omitempty"`
	OrganizationID *uuid.UUID        `json:"organization_id,omitempty"`
	CategoryID     *uuid.UUID        `json:"category_id,omitempty"`
	CreatedAfter   *time.Time        `json:"created_after,omitempty"`
	CreatedBefore  *time.Time        `json:"created_before,omitempty"`
	UpdatedAfter   *time.Time        `json:"updated_after,omitempty"`
	UpdatedBefore  *time.Time        `json:"updated_before,omitempty"`
	IsOverdue      *bool             `json:"is_overdue,omitempty"`
	Limit          int               `json:"limit,omitempty"`
	Offset         int               `json:"offset,omitempty"`
	SortBy         string            `json:"sort_by,omitempty"`    // "created_at", "updated_at", "priority"
	SortOrder      string            `json:"sort_order,omitempty"` // "asc", "desc"
}

type TicketRepository interface {
	CreateTicket(ctx context.Context, createFn func() (*tickets.Ticket, error)) (*tickets.Ticket, error)
	UpdateTicket(
		ctx context.Context,
		id uuid.UUID,
		updateFn func(*tickets.Ticket) (bool, error),
	) (*tickets.Ticket, error)
	GetTicket(ctx context.Context, id uuid.UUID) (*tickets.Ticket, error)
	ListTickets(ctx context.Context, filter TicketFilter) ([]*tickets.Ticket, error)
	DeleteTicket(ctx context.Context, id uuid.UUID) error
}

// CategoryFilter defines filtering options for category queries
type CategoryFilter struct {
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	Name           *string    `json:"name,omitempty"`
	IsRootOnly     bool       `json:"is_root_only,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	Offset         int        `json:"offset,omitempty"`
	SortBy         string     `json:"sort_by,omitempty"`    // "name", "created_at", "updated_at"
	SortOrder      string     `json:"sort_order,omitempty"` // "asc", "desc"
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
	ListCategories(ctx context.Context, filter CategoryFilter) ([]*categories.Category, error)
	GetCategoryHierarchy(ctx context.Context, rootID uuid.UUID) (*CategoryTree, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

// OrganizationFilter defines filtering options for organization queries
type OrganizationFilter struct {
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
	Name       *string    `json:"name,omitempty"`
	Domain     *string    `json:"domain,omitempty"`
	IsRootOnly bool       `json:"is_root_only,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
	SortBy     string     `json:"sort_by,omitempty"`    // "name", "created_at", "updated_at", "domain"
	SortOrder  string     `json:"sort_order,omitempty"` // "asc", "desc"
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
	ListOrganizations(ctx context.Context, filter OrganizationFilter) ([]*organizations.Organization, error)
	GetOrganizationHierarchy(ctx context.Context, rootID uuid.UUID) (*OrganizationTree, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
}
