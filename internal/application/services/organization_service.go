package services

import (
	"context"
	"errors"

	"simpleservicedesk/internal/domain/organizations"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/interfaces"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
)

// organizationService implements OrganizationService interface
type organizationService struct {
	organizationRepo interfaces.OrganizationRepository
	userRepo         interfaces.UserRepository
	ticketRepo       interfaces.TicketRepository
}

// NewOrganizationService creates a new OrganizationService implementation
func NewOrganizationService(
	organizationRepo interfaces.OrganizationRepository,
	userRepo interfaces.UserRepository,
	ticketRepo interfaces.TicketRepository,
) OrganizationService {
	return &organizationService{
		organizationRepo: organizationRepo,
		userRepo:         userRepo,
		ticketRepo:       ticketRepo,
	}
}

// CreateOrganization creates a new organization with business logic validation
func (s *organizationService) CreateOrganization(
	ctx context.Context,
	req CreateOrganizationRequest,
) (*organizations.Organization, error) {
	// Validate required fields
	if req.Name == "" {
		return nil, errors.New("organization name is required")
	}
	if req.Domain == "" {
		return nil, errors.New("organization domain is required")
	}

	// Create organization through repository
	organization, err := s.organizationRepo.CreateOrganization(ctx, func() (*organizations.Organization, error) {
		org, err := organizations.CreateOrganization(req.Name, req.Domain)
		if err != nil {
			return nil, err
		}

		// Set parent if provided
		if req.ParentID != nil {
			if parentErr := org.ChangeParent(req.ParentID); parentErr != nil {
				return nil, parentErr
			}
		}

		return org, nil
	})
	if err != nil {
		return nil, err
	}

	return organization, nil
}

// GetOrganization retrieves an organization by ID
func (s *organizationService) GetOrganization(ctx context.Context, id uuid.UUID) (*organizations.Organization, error) {
	return s.organizationRepo.GetOrganization(ctx, id)
}

// UpdateOrganization updates organization information with business logic
func (s *organizationService) UpdateOrganization(
	ctx context.Context,
	id uuid.UUID,
	req UpdateOrganizationRequest,
) (*organizations.Organization, error) {
	organization, err := s.organizationRepo.UpdateOrganization(
		ctx,
		id,
		func(organization *organizations.Organization) (bool, error) {
			updated := false

			// Update common entity fields (name, parent, active status)
			if commonUpdated, err := updateEntityFields(
				organization,
				req.Name,
				req.ParentID,
				req.IsActive,
				organization.ChangeName,
				organization.ChangeParent,
			); err != nil {
				return false, err
			} else if commonUpdated {
				updated = true
			}

			// Update domain if provided (specific to organization)
			if domainUpdated, err := updateStringField(req.Domain, organization.ChangeDomain); err != nil {
				return false, err
			} else if domainUpdated {
				updated = true
			}

			return updated, nil
		})

	if err != nil {
		return nil, err
	}

	return organization, nil
}

// DeleteOrganization deletes an organization by ID
func (s *organizationService) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	return s.organizationRepo.DeleteOrganization(ctx, id)
}

// ListOrganizations returns a paginated list of organizations
func (s *organizationService) ListOrganizations(
	ctx context.Context,
	filter queries.OrganizationFilter,
) ([]*organizations.Organization, int64, error) {
	organizationList, err := s.organizationRepo.ListOrganizations(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// For count, we return the length of the list
	// In a real implementation, we would have a separate count method
	count := int64(len(organizationList))

	return organizationList, count, nil
}

// GetOrganizationUsers returns users for a specific organization
func (s *organizationService) GetOrganizationUsers(
	ctx context.Context,
	orgID uuid.UUID,
	filter queries.UserFilter,
) ([]*users.User, int64, error) {
	// Set the organization filter
	filter.OrganizationID = &orgID

	userList, err := s.userRepo.ListUsers(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.userRepo.CountUsers(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return userList, count, nil
}

// GetOrganizationTickets returns tickets for a specific organization
func (s *organizationService) GetOrganizationTickets(
	ctx context.Context,
	orgID uuid.UUID,
	filter queries.TicketFilter,
) ([]*tickets.Ticket, int64, error) {
	// First, check if the organization exists
	_, err := s.organizationRepo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, 0, err
	}

	// Set the organization filter
	filter.OrganizationID = &orgID

	ticketList, err := s.ticketRepo.ListTickets(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// For count, we return the length of the list
	// In a real implementation, we would have a separate count method
	count := int64(len(ticketList))

	return ticketList, count, nil
}
