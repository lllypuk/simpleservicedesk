//go:build integration
// +build integration

package shared

import (
	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/organizations"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// TestUserData provides common test user data
type TestUserData struct {
	Name     string
	Email    string
	Password string
}

// Common test users
var (
	TestUser1 = TestUserData{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "password123",
	}

	TestUser2 = TestUserData{
		Name:     "Jane Smith",
		Email:    "jane.smith@example.com",
		Password: "password456",
	}

	TestUser3 = TestUserData{
		Name:     "Bob Wilson",
		Email:    "bob.wilson@example.com",
		Password: "password789",
	}
)

// CreateUserRequest creates an OpenAPI user request
func (td TestUserData) CreateUserRequest() openapi.CreateUserRequest {
	return openapi.CreateUserRequest{
		Name:     td.Name,
		Email:    openapi_types.Email(td.Email),
		Password: td.Password,
	}
}

// CreateDomainUser creates a domain user entity
func (td TestUserData) CreateDomainUser() (*users.User, error) {
	return users.NewUser(uuid.New(), td.Name, td.Email, []byte(td.Password))
}

// TestTicketData provides common test ticket data
type TestTicketData struct {
	Title          string
	Description    string
	Priority       tickets.Priority
	OrganizationID uuid.UUID
	AuthorID       uuid.UUID
	CategoryID     *uuid.UUID
}

// Common test tickets
func NewTestTicket1(orgID, authorID uuid.UUID) TestTicketData {
	return TestTicketData{
		Title:          "Test Ticket 1",
		Description:    "Description for test ticket 1",
		Priority:       tickets.PriorityNormal,
		OrganizationID: orgID,
		AuthorID:       authorID,
		CategoryID:     nil,
	}
}

func NewTestTicket2(orgID, authorID uuid.UUID) TestTicketData {
	return TestTicketData{
		Title:          "High Priority Ticket",
		Description:    "This is a high priority test ticket",
		Priority:       tickets.PriorityHigh,
		OrganizationID: orgID,
		AuthorID:       authorID,
		CategoryID:     nil,
	}
}

func NewTestTicket3(orgID, authorID uuid.UUID) TestTicketData {
	return TestTicketData{
		Title:          "Critical Issue",
		Description:    "This is a critical test ticket",
		Priority:       tickets.PriorityCritical,
		OrganizationID: orgID,
		AuthorID:       authorID,
		CategoryID:     nil,
	}
}

// CreateDomainTicket creates a domain ticket entity
func (td TestTicketData) CreateDomainTicket() (*tickets.Ticket, error) {
	return tickets.NewTicket(
		uuid.New(),
		td.Title,
		td.Description,
		td.Priority,
		td.OrganizationID,
		td.AuthorID,
		td.CategoryID,
	)
}

// TestOrganizationData provides common test organization data
type TestOrganizationData struct {
	Name     string
	Domain   string
	ParentID *uuid.UUID
}

// Common test organizations
var (
	TestOrg1 = TestOrganizationData{
		Name:     "Example Corp",
		Domain:   "example.com",
		ParentID: nil,
	}

	TestOrg2 = TestOrganizationData{
		Name:     "Tech Solutions Inc",
		Domain:   "techsolutions.net",
		ParentID: nil,
	}

	TestOrg3 = TestOrganizationData{
		Name:     "Innovation Labs",
		Domain:   "innovationlabs.org",
		ParentID: nil,
	}
)

// NewSubOrganization creates test data for a sub-organization
func NewSubOrganization(name, domain string, parentID uuid.UUID) TestOrganizationData {
	return TestOrganizationData{
		Name:     name,
		Domain:   domain,
		ParentID: &parentID,
	}
}

// CreateOrganizationRequest creates an OpenAPI organization request
func (td TestOrganizationData) CreateOrganizationRequest() openapi.CreateOrganizationRequest {
	return openapi.CreateOrganizationRequest{
		Name:     td.Name,
		Domain:   &td.Domain,
		ParentId: td.ParentID,
	}
}

// CreateDomainOrganization creates a domain organization entity
func (td TestOrganizationData) CreateDomainOrganization() (*organizations.Organization, error) {
	if td.ParentID == nil {
		return organizations.CreateRootOrganization(td.Name, td.Domain)
	}
	return organizations.CreateSubOrganization(td.Name, td.Domain, *td.ParentID)
}
