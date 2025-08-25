package queries

import (
	"time"

	"github.com/google/uuid"

	"simpleservicedesk/internal/domain/tickets"
)

// BaseFilter contains common pagination and sorting fields
type BaseFilter struct {
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
}

// TimeRangeFilter contains common date filtering fields
type TimeRangeFilter struct {
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`
}

// TicketFilter - SINGLE source of truth for ticket filtering
type TicketFilter struct {
	BaseFilter
	TimeRangeFilter

	Status         *tickets.Status   `json:"status,omitempty"`
	Priority       *tickets.Priority `json:"priority,omitempty"`
	AssigneeID     *uuid.UUID        `json:"assignee_id,omitempty"`
	AuthorID       *uuid.UUID        `json:"author_id,omitempty"`
	OrganizationID *uuid.UUID        `json:"organization_id,omitempty"`
	CategoryID     *uuid.UUID        `json:"category_id,omitempty"`
	IsOverdue      *bool             `json:"is_overdue,omitempty"`
}

// CategoryFilter - SINGLE source of truth for category filtering
type CategoryFilter struct {
	BaseFilter

	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	Name           *string    `json:"name,omitempty"`
	IsRootOnly     bool       `json:"is_root_only,omitempty"`
}

// OrganizationFilter - SINGLE source of truth for organization filtering
type OrganizationFilter struct {
	BaseFilter

	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
	Name       *string    `json:"name,omitempty"`
	Domain     *string    `json:"domain,omitempty"`
	IsRootOnly bool       `json:"is_root_only,omitempty"`
}

// UserFilter - SINGLE source of truth for user filtering
type UserFilter struct {
	BaseFilter

	Name           *string    `json:"name,omitempty"`
	Email          *string    `json:"email,omitempty"`
	Role           *string    `json:"role,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
}
