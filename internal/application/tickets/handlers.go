package tickets

import (
	"context"

	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"

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
	ListTickets(ctx context.Context, filter queries.TicketFilter) ([]*tickets.Ticket, error)
	DeleteTicket(ctx context.Context, id uuid.UUID) error
}

type UserRepository interface {
	GetUser(ctx context.Context, id uuid.UUID) (*users.User, error)
}

type TicketHandlers struct {
	repo     TicketRepository
	userRepo UserRepository
}

func SetupHandlers(repo TicketRepository, userRepo ...UserRepository) TicketHandlers {
	var usersRepo UserRepository
	if len(userRepo) > 0 {
		usersRepo = userRepo[0]
	}

	return TicketHandlers{
		repo:     repo,
		userRepo: usersRepo,
	}
}
