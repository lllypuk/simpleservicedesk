//go:build !integration
// +build !integration

package queries_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"simpleservicedesk/internal/queries"
)

func TestBaseFilterValidate(t *testing.T) {
	tests := []struct {
		name        string
		filter      queries.BaseFilter
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid filter",
			filter: queries.BaseFilter{
				Limit:     20,
				Offset:    0,
				SortBy:    "created_at",
				SortOrder: "desc",
			},
			expectError: false,
		},
		{
			name: "negative limit",
			filter: queries.BaseFilter{
				Limit: -1,
			},
			expectError: true,
			errorMsg:    "limit must be non-negative",
		},
		{
			name: "limit too large",
			filter: queries.BaseFilter{
				Limit: 1001,
			},
			expectError: true,
			errorMsg:    "limit too large",
		},
		{
			name: "negative offset",
			filter: queries.BaseFilter{
				Offset: -1,
			},
			expectError: true,
			errorMsg:    "offset must be non-negative",
		},
		{
			name: "invalid sort field",
			filter: queries.BaseFilter{
				SortBy: "invalid_field",
			},
			expectError: true,
			errorMsg:    "invalid sort field",
		},
		{
			name: "invalid sort order",
			filter: queries.BaseFilter{
				SortOrder: "invalid_order",
			},
			expectError: true,
			errorMsg:    "invalid sort order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTimeRangeFilterValidate(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name        string
		filter      queries.TimeRangeFilter
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid time range",
			filter: queries.TimeRangeFilter{
				CreatedAfter:  &past,
				CreatedBefore: &now,
			},
			expectError: false,
		},
		{
			name: "created_after in future",
			filter: queries.TimeRangeFilter{
				CreatedAfter: &future,
			},
			expectError: true,
			errorMsg:    "created_after cannot be in the future",
		},
		{
			name: "created_before in future",
			filter: queries.TimeRangeFilter{
				CreatedBefore: &future,
			},
			expectError: true,
			errorMsg:    "created_before cannot be in the future",
		},
		{
			name: "created_after after created_before",
			filter: queries.TimeRangeFilter{
				CreatedAfter:  &now,
				CreatedBefore: &past,
			},
			expectError: true,
			errorMsg:    "created_after must be before created_before",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTicketFilterValidateAndSetDefaults(t *testing.T) {
	t.Run("sets defaults correctly", func(t *testing.T) {
		filter := queries.TicketFilter{}

		filter, err := filter.ValidateAndSetDefaults()

		require.NoError(t, err)
		assert.Equal(t, 20, filter.Limit)
		assert.Equal(t, "created_at", filter.SortBy)
		assert.Equal(t, "desc", filter.SortOrder)
	})

	t.Run("preserves existing values", func(t *testing.T) {
		filter := queries.TicketFilter{
			BaseFilter: queries.BaseFilter{
				Limit:     50,
				SortBy:    "priority",
				SortOrder: "asc",
			},
		}

		filter, err := filter.ValidateAndSetDefaults()

		require.NoError(t, err)
		assert.Equal(t, 50, filter.Limit)
		assert.Equal(t, "priority", filter.SortBy)
		assert.Equal(t, "asc", filter.SortOrder)
	})
}

func TestUserFilterValidate(t *testing.T) {
	tests := []struct {
		name        string
		filter      queries.UserFilter
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid user filter",
			filter: queries.UserFilter{
				BaseFilter: queries.BaseFilter{
					Limit: 20,
				},
				Name:  stringPtr("John Doe"),
				Email: stringPtr("john@example.com"),
				Role:  stringPtr("admin"),
			},
			expectError: false,
		},
		{
			name: "empty name filter",
			filter: queries.UserFilter{
				Name: stringPtr("   "),
			},
			expectError: true,
			errorMsg:    "user name filter cannot be empty",
		},
		{
			name: "empty email filter",
			filter: queries.UserFilter{
				Email: stringPtr(""),
			},
			expectError: true,
			errorMsg:    "user email filter cannot be empty",
		},
		{
			name: "invalid role",
			filter: queries.UserFilter{
				Role: stringPtr("invalid_role"),
			},
			expectError: true,
			errorMsg:    "invalid user role",
		},
		{
			name: "valid role uppercase",
			filter: queries.UserFilter{
				Role: stringPtr("ADMIN"),
			},
			expectError: false, // Should pass as validation converts to lowercase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCategoryFilterValidate(t *testing.T) {
	tests := []struct {
		name        string
		filter      queries.CategoryFilter
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid category filter",
			filter: queries.CategoryFilter{
				BaseFilter: queries.BaseFilter{
					Limit: 50,
				},
				Name:           stringPtr("Technology"),
				OrganizationID: &uuid.UUID{},
				IsActive:       boolPtr(true),
			},
			expectError: false,
		},
		{
			name: "empty name filter",
			filter: queries.CategoryFilter{
				Name: stringPtr("  "),
			},
			expectError: true,
			errorMsg:    "category name filter cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOrganizationFilterValidate(t *testing.T) {
	tests := []struct {
		name        string
		filter      queries.OrganizationFilter
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid organization filter",
			filter: queries.OrganizationFilter{
				BaseFilter: queries.BaseFilter{
					Limit: 30,
				},
				Name:     stringPtr("ACME Corp"),
				Domain:   stringPtr("acme.com"),
				IsActive: boolPtr(true),
			},
			expectError: false,
		},
		{
			name: "empty name filter",
			filter: queries.OrganizationFilter{
				Name: stringPtr(""),
			},
			expectError: true,
			errorMsg:    "organization name filter cannot be empty",
		},
		{
			name: "empty domain filter",
			filter: queries.OrganizationFilter{
				Domain: stringPtr("   "),
			},
			expectError: true,
			errorMsg:    "organization domain filter cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
