package api

// Common error messages used across API handlers
const (
	MsgInvalidRequestFormat = "invalid request format"
	MsgInternalServerError  = "internal server error"
	MsgTicketNotFound       = "ticket not found"
	MsgCategoryNotFound     = "category not found"
	MsgOrganizationNotFound = "organization not found"
	MsgUserNotFound         = "user not found"
)

// Common default values
const (
	DefaultLimit = 50
)
