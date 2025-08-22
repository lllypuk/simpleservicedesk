package application

import (
	"context"
	"time"

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
	UpdateTicket(ctx context.Context, id uuid.UUID, updateFn func(*tickets.Ticket) (bool, error)) (*tickets.Ticket, error)
	GetTicket(ctx context.Context, id uuid.UUID) (*tickets.Ticket, error)
	ListTickets(ctx context.Context, filter TicketFilter) ([]*tickets.Ticket, error)
	DeleteTicket(ctx context.Context, id uuid.UUID) error
}
