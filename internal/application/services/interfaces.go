package services

import (
	"context"

	"simpleservicedesk/internal/domain/categories"
	"simpleservicedesk/internal/domain/organizations"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
)

// Common request/response types for services

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Name     string
	Email    string
	Password string
}

type UpdateUserRequest struct {
	Name     *string
	Email    *string
	Password *string
	IsActive *bool
}

// CreateTicketRequest represents a request to create a new ticket
type CreateTicketRequest struct {
	Title          string
	Description    string
	Priority       tickets.Priority
	CategoryID     *uuid.UUID
	OrganizationID *uuid.UUID
	AssignedToID   *uuid.UUID
}

type UpdateTicketRequest struct {
	Title          *string
	Description    *string
	Priority       *tickets.Priority
	CategoryID     *uuid.UUID
	OrganizationID *uuid.UUID
	AssignedToID   *uuid.UUID
}

type AddTicketCommentRequest struct {
	Content    string
	IsInternal bool
}

// CreateCategoryRequest represents a request to create a new category
type CreateCategoryRequest struct {
	Name           string
	Description    *string
	OrganizationID uuid.UUID
	ParentID       *uuid.UUID
}

type UpdateCategoryRequest struct {
	Name        *string
	Description *string
	ParentID    *uuid.UUID
	IsActive    *bool
}

// CreateOrganizationRequest represents a request to create a new organization
type CreateOrganizationRequest struct {
	Name        string
	Domain      string
	Description *string
	ParentID    *uuid.UUID
}

type UpdateOrganizationRequest struct {
	Name        *string
	Domain      *string
	Description *string
	ParentID    *uuid.UUID
	IsActive    *bool
}

// Service interfaces

// UserService provides business logic for user operations
type UserService interface {
	CreateUser(ctx context.Context, req CreateUserRequest) (*users.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*users.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*users.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, filter queries.UserFilter) ([]*users.User, int64, error)
	UpdateUserRole(ctx context.Context, id uuid.UUID, role users.Role) (*users.User, error)
	GetUserTickets(ctx context.Context, userID uuid.UUID, filter queries.TicketFilter) ([]*tickets.Ticket, int64, error)
}

// TicketService provides business logic for ticket operations
type TicketService interface {
	CreateTicket(ctx context.Context, createdByID uuid.UUID, req CreateTicketRequest) (*tickets.Ticket, error)
	GetTicket(ctx context.Context, id uuid.UUID) (*tickets.Ticket, error)
	UpdateTicket(ctx context.Context, id uuid.UUID, req UpdateTicketRequest) (*tickets.Ticket, error)
	DeleteTicket(ctx context.Context, id uuid.UUID) error
	ListTickets(ctx context.Context, filter queries.TicketFilter) ([]*tickets.Ticket, int64, error)
	UpdateTicketStatus(ctx context.Context, id uuid.UUID, status tickets.Status) (*tickets.Ticket, error)
	AssignTicket(ctx context.Context, id uuid.UUID, assignedToID *uuid.UUID) (*tickets.Ticket, error)
	AddComment(
		ctx context.Context,
		ticketID uuid.UUID,
		authorID uuid.UUID,
		req AddTicketCommentRequest,
	) (*tickets.Comment, error)
	GetTicketComments(
		ctx context.Context,
		ticketID uuid.UUID,
		filter queries.CommentFilter,
	) ([]*tickets.Comment, int64, error)
}

// CategoryService provides business logic for category operations
type CategoryService interface {
	CreateCategory(ctx context.Context, req CreateCategoryRequest) (*categories.Category, error)
	GetCategory(ctx context.Context, id uuid.UUID) (*categories.Category, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, req UpdateCategoryRequest) (*categories.Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	ListCategories(ctx context.Context, filter queries.CategoryFilter) ([]*categories.Category, int64, error)
	GetCategoryTickets(
		ctx context.Context,
		categoryID uuid.UUID,
		filter queries.TicketFilter,
	) ([]*tickets.Ticket, int64, error)
}

// OrganizationService provides business logic for organization operations
type OrganizationService interface {
	CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*organizations.Organization, error)
	GetOrganization(ctx context.Context, id uuid.UUID) (*organizations.Organization, error)
	UpdateOrganization(
		ctx context.Context,
		id uuid.UUID,
		req UpdateOrganizationRequest,
	) (*organizations.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	ListOrganizations(
		ctx context.Context,
		filter queries.OrganizationFilter,
	) ([]*organizations.Organization, int64, error)
	GetOrganizationUsers(ctx context.Context, orgID uuid.UUID, filter queries.UserFilter) ([]*users.User, int64, error)
	GetOrganizationTickets(
		ctx context.Context,
		orgID uuid.UUID,
		filter queries.TicketFilter,
	) ([]*tickets.Ticket, int64, error)
}
