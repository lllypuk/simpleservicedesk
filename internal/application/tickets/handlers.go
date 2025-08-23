package tickets

import (
	"context"

	"simpleservicedesk/internal/domain/tickets"

	"github.com/google/uuid"
)

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

type TicketFilter struct {
	Status         *tickets.Status   `json:"status,omitempty"`
	Priority       *tickets.Priority `json:"priority,omitempty"`
	AssigneeID     *uuid.UUID        `json:"assignee_id,omitempty"`
	AuthorID       *uuid.UUID        `json:"author_id,omitempty"`
	OrganizationID *uuid.UUID        `json:"organization_id,omitempty"`
	CategoryID     *uuid.UUID        `json:"category_id,omitempty"`
	Limit          int               `json:"limit,omitempty"`
	Offset         int               `json:"offset,omitempty"`
}

type TicketHandlers struct {
	repo TicketRepository
}

func SetupHandlers(repo TicketRepository) TicketHandlers {
	return TicketHandlers{
		repo: repo,
	}
}
