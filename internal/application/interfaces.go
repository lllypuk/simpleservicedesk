package application

import (
	"simpleservicedesk/internal/interfaces"
)

// UserRepository re-exports the user repository interface for backward compatibility
type UserRepository = interfaces.UserRepository

// TicketRepository re-exports the ticket repository interface for backward compatibility
type TicketRepository = interfaces.TicketRepository

// CategoryRepository re-exports the category repository interface for backward compatibility
type CategoryRepository = interfaces.CategoryRepository

// OrganizationRepository re-exports the organization repository interface for backward compatibility
type OrganizationRepository = interfaces.OrganizationRepository

// CategoryTree re-exports the category tree structure for backward compatibility
type CategoryTree = interfaces.CategoryTree
type OrganizationTree = interfaces.OrganizationTree
