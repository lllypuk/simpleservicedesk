package queries

import (
	"fmt"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"
)

const (
	defaultLimit    = 20
	maxAllowedLimit = 100
)

// FromOpenAPITicketParams converts OpenAPI parameters to TicketFilter
func FromOpenAPITicketParams(params openapi.GetTicketsParams) (TicketFilter, error) {
	// Validate page parameter
	if params.Page != nil && *params.Page < 1 {
		return TicketFilter{}, fmt.Errorf("page must be positive, got: %d", *params.Page)
	}

	// Validate limit parameter
	if params.Limit != nil {
		if *params.Limit < 1 {
			return TicketFilter{}, fmt.Errorf("limit must be positive, got: %d", *params.Limit)
		}
		if *params.Limit > maxAllowedLimit {
			return TicketFilter{}, fmt.Errorf("limit too large, maximum: %d, got: %d", maxAllowedLimit, *params.Limit)
		}
	}

	filter := TicketFilter{
		BaseFilter: BaseFilter{
			Limit:     getIntValue(params.Limit),
			Offset:    calculateOffset(params.Page, params.Limit),
			SortBy:    "created_at", // Default sort
			SortOrder: "desc",       // Default order
		},
	}

	// Convert status with validation
	if params.Status != nil {
		status, err := tickets.ParseStatus(string(*params.Status))
		if err != nil {
			return filter, fmt.Errorf("invalid status: %w", err)
		}
		filter.Status = &status
	}

	// Convert priority with validation
	if params.Priority != nil {
		priority, err := tickets.ParsePriority(string(*params.Priority))
		if err != nil {
			return filter, fmt.Errorf("invalid priority: %w", err)
		}
		filter.Priority = &priority
	}

	// Direct UUID mappings
	filter.AssigneeID = params.AssigneeId
	filter.AuthorID = params.AuthorId
	filter.OrganizationID = params.OrganizationId
	filter.CategoryID = params.CategoryId

	return filter, nil
}

// FromOpenAPICategoryParams converts OpenAPI parameters to CategoryFilter
func FromOpenAPICategoryParams(params openapi.GetCategoriesParams) (CategoryFilter, error) {
	filter := CategoryFilter{
		BaseFilter: BaseFilter{
			Limit:     defaultLimit, // Default limit for categories
			Offset:    0,            // No pagination in current params
			SortBy:    "name",
			SortOrder: "asc",
		},
	}

	// Direct field mappings
	filter.OrganizationID = params.OrganizationId
	filter.ParentID = params.ParentId
	filter.IsActive = params.IsActive

	return filter, nil
}

// FromOpenAPIOrganizationParams converts OpenAPI parameters to OrganizationFilter
func FromOpenAPIOrganizationParams(params openapi.GetOrganizationsParams) (OrganizationFilter, error) {
	filter := OrganizationFilter{
		BaseFilter: BaseFilter{
			Limit:     getIntValue(params.Limit),
			Offset:    calculateOffset(params.Page, params.Limit),
			SortBy:    "name",
			SortOrder: "asc",
		},
	}

	// Direct field mappings
	filter.Name = params.Name
	filter.Domain = params.Domain
	filter.IsActive = params.IsActive
	filter.ParentID = params.ParentId

	return filter, nil
}

// FromOpenAPIUserParams converts OpenAPI parameters to UserFilter
func FromOpenAPIUserParams(params openapi.GetUsersParams) (UserFilter, error) {
	filter := UserFilter{
		BaseFilter: BaseFilter{
			Limit:     getIntValue(params.Limit),
			Offset:    calculateOffset(params.Page, params.Limit),
			SortBy:    "name",
			SortOrder: "asc",
		},
	}

	// Direct field mappings
	filter.Name = params.Name
	filter.Email = params.Email
	filter.IsActive = params.IsActive
	filter.OrganizationID = params.OrganizationId

	// Convert role if present
	if params.Role != nil {
		roleStr := string(*params.Role)
		filter.Role = &roleStr
	}

	return filter, nil
}

// Helper functions for safe pointer dereferencing and type conversions

func getIntValue(ptr *int) int {
	if ptr == nil {
		return defaultLimit // Default limit
	}
	return *ptr
}

// calculateOffset converts page/limit to offset for database queries
func calculateOffset(page, limit *int) int {
	if page == nil || limit == nil {
		return 0
	}
	if *page <= 1 {
		return 0
	}
	return (*page - 1) * (*limit)
}
