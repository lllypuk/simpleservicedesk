//go:build integration
// +build integration

package repositories_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"simpleservicedesk/internal/application"
	domain "simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/infrastructure/tickets"
	"simpleservicedesk/test/integration/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// TicketRepositoryIntegrationSuite tests MongoDB repository
type TicketRepositoryIntegrationSuite struct {
	shared.MongoIntegrationSuite
	mongoRepo *tickets.MongoRepo
}

func (s *TicketRepositoryIntegrationSuite) SetupSuite() {
	s.MongoIntegrationSuite.SetupSuite()
	s.mongoRepo = tickets.NewMongoRepo(s.MongoDB)
}

func (s *TicketRepositoryIntegrationSuite) SetupTest() {
	s.MongoIntegrationSuite.SetupTest()
	// Clear MongoDB collections before each test
	ctx := context.Background()
	err := s.mongoRepo.Clear(ctx)
	s.Require().NoError(err)
}

// Test the same functionality across both repository implementations
func (s *TicketRepositoryIntegrationSuite) TestRepositoryCompatibility() {
	repos := map[string]application.TicketRepository{"MongoDB": s.mongoRepo}

	for repoName, repo := range repos {
		s.Run(repoName, func() {
			s.testBasicCRUD(repo)
			s.testComplexTicketOperations(repo)
			s.testFilteringAndSorting(repo)
			s.testErrorHandling(repo)
		})
	}
}

func (s *TicketRepositoryIntegrationSuite) testBasicCRUD(repo application.TicketRepository) {
	ctx := context.Background()

	// Test data
	orgID := uuid.New()
	authorID := uuid.New()
	assigneeID := uuid.New()

	s.Run("Create ticket", func() {
		testTicket := shared.NewTestTicket1(orgID, authorID)

		ticket, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return testTicket.CreateDomainTicket()
		})

		s.Require().NoError(err)
		s.Equal(testTicket.Title, ticket.Title())
		s.Equal(testTicket.Description, ticket.Description())
		s.Equal(testTicket.Priority, ticket.Priority())
		s.Equal(domain.StatusNew, ticket.Status())

		// Test Get
		retrievedTicket, err := repo.GetTicket(ctx, ticket.ID())
		s.Require().NoError(err)
		s.Equal(ticket.ID(), retrievedTicket.ID())
		s.Equal(ticket.Title(), retrievedTicket.Title())
		s.Equal(ticket.Description(), retrievedTicket.Description())

		// Test Update
		updatedTicket, err := repo.UpdateTicket(ctx, ticket.ID(), func(t *domain.Ticket) (bool, error) {
			if updateErr := t.UpdateTitle("Updated Title"); updateErr != nil {
				return false, updateErr
			}
			if assignErr := t.AssignTo(assigneeID); assignErr != nil {
				return false, assignErr
			}
			return true, nil
		})

		s.Require().NoError(err)
		s.Equal("Updated Title", updatedTicket.Title())
		s.Equal(&assigneeID, updatedTicket.AssigneeID())

		// Verify update persisted
		retrievedAgain, err := repo.GetTicket(ctx, ticket.ID())
		s.Require().NoError(err)
		s.Equal("Updated Title", retrievedAgain.Title())
		s.Equal(&assigneeID, retrievedAgain.AssigneeID())

		// Test Delete
		err = repo.DeleteTicket(ctx, ticket.ID())
		s.Require().NoError(err)

		// Verify deletion
		_, err = repo.GetTicket(ctx, ticket.ID())
		s.Equal(domain.ErrTicketNotFound, err)
	})
}

func (s *TicketRepositoryIntegrationSuite) testComplexTicketOperations(repo application.TicketRepository) {
	ctx := context.Background()

	s.Run("Ticket with comments and attachments", func() {
		orgID := uuid.New()
		authorID := uuid.New()
		commenterID := uuid.New()

		testTicket := shared.NewTestTicket2(orgID, authorID)

		// Create ticket
		ticket, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return testTicket.CreateDomainTicket()
		})
		s.Require().NoError(err)

		// Add comments and attachments
		_, err = repo.UpdateTicket(ctx, ticket.ID(), func(t *domain.Ticket) (bool, error) {
			// Add a comment
			if commentErr := t.AddComment(commenterID, "This is a test comment", false); commentErr != nil {
				return false, commentErr
			}

			// Add an internal comment
			if internalCommentErr := t.AddComment(commenterID, "Internal note", true); internalCommentErr != nil {
				return false, internalCommentErr
			}

			// Add an attachment
			if attachErr := t.AddAttachment("test.pdf", 1024, "application/pdf", "/uploads/test.pdf", commenterID); attachErr != nil {
				return false, attachErr
			}

			return true, nil
		})
		s.Require().NoError(err)

		// Verify complex data persisted
		retrievedTicket, err := repo.GetTicket(ctx, ticket.ID())
		s.Require().NoError(err)

		comments := retrievedTicket.Comments()
		s.Len(comments, 2)
		s.Equal("This is a test comment", comments[0].Content)
		s.False(comments[0].IsInternal)
		s.Equal("Internal note", comments[1].Content)
		s.True(comments[1].IsInternal)

		attachments := retrievedTicket.Attachments()
		s.Len(attachments, 1)
		s.Equal("test.pdf", attachments[0].FileName)
		s.Equal(int64(1024), attachments[0].FileSize)
		s.Equal("application/pdf", attachments[0].MimeType)
	})

	s.Run("Status transitions", func() {
		orgID := uuid.New()
		authorID := uuid.New()

		testTicket := shared.NewTestTicket3(orgID, authorID)

		// Create ticket
		ticket, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return testTicket.CreateDomainTicket()
		})
		s.Require().NoError(err)
		s.Equal(domain.StatusNew, ticket.Status())

		// Transition to InProgress
		_, err = repo.UpdateTicket(ctx, ticket.ID(), func(t *domain.Ticket) (bool, error) {
			return true, t.ChangeStatus(domain.StatusInProgress)
		})
		s.Require().NoError(err)

		// Transition to Resolved
		_, err = repo.UpdateTicket(ctx, ticket.ID(), func(t *domain.Ticket) (bool, error) {
			return true, t.ChangeStatus(domain.StatusResolved)
		})
		s.Require().NoError(err)

		// Verify final status
		finalTicket, err := repo.GetTicket(ctx, ticket.ID())
		s.Require().NoError(err)
		s.Equal(domain.StatusResolved, finalTicket.Status())
		s.NotNil(finalTicket.ResolvedAt())
	})
}

