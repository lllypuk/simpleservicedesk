package services

import (
	"context"
	"errors"

	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/interfaces"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
)

// ticketService implements TicketService interface
type ticketService struct {
	ticketRepo interfaces.TicketRepository
}

// NewTicketService creates a new TicketService implementation
func NewTicketService(ticketRepo interfaces.TicketRepository) TicketService {
	return &ticketService{
		ticketRepo: ticketRepo,
	}
}

// CreateTicket creates a new ticket with business logic validation
func (s *ticketService) CreateTicket(
	ctx context.Context,
	createdByID uuid.UUID,
	req CreateTicketRequest,
) (*tickets.Ticket, error) {
	// Create ticket through repository
	ticket, err := s.ticketRepo.CreateTicket(ctx, func() (*tickets.Ticket, error) {
		// Handle organization ID - if nil, use empty UUID for now
		orgID := uuid.Nil
		if req.OrganizationID != nil {
			orgID = *req.OrganizationID
		}

		return tickets.NewTicket(
			uuid.New(),
			req.Title,
			req.Description,
			req.Priority,
			orgID,
			createdByID,
			req.CategoryID,
		)
	})
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

// GetTicket retrieves a ticket by ID
func (s *ticketService) GetTicket(ctx context.Context, id uuid.UUID) (*tickets.Ticket, error) {
	return s.ticketRepo.GetTicket(ctx, id)
}

// UpdateTicket updates ticket information with business logic
func (s *ticketService) UpdateTicket(
	ctx context.Context,
	id uuid.UUID,
	req UpdateTicketRequest,
) (*tickets.Ticket, error) {
	ticket, err := s.ticketRepo.UpdateTicket(ctx, id, func(ticket *tickets.Ticket) (bool, error) {
		return s.applyTicketUpdates(ticket, req)
	})

	if err != nil {
		return nil, err
	}

	return ticket, nil
}

// applyTicketUpdates applies all ticket updates from the request
func (s *ticketService) applyTicketUpdates(ticket *tickets.Ticket, req UpdateTicketRequest) (bool, error) {
	updated := false

	if titleUpdated, err := s.updateTicketTitle(ticket, req.Title); err != nil {
		return false, err
	} else if titleUpdated {
		updated = true
	}

	if descUpdated, err := s.updateTicketDescription(ticket, req.Description); err != nil {
		return false, err
	} else if descUpdated {
		updated = true
	}

	if priorityUpdated, err := s.updateTicketPriority(ticket, req.Priority); err != nil {
		return false, err
	} else if priorityUpdated {
		updated = true
	}

	if categoryUpdated := s.updateTicketCategory(ticket, req.CategoryID); categoryUpdated {
		updated = true
	}

	if assigneeUpdated, err := s.updateTicketAssignee(ticket, req.AssignedToID); err != nil {
		return false, err
	} else if assigneeUpdated {
		updated = true
	}

	return updated, nil
}

// updateTicketTitle updates the ticket title if provided
func (s *ticketService) updateTicketTitle(ticket *tickets.Ticket, title *string) (bool, error) {
	if title == nil {
		return false, nil
	}
	if err := ticket.UpdateTitle(*title); err != nil {
		return false, err
	}
	return true, nil
}

// updateTicketDescription updates the ticket description if provided
func (s *ticketService) updateTicketDescription(ticket *tickets.Ticket, description *string) (bool, error) {
	if description == nil {
		return false, nil
	}
	if err := ticket.UpdateDescription(*description); err != nil {
		return false, err
	}
	return true, nil
}

// updateTicketPriority updates the ticket priority if provided
func (s *ticketService) updateTicketPriority(ticket *tickets.Ticket, priority *tickets.Priority) (bool, error) {
	if priority == nil {
		return false, nil
	}
	if err := ticket.UpdatePriority(*priority); err != nil {
		return false, err
	}
	return true, nil
}

// updateTicketCategory updates the ticket category if provided
func (s *ticketService) updateTicketCategory(ticket *tickets.Ticket, categoryID *uuid.UUID) bool {
	if categoryID == nil {
		return false
	}
	ticket.SetCategory(categoryID)
	return true
}

// updateTicketAssignee updates the ticket assignee if provided
func (s *ticketService) updateTicketAssignee(ticket *tickets.Ticket, assigneeID *uuid.UUID) (bool, error) {
	if assigneeID == nil {
		return false, nil
	}
	if *assigneeID == uuid.Nil {
		ticket.Unassign()
	} else {
		if err := ticket.AssignTo(*assigneeID); err != nil {
			return false, err
		}
	}
	return true, nil
}

// DeleteTicket deletes a ticket by ID
func (s *ticketService) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	return s.ticketRepo.DeleteTicket(ctx, id)
}

