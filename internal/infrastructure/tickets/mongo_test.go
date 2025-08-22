package tickets_test

import (
	"context"
	"testing"
	"time"

	"simpleservicedesk/internal/application"
	domain "simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/infrastructure/tickets"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupMongoTest(t *testing.T) (*tickets.MongoRepo, func()) {
	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx, "mongodb/mongodb-community-server:8.0-ubi8")
	require.NoError(t, err)

	// Get connection string
	endpoint, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	require.NoError(t, err)

	// Test connection
	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	// Create test database
	db := client.Database("test_tickets")
	repo := tickets.NewMongoRepo(db)

	cleanup := func() {
		client.Disconnect(ctx)
		mongoContainer.Terminate(ctx)
	}

	return repo, cleanup
}

func createTestTicket(t *testing.T) *domain.Ticket {
	ticket, err := domain.NewTicket(
		uuid.New(),
		"Test Ticket",
		"Test Description",
		domain.PriorityNormal,
		uuid.New(),
		uuid.New(),
		nil,
	)
	require.NoError(t, err)
	return ticket
}

func TestMongoRepo_CreateTicket(t *testing.T) {
	repo, cleanup := setupMongoTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		originalTicket := createTestTicket(t)

		createdTicket, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return originalTicket, nil
		})

		require.NoError(t, err)
		assert.Equal(t, originalTicket.ID(), createdTicket.ID())
		assert.Equal(t, originalTicket.Title(), createdTicket.Title())
		assert.Equal(t, originalTicket.Description(), createdTicket.Description())
		assert.Equal(t, originalTicket.Status(), createdTicket.Status())
		assert.Equal(t, originalTicket.Priority(), createdTicket.Priority())
	})

	t.Run("createFn returns error", func(t *testing.T) {
		expectedErr := assert.AnError

		createdTicket, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return nil, expectedErr
		})

		assert.Nil(t, createdTicket)
		assert.Equal(t, expectedErr, err)
	})
}

func TestMongoRepo_GetTicket(t *testing.T) {
	repo, cleanup := setupMongoTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("existing ticket", func(t *testing.T) {
		// Create a ticket first
		originalTicket := createTestTicket(t)
		_, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return originalTicket, nil
		})
		require.NoError(t, err)

		// Retrieve the ticket
		retrievedTicket, err := repo.GetTicket(ctx, originalTicket.ID())

		require.NoError(t, err)
		assert.Equal(t, originalTicket.ID(), retrievedTicket.ID())
		assert.Equal(t, originalTicket.Title(), retrievedTicket.Title())
		assert.Equal(t, originalTicket.Description(), retrievedTicket.Description())
		assert.Equal(t, originalTicket.Status(), retrievedTicket.Status())
		assert.Equal(t, originalTicket.Priority(), retrievedTicket.Priority())
	})

	t.Run("non-existing ticket", func(t *testing.T) {
		nonExistingID := uuid.New()

		retrievedTicket, err := repo.GetTicket(ctx, nonExistingID)

		assert.Nil(t, retrievedTicket)
		assert.Equal(t, domain.ErrTicketNotFound, err)
	})
}

func TestMongoRepo_UpdateTicket(t *testing.T) {
	repo, cleanup := setupMongoTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		// Create a ticket first
		originalTicket := createTestTicket(t)
		_, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return originalTicket, nil
		})
		require.NoError(t, err)

		// Update the ticket
		newTitle := "Updated Title"
		updatedTicket, err := repo.UpdateTicket(ctx, originalTicket.ID(), func(ticket *domain.Ticket) (bool, error) {
			return true, ticket.UpdateTitle(newTitle)
		})

		require.NoError(t, err)
		assert.Equal(t, newTitle, updatedTicket.Title())

		// Verify the update was persisted
		retrievedTicket, err := repo.GetTicket(ctx, originalTicket.ID())
		require.NoError(t, err)
		assert.Equal(t, newTitle, retrievedTicket.Title())
	})

	t.Run("update function returns no changes", func(t *testing.T) {
		// Create a ticket first
		originalTicket := createTestTicket(t)
		_, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return originalTicket, nil
		})
		require.NoError(t, err)

		// Call update but don't make changes
		updatedTicket, err := repo.UpdateTicket(ctx, originalTicket.ID(), func(_ *domain.Ticket) (bool, error) {
			return false, nil // No changes
		})

		require.NoError(t, err)
		assert.Equal(t, originalTicket.Title(), updatedTicket.Title())
	})

	t.Run("non-existing ticket", func(t *testing.T) {
		nonExistingID := uuid.New()

		updatedTicket, err := repo.UpdateTicket(ctx, nonExistingID, func(_ *domain.Ticket) (bool, error) {
			return true, nil
		})

		assert.Nil(t, updatedTicket)
		assert.Equal(t, domain.ErrTicketNotFound, err)
	})

	t.Run("update function returns error", func(t *testing.T) {
		// Create a ticket first
		originalTicket := createTestTicket(t)
		_, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return originalTicket, nil
		})
		require.NoError(t, err)

		expectedErr := assert.AnError

		updatedTicket, err := repo.UpdateTicket(ctx, originalTicket.ID(), func(_ *domain.Ticket) (bool, error) {
			return false, expectedErr
		})

		assert.Nil(t, updatedTicket)
		assert.Equal(t, expectedErr, err)
	})
}