func (s *TicketRepositoryIntegrationSuite) testFilteringAndSorting(repo application.TicketRepository) {
	ctx := context.Background()

	s.Run("Advanced filtering and sorting", func() {
		// Clear any existing data for this specific test
		if mongoRepo, ok := repo.(*tickets.MongoRepo); ok {
			err := mongoRepo.Clear(ctx)
			s.Require().NoError(err)
		}

		// Create test data
		orgID1 := uuid.New()
		orgID2 := uuid.New()
		authorID1 := uuid.New()
		authorID2 := uuid.New()
		assigneeID := uuid.New()

		// Create tickets with different properties
		testTickets := []struct {
			title    string
			priority domain.Priority
			orgID    uuid.UUID
			authorID uuid.UUID
			assigned bool
		}{
			{"High Priority Ticket", domain.PriorityHigh, orgID1, authorID1, true},
			{"Normal Priority Ticket", domain.PriorityNormal, orgID2, authorID2, false},
			{"Critical Ticket", domain.PriorityCritical, orgID1, authorID1, false},
			{"Low Priority Ticket", domain.PriorityLow, orgID2, authorID1, true},
		}

		for i, testTicket := range testTickets {
			ticket, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
				t, newTicketErr := domain.NewTicket(
					uuid.New(),
					testTicket.title,
					"Description for "+testTicket.title,
					testTicket.priority,
					testTicket.orgID,
					testTicket.authorID,
					nil,
				)
				if newTicketErr != nil {
					return nil, newTicketErr
				}

				// Set different creation times for sorting tests
				t.SetCreatedAt(time.Now().Add(time.Duration(i) * time.Minute))
				return t, nil
			})
			s.Require().NoError(err)

			// Assign if needed
			if testTicket.assigned {
				_, err = repo.UpdateTicket(ctx, ticket.ID(), func(t *domain.Ticket) (bool, error) {
					return true, t.AssignTo(assigneeID)
				})
				s.Require().NoError(err)
			}
		}

		// Test filtering by priority
		highPriority := domain.PriorityHigh
		result, err := repo.ListTickets(ctx, application.TicketFilter{
			Priority: &highPriority,
		})
		s.Require().NoError(err)
		s.Len(result, 1)
		s.Equal("High Priority Ticket", result[0].Title())

		// Test filtering by organization
		result, err = repo.ListTickets(ctx, application.TicketFilter{
			OrganizationID: &orgID1,
		})
		s.Require().NoError(err)
		s.Len(result, 2) // High and Critical tickets

		// Test filtering by assignee
		result, err = repo.ListTickets(ctx, application.TicketFilter{
			AssigneeID: &assigneeID,
		})
		s.Require().NoError(err)
		s.Len(result, 2) // Assigned tickets

		// Test filtering by author
		result, err = repo.ListTickets(ctx, application.TicketFilter{
			AuthorID: &authorID1,
		})
		s.Require().NoError(err)
		s.Len(result, 3) // authorID1 created 3 tickets

		// Test pagination
		result, err = repo.ListTickets(ctx, application.TicketFilter{
			Limit:  2,
			Offset: 1,
		})
		s.Require().NoError(err)
		s.Len(result, 2)

		// Test sorting by priority
		result, err = repo.ListTickets(ctx, application.TicketFilter{
			SortBy:    "priority",
			SortOrder: "desc",
		})
		s.Require().NoError(err)
		s.Len(result, 4)
		// Should be ordered: Critical, High, Normal, Low
		s.Equal(domain.PriorityCritical, result[0].Priority())
		s.Equal(domain.PriorityHigh, result[1].Priority())
		s.Equal(domain.PriorityNormal, result[2].Priority())
		s.Equal(domain.PriorityLow, result[3].Priority())

		// Test date filtering
		now := time.Now()
		result, err = repo.ListTickets(ctx, application.TicketFilter{
			CreatedAfter:  &[]time.Time{now.Add(-1 * time.Hour)}[0],
			CreatedBefore: &[]time.Time{now.Add(1 * time.Hour)}[0],
		})
		s.Require().NoError(err)
		s.Len(result, 4) // All tickets should be within this range
	})
}

