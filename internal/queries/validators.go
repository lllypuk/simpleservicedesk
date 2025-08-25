package queries

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	// Sort order constants
	sortOrderAsc  = "asc"
	sortOrderDesc = "desc"

	// Default sort field
	defaultSortByName = "name"

	// Validation limits
	maxLimit = 1000
)

// Validate checks TicketFilter for business rule compliance
func (f TicketFilter) Validate() error {
	if err := f.BaseFilter.Validate(); err != nil {
		return fmt.Errorf("base filter validation: %w", err)
	}

	if err := f.TimeRangeFilter.Validate(); err != nil {
		return fmt.Errorf("time range validation: %w", err)
	}

	// Ticket-specific validation rules
	// Status and Priority are validated during parsing, so no additional checks needed

	return nil
}

// Validate checks CategoryFilter for business rule compliance
func (f CategoryFilter) Validate() error {
	if err := f.BaseFilter.Validate(); err != nil {
		return fmt.Errorf("base filter validation: %w", err)
	}

	// Category-specific validation
	if f.Name != nil && strings.TrimSpace(*f.Name) == "" {
		return errors.New("category name filter cannot be empty")
	}

	return nil
}

// Validate checks OrganizationFilter for business rule compliance
func (f OrganizationFilter) Validate() error {
	if err := f.BaseFilter.Validate(); err != nil {
		return fmt.Errorf("base filter validation: %w", err)
	}

	// Organization-specific validation
	if f.Name != nil && strings.TrimSpace(*f.Name) == "" {
		return errors.New("organization name filter cannot be empty")
	}

	if f.Domain != nil && strings.TrimSpace(*f.Domain) == "" {
		return errors.New("organization domain filter cannot be empty")
	}

	return nil
}

// Validate checks UserFilter for business rule compliance
func (f UserFilter) Validate() error {
	if err := f.BaseFilter.Validate(); err != nil {
		return fmt.Errorf("base filter validation: %w", err)
	}

	// User-specific validation
	if f.Name != nil && strings.TrimSpace(*f.Name) == "" {
		return errors.New("user name filter cannot be empty")
	}

	if f.Email != nil && strings.TrimSpace(*f.Email) == "" {
		return errors.New("user email filter cannot be empty")
	}

	if f.Role != nil {
		validRoles := []string{"admin", "agent", "customer"}
		if !contains(validRoles, strings.ToLower(*f.Role)) {
			return fmt.Errorf("invalid user role: %s (must be one of: %s)", *f.Role, strings.Join(validRoles, ", "))
		}
	}

	return nil
}

// Validate checks BaseFilter for common validation rules
func (f BaseFilter) Validate() error {
	if f.Limit < 0 {
		return fmt.Errorf("limit must be non-negative, got: %d", f.Limit)
	}
	if f.Limit > maxLimit {
		return fmt.Errorf("limit too large, maximum: %d, got: %d", maxLimit, f.Limit)
	}
	if f.Offset < 0 {
		return fmt.Errorf("offset must be non-negative, got: %d", f.Offset)
	}

	// Validate sort fields
	if f.SortBy != "" {
		validSortFields := []string{
			"created_at", "updated_at", "status", "priority",
			"name", "title", "email", "domain", "id",
		}
		if !contains(validSortFields, f.SortBy) {
			return fmt.Errorf("invalid sort field: %s", f.SortBy)
		}
	}

	if f.SortOrder != "" && f.SortOrder != sortOrderAsc && f.SortOrder != sortOrderDesc {
		return fmt.Errorf("invalid sort order: %s (must be '%s' or '%s')", f.SortOrder, sortOrderAsc, sortOrderDesc)
	}

	return nil
}

// Validate checks TimeRangeFilter for date logic
func (f TimeRangeFilter) Validate() error {
	now := time.Now()

	if f.CreatedAfter != nil && f.CreatedAfter.After(now) {
		return errors.New("created_after cannot be in the future")
	}

	if f.CreatedBefore != nil && f.CreatedBefore.After(now) {
		return errors.New("created_before cannot be in the future")
	}

	if f.CreatedAfter != nil && f.CreatedBefore != nil {
		if f.CreatedAfter.After(*f.CreatedBefore) {
			return errors.New("created_after must be before created_before")
		}

		// Check if date range is reasonable (not more than 10 years)
		if f.CreatedBefore.Sub(*f.CreatedAfter) > 10*365*24*time.Hour {
			return errors.New("date range too large (maximum 10 years)")
		}
	}

	if f.UpdatedAfter != nil && f.UpdatedAfter.After(now) {
		return errors.New("updated_after cannot be in the future")
	}

	if f.UpdatedBefore != nil && f.UpdatedBefore.After(now) {
		return errors.New("updated_before cannot be in the future")
	}

	if f.UpdatedAfter != nil && f.UpdatedBefore != nil {
		if f.UpdatedAfter.After(*f.UpdatedBefore) {
			return errors.New("updated_after must be before updated_before")
		}

		// Check if date range is reasonable (not more than 10 years)
		if f.UpdatedBefore.Sub(*f.UpdatedAfter) > 10*365*24*time.Hour {
			return errors.New("updated date range too large (maximum 10 years)")
		}
	}

	return nil
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidateAndSetDefaults validates the filter and sets sensible defaults
func (f TicketFilter) ValidateAndSetDefaults() (TicketFilter, error) {
	// Set defaults
	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}
	if f.SortOrder == "" {
		f.SortOrder = "desc"
	}

	return f, f.Validate()
}

// ValidateAndSetDefaults validates the filter and sets sensible defaults
func (f CategoryFilter) ValidateAndSetDefaults() (CategoryFilter, error) {
	// Set defaults
	if f.Limit == 0 {
		f.Limit = 50 // Categories usually have fewer items
	}
	if f.SortBy == "" {
		f.SortBy = defaultSortByName
	}
	if f.SortOrder == "" {
		f.SortOrder = sortOrderAsc
	}

	return f, f.Validate()
}

// ValidateAndSetDefaults validates the filter and sets sensible defaults
func (f OrganizationFilter) ValidateAndSetDefaults() (OrganizationFilter, error) {
	// Set defaults
	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.SortBy == "" {
		f.SortBy = defaultSortByName
	}
	if f.SortOrder == "" {
		f.SortOrder = sortOrderAsc
	}

	return f, f.Validate()
}

// ValidateAndSetDefaults validates the filter and sets sensible defaults
func (f UserFilter) ValidateAndSetDefaults() (UserFilter, error) {
	// Set defaults
	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.SortBy == "" {
		f.SortBy = defaultSortByName
	}
	if f.SortOrder == "" {
		f.SortOrder = sortOrderAsc
	}

	return f, f.Validate()
}