func TestMongoRepo_DeleteTicket(t *testing.T) {
	repo, cleanup := setupMongoTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		// Create a ticket first
		originalTicket := createTestTicket(t)
		_, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return originalTicket, nil
		})
		require.NoError(t, err)

		// Delete the ticket
		err = repo.DeleteTicket(ctx, originalTicket.ID())
		require.NoError(t, err)

		// Verify the ticket is deleted
		retrievedTicket, err := repo.GetTicket(ctx, originalTicket.ID())
		assert.Nil(t, retrievedTicket)
		assert.Equal(t, domain.ErrTicketNotFound, err)
	})

	t.Run("non-existing ticket", func(t *testing.T) {
		nonExistingID := uuid.New()

		err := repo.DeleteTicket(ctx, nonExistingID)
		assert.Equal(t, domain.ErrTicketNotFound, err)
	})
}

func TestMongoRepo_ListTickets(t *testing.T) {
	repo, cleanup := setupMongoTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create test sliceTickets with different properties
	orgID1 := uuid.New()
	orgID2 := uuid.New()
	authorID1 := uuid.New()
	authorID2 := uuid.New()

	var sliceTickets []*domain.Ticket

	// Ticket 1: High priority, orgID1, authorID1
	ticket1, err := domain.NewTicket(
		uuid.New(),
		"Ticket 1",
		"Description 1",
		domain.PriorityHigh,
		orgID1,
		authorID1,
		nil,
	)
	require.NoError(t, err)
	sliceTickets = append(sliceTickets, ticket1)

	// Ticket 2: Normal priority, orgID2, authorID2
	ticket2, err := domain.NewTicket(
		uuid.New(),
		"Ticket 2",
		"Description 2",
		domain.PriorityNormal,
		orgID2,
		authorID2,
		nil,
	)
	require.NoError(t, err)
	sliceTickets = append(sliceTickets, ticket2)

	// Ticket 3: Critical priority, orgID1, authorID1
	ticket3, err := domain.NewTicket(
		uuid.New(),
		"Ticket 3",
		"Description 3",
		domain.PriorityCritical,
		orgID1,
		authorID1,
		nil,
	)
	require.NoError(t, err)
	sliceTickets = append(sliceTickets, ticket3)

	// Insert all sliceTickets
	for _, ticket := range sliceTickets {
		_, createErr := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return ticket, nil
		})
		require.NoError(t, createErr)
	}

	t.Run("no filter", func(t *testing.T) {
		result, ticketErr := repo.ListTickets(ctx, application.TicketFilter{})
		require.NoError(t, ticketErr)
		assert.Len(t, result, 3)
	})

	t.Run("filter by priority", func(t *testing.T) {
		priority := domain.PriorityHigh
		result, ticketErr := repo.ListTickets(ctx, application.TicketFilter{
			Priority: &priority,
		})
		require.NoError(t, ticketErr)
		assert.Len(t, result, 1)
		assert.Equal(t, domain.PriorityHigh, result[0].Priority())
	})

	t.Run("filter by organization", func(t *testing.T) {
		result, ticketErr := repo.ListTickets(ctx, application.TicketFilter{
			OrganizationID: &orgID1,
		})
		require.NoError(t, ticketErr)
		assert.Len(t, result, 2) // ticket1 and ticket3
	})

	t.Run("filter by author", func(t *testing.T) {
		result, ticketErr := repo.ListTickets(ctx, application.TicketFilter{
			AuthorID: &authorID2,
		})
		require.NoError(t, ticketErr)
		assert.Len(t, result, 1) // ticket2
		assert.Equal(t, "Ticket 2", result[0].Title())
	})

	t.Run("with limit", func(t *testing.T) {
		result, ticketErr := repo.ListTickets(ctx, application.TicketFilter{
			Limit: 2,
		})
		require.NoError(t, ticketErr)
		assert.Len(t, result, 2)
	})

	t.Run("with offset", func(t *testing.T) {
		result, ticketErr := repo.ListTickets(ctx, application.TicketFilter{
			Offset: 1,
		})
		require.NoError(t, ticketErr)
		assert.Len(t, result, 2)
	})

	t.Run("sort by priority descending", func(t *testing.T) {
		result, ticketErr := repo.ListTickets(ctx, application.TicketFilter{
			SortBy:    "priority",
			SortOrder: "desc",
		})
		require.NoError(t, ticketErr)
		assert.Len(t, result, 3)
		// Should be ordered: Critical, High, Normal
		assert.Equal(t, domain.PriorityCritical, result[0].Priority())
		assert.Equal(t, domain.PriorityHigh, result[1].Priority())
		assert.Equal(t, domain.PriorityNormal, result[2].Priority())
	})

	t.Run("date range filter", func(t *testing.T) {
		now := time.Now()
		before := now.Add(-1 * time.Hour)
		after := now.Add(1 * time.Hour)

		result, ticketErr := repo.ListTickets(ctx, application.TicketFilter{
			CreatedAfter:  &before,
			CreatedBefore: &after,
		})
		require.NoError(t, ticketErr)
		assert.Len(t, result, 3) // All sliceTickets should be within this range
	})
}