func (s *TicketRepositoryIntegrationSuite) testErrorHandling(repo application.TicketRepository) {
	ctx := context.Background()

	s.Run("Error handling", func() {
		// Test getting non-existent ticket
		_, err := repo.GetTicket(ctx, uuid.New())
		s.Equal(domain.ErrTicketNotFound, err)

		// Test updating non-existent ticket
		_, err = repo.UpdateTicket(ctx, uuid.New(), func(_ *domain.Ticket) (bool, error) {
			return true, nil
		})
		s.Equal(domain.ErrTicketNotFound, err)

		// Test deleting non-existent ticket
		err = repo.DeleteTicket(ctx, uuid.New())
		s.Equal(domain.ErrTicketNotFound, err)

		// Test createFn returning error
		expectedErr := errors.New("test error")
		_, err = repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return nil, expectedErr
		})
		s.Equal(expectedErr, err)

		// Test updateFn returning error
		testTicket := shared.NewTestTicket1(uuid.New(), uuid.New())
		ticket, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return testTicket.CreateDomainTicket()
		})
		s.Require().NoError(err)

		_, err = repo.UpdateTicket(ctx, ticket.ID(), func(_ *domain.Ticket) (bool, error) {
			return false, expectedErr
		})
		s.Equal(expectedErr, err)
	})
}

func (s *TicketRepositoryIntegrationSuite) TestConcurrentOperations() {
	// Test with MongoDB repo
	repo := s.mongoRepo
	ctx := context.Background()

	s.Run("Concurrent access", func() {
		testTicket := shared.NewTestTicket1(uuid.New(), uuid.New())

		// Create a ticket
		ticket, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return testTicket.CreateDomainTicket()
		})
		s.Require().NoError(err)

		// Test concurrent reads and updates
		done := make(chan bool, 20)

		// Concurrent readers
		for range 10 {
			go func() {
				defer func() { done <- true }()
				_, getErr := repo.GetTicket(ctx, ticket.ID())
				s.NoError(getErr)
			}()
		}

		// Concurrent writers
		for range 10 {
			go func() {
				defer func() { done <- true }()
				_, updateErr := repo.UpdateTicket(ctx, ticket.ID(), func(t *domain.Ticket) (bool, error) {
					return true, t.UpdateDescription("Updated by goroutine")
				})
				s.NoError(updateErr)
			}()
		}

		// Wait for all goroutines
		for range 20 {
			<-done
		}

		// Verify ticket still exists and is valid
		finalTicket, err := repo.GetTicket(ctx, ticket.ID())
		s.Require().NoError(err)
		s.NotNil(finalTicket)
		s.Equal("Updated by goroutine", finalTicket.Description())
	})
}

func (s *TicketRepositoryIntegrationSuite) TestDataConsistency() {
	repo := s.mongoRepo

	s.Run("Mongo data consistency", func() {
		ctx := context.Background()

		// Create tickets with specific data
		orgID := uuid.New()
		authorID := uuid.New()

		testTicket := shared.NewTestTicket2(orgID, authorID)

		originalTicket, err := repo.CreateTicket(ctx, func() (*domain.Ticket, error) {
			return testTicket.CreateDomainTicket()
		})
		s.Require().NoError(err)

		// Perform multiple operations
		_, err = repo.UpdateTicket(ctx, originalTicket.ID(), func(t *domain.Ticket) (bool, error) {
			// Multiple changes in one transaction
			if updateErr := t.UpdateTitle("Updated Title"); updateErr != nil {
				return false, updateErr
			}
			if updateErr := t.UpdateDescription("Updated Description"); updateErr != nil {
				return false, updateErr
			}
			if updateErr := t.UpdatePriority(domain.PriorityCritical); updateErr != nil {
				return false, updateErr
			}
			if updateErr := t.ChangeStatus(domain.StatusInProgress); updateErr != nil {
				return false, updateErr
			}
			return true, nil
		})
		s.Require().NoError(err)

		// Verify all changes are persisted together
		finalTicket, err := repo.GetTicket(ctx, originalTicket.ID())
		s.Require().NoError(err)
		s.Equal("Updated Title", finalTicket.Title())
		s.Equal("Updated Description", finalTicket.Description())
		s.Equal(domain.PriorityCritical, finalTicket.Priority())
		s.Equal(domain.StatusInProgress, finalTicket.Status())

		// Verify timestamps are updated
		s.True(finalTicket.UpdatedAt().After(finalTicket.CreatedAt()))
	})
}

func TestTicketRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(TicketRepositoryIntegrationSuite))
}
