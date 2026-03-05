package tickets_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"simpleservicedesk/generated/openapi"
	apptickets "simpleservicedesk/internal/application/tickets"
	authdomain "simpleservicedesk/internal/domain/auth"
	ticketdomain "simpleservicedesk/internal/domain/tickets"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"
	"simpleservicedesk/pkg/contextkeys"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type ticketRepoSpy struct {
	listCalled bool
	listFilter queries.TicketFilter
}

func (r *ticketRepoSpy) CreateTicket(
	_ context.Context,
	_ func() (*ticketdomain.Ticket, error),
) (*ticketdomain.Ticket, error) {
	panic("unexpected CreateTicket call")
}

func (r *ticketRepoSpy) UpdateTicket(
	_ context.Context,
	_ uuid.UUID,
	_ func(*ticketdomain.Ticket) (bool, error),
) (*ticketdomain.Ticket, error) {
	panic("unexpected UpdateTicket call")
}

func (r *ticketRepoSpy) GetTicket(_ context.Context, _ uuid.UUID) (*ticketdomain.Ticket, error) {
	panic("unexpected GetTicket call")
}

func (r *ticketRepoSpy) ListTickets(_ context.Context, filter queries.TicketFilter) ([]*ticketdomain.Ticket, error) {
	r.listCalled = true
	r.listFilter = filter
	return []*ticketdomain.Ticket{}, nil
}

func (r *ticketRepoSpy) DeleteTicket(_ context.Context, _ uuid.UUID) error {
	panic("unexpected DeleteTicket call")
}

func TestGetTicketsUsesAuthContext(t *testing.T) {
	t.Run("customer role is forced to own author id", func(t *testing.T) {
		repo := &ticketRepoSpy{}
		handlers := apptickets.SetupHandlers(repo)

		customerID := uuid.New()
		otherAuthorID := uuid.New()

		params := openapi.GetTicketsParams{
			AuthorId: &otherAuthorID,
		}

		c, rec := newTicketContextWithClaims(&authdomain.Claims{
			UserID: customerID.String(),
			Role:   userdomain.RoleCustomer,
		})

		err := handlers.GetTickets(c, params)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.True(t, repo.listCalled)
		require.NotNil(t, repo.listFilter.AuthorID)
		require.Equal(t, customerID, *repo.listFilter.AuthorID)
	})

	t.Run("agent role keeps explicit author filter", func(t *testing.T) {
		repo := &ticketRepoSpy{}
		handlers := apptickets.SetupHandlers(repo)

		authorID := uuid.New()
		params := openapi.GetTicketsParams{
			AuthorId: &authorID,
		}

		c, rec := newTicketContextWithClaims(&authdomain.Claims{
			UserID: uuid.NewString(),
			Role:   userdomain.RoleAgent,
		})

		err := handlers.GetTickets(c, params)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
		require.True(t, repo.listCalled)
		require.NotNil(t, repo.listFilter.AuthorID)
		require.Equal(t, authorID, *repo.listFilter.AuthorID)
	})

	t.Run("missing auth claims returns unauthorized", func(t *testing.T) {
		repo := &ticketRepoSpy{}
		handlers := apptickets.SetupHandlers(repo)

		c, rec := newTicketContextWithClaims(nil)

		err := handlers.GetTickets(c, openapi.GetTicketsParams{})
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.False(t, repo.listCalled)
	})
}

func newTicketContextWithClaims(claims *authdomain.Claims) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
	if claims != nil {
		ctx := context.WithValue(req.Context(), contextkeys.AuthClaimsCtxKey, claims)
		req = req.WithContext(ctx)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}
