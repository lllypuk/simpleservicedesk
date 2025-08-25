//go:build !integration
// +build !integration

package queries_test

import (
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/queries"
)

func TestFromOpenAPITicketParams(t *testing.T) {
	t.Run("successful conversion with all params", func(t *testing.T) {
		status := openapi.New
		priority := openapi.High
		page := 2
		limit := 50

		params := openapi.GetTicketsParams{
			Status:   &status,
			Priority: &priority,
			Page:     &page,
			Limit:    &limit,
		}

		filter, err := queries.FromOpenAPITicketParams(params)

		require.NoError(t, err)
		assert.NotNil(t, filter.Status)
		assert.NotNil(t, filter.Priority)
		assert.Equal(t, 50, filter.Limit)
		assert.Equal(t, 50, filter.Offset) // (2-1) * 50
	})

	t.Run("successful conversion with default values", func(t *testing.T) {
		params := openapi.GetTicketsParams{}

		filter, err := queries.FromOpenAPITicketParams(params)

		require.NoError(t, err)
		assert.Nil(t, filter.Status)
		assert.Nil(t, filter.Priority)
		assert.Equal(t, 20, filter.Limit) // Default
		assert.Equal(t, 0, filter.Offset) // No page specified
	})

	t.Run("invalid status returns error", func(t *testing.T) {
		invalidStatus := openapi.TicketStatus("invalid_status")
		params := openapi.GetTicketsParams{
			Status: &invalidStatus,
		}

		_, err := queries.FromOpenAPITicketParams(params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})

	t.Run("invalid priority returns error", func(t *testing.T) {
		invalidPriority := openapi.TicketPriority("invalid_priority")
		params := openapi.GetTicketsParams{
			Priority: &invalidPriority,
		}

		_, err := queries.FromOpenAPITicketParams(params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid priority")
	})
}

func TestFromOpenAPICategoryParams(t *testing.T) {
	t.Run("successful conversion", func(t *testing.T) {
		orgIDValue := openapi_types.UUID{}
		parentIDValue := openapi_types.UUID{}
		isActive := true

		params := openapi.GetCategoriesParams{
			OrganizationId: &orgIDValue,
			ParentId:       &parentIDValue,
			IsActive:       &isActive,
		}

		filter, err := queries.FromOpenAPICategoryParams(params)

		require.NoError(t, err)
		assert.NotNil(t, filter.OrganizationID)
		assert.NotNil(t, filter.ParentID)
		assert.NotNil(t, filter.IsActive)
		assert.True(t, *filter.IsActive)
		assert.Equal(t, 20, filter.Limit)
		assert.Equal(t, "name", filter.SortBy)
		assert.Equal(t, "asc", filter.SortOrder)
	})
}

func TestFromOpenAPIOrganizationParams(t *testing.T) {
	t.Run("successful conversion", func(t *testing.T) {
		name := "Test Org"
		domain := "test.com"
		isActive := false
		page := 1
		limit := 30

		params := openapi.GetOrganizationsParams{
			Name:     &name,
			Domain:   &domain,
			IsActive: &isActive,
			Page:     &page,
			Limit:    &limit,
		}

		filter, err := queries.FromOpenAPIOrganizationParams(params)

		require.NoError(t, err)
		assert.NotNil(t, filter.Name)
		assert.Equal(t, "Test Org", *filter.Name)
		assert.NotNil(t, filter.Domain)
		assert.Equal(t, "test.com", *filter.Domain)
		assert.NotNil(t, filter.IsActive)
		assert.False(t, *filter.IsActive)
		assert.Equal(t, 30, filter.Limit)
		assert.Equal(t, 0, filter.Offset) // Page 1
	})
}

func TestFromOpenAPIUserParams(t *testing.T) {
	t.Run("successful conversion", func(t *testing.T) {
		name := "John Doe"
		email := "john@example.com"
		role := openapi.Admin
		isActive := true
		page := 3
		limit := 10

		params := openapi.GetUsersParams{
			Name:     &name,
			Email:    &email,
			Role:     &role,
			IsActive: &isActive,
			Page:     &page,
			Limit:    &limit,
		}

		filter, err := queries.FromOpenAPIUserParams(params)

		require.NoError(t, err)
		assert.NotNil(t, filter.Name)
		assert.Equal(t, "John Doe", *filter.Name)
		assert.NotNil(t, filter.Email)
		assert.Equal(t, "john@example.com", *filter.Email)
		assert.NotNil(t, filter.Role)
		assert.Equal(t, "admin", *filter.Role)
		assert.NotNil(t, filter.IsActive)
		assert.True(t, *filter.IsActive)
		assert.Equal(t, 10, filter.Limit)
		assert.Equal(t, 20, filter.Offset) // (3-1) * 10
	})
}
