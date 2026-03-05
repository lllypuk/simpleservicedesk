package categories

import (
	"context"

	"simpleservicedesk/internal/domain/categories"
	ticketdomain "simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/queries"

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
	ListCategories(ctx context.Context, filter queries.CategoryFilter) ([]*categories.Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}

type TicketRepository interface {
	ListTickets(ctx context.Context, filter queries.TicketFilter) ([]*ticketdomain.Ticket, error)
}

type CategoryHandlers struct {
	repo       CategoryRepository
	ticketRepo TicketRepository
}

func SetupHandlers(repo CategoryRepository, ticketRepo TicketRepository) CategoryHandlers {
	return CategoryHandlers{
		repo:       repo,
		ticketRepo: ticketRepo,
	}
}