// ListTickets returns a paginated list of tickets
func (s *ticketService) ListTickets(
	ctx context.Context,
	filter queries.TicketFilter,
) ([]*tickets.Ticket, int64, error) {
	ticketList, err := s.ticketRepo.ListTickets(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// For count, we return the length of the list
	// In a real implementation, we would have a separate count method
	count := int64(len(ticketList))

	return ticketList, count, nil
}

// UpdateTicketStatus updates a ticket's status with validation
func (s *ticketService) UpdateTicketStatus(
	ctx context.Context,
	id uuid.UUID,
	status tickets.Status,
) (*tickets.Ticket, error) {
	ticket, err := s.ticketRepo.UpdateTicket(ctx, id, func(ticket *tickets.Ticket) (bool, error) {
		if err := ticket.ChangeStatus(status); err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return ticket, nil
}

// AssignTicket assigns or unassigns a ticket to/from a user
func (s *ticketService) AssignTicket(
	ctx context.Context,
	id uuid.UUID,
	assignedToID *uuid.UUID,
) (*tickets.Ticket, error) {
	ticket, err := s.ticketRepo.UpdateTicket(ctx, id, func(ticket *tickets.Ticket) (bool, error) {
		if assignedToID == nil || *assignedToID == uuid.Nil {
			ticket.Unassign()
		} else {
			if err := ticket.AssignTo(*assignedToID); err != nil {
				return false, err
			}
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return ticket, nil
}

// AddComment adds a comment to a ticket
func (s *ticketService) AddComment(
	ctx context.Context,
	ticketID uuid.UUID,
	authorID uuid.UUID,
	req AddTicketCommentRequest,
) (*tickets.Comment, error) {
	// Validate content
	if req.Content == "" {
		return nil, errors.New("comment content is required")
	}

	// Get the ticket first to validate it exists
	ticket, err := s.ticketRepo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Add comment to the ticket (returns error, not comment object)
	err = ticket.AddComment(authorID, req.Content, req.IsInternal)
	if err != nil {
		return nil, err
	}

	// Update the ticket with the new comment
	_, err = s.ticketRepo.UpdateTicket(ctx, ticketID, func(_ *tickets.Ticket) (bool, error) {
		// The comment is already added to the ticket, just return true to save
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	// Return the last comment (the one we just added)
	comments := ticket.Comments()
	if len(comments) == 0 {
		return nil, errors.New("failed to add comment")
	}
	return &comments[len(comments)-1], nil
}

// GetTicketComments returns comments for a specific ticket
func (s *ticketService) GetTicketComments(
	ctx context.Context,
	ticketID uuid.UUID,
	_ queries.CommentFilter,
) ([]*tickets.Comment, int64, error) {
	// Get the ticket
	ticket, err := s.ticketRepo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, 0, err
	}

	// Get comments from the ticket
	comments := ticket.Comments()

	// Convert to pointer slice
	commentPtrs := make([]*tickets.Comment, len(comments))
	for i := range comments {
		commentPtrs[i] = &comments[i]
	}

	// Apply filtering if needed
	// For now, return all comments
	// In a real implementation, we would apply the filter

	return commentPtrs, int64(len(commentPtrs)), nil
}
