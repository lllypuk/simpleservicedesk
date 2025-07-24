package users_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	domain "simpleservicedesk/internal/domain/users"
)

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		role  domain.Role
		valid bool
	}{
		{domain.RoleCustomer, true},
		{domain.RoleAgent, true},
		{domain.RoleAdmin, true},
		{domain.Role("invalid"), false},
		{domain.Role(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			require.Equal(t, tt.valid, tt.role.IsValid())
		})
	}
}

func TestRole_Permissions(t *testing.T) {
	tests := []struct {
		role                      domain.Role
		canCreateTickets          bool
		canAssignTickets          bool
		canViewAllTickets         bool
		canManageUsers            bool
		canViewInternalComments   bool
		canCreateInternalComments bool
	}{
		{domain.RoleCustomer, true, false, false, false, false, false},
		{domain.RoleAgent, true, true, true, false, true, true},
		{domain.RoleAdmin, true, true, true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			require.Equal(t, tt.canCreateTickets, tt.role.CanCreateTickets())
			require.Equal(t, tt.canAssignTickets, tt.role.CanAssignTickets())
			require.Equal(t, tt.canViewAllTickets, tt.role.CanViewAllTickets())
			require.Equal(t, tt.canManageUsers, tt.role.CanManageUsers())
			require.Equal(t, tt.canViewInternalComments, tt.role.CanViewInternalComments())
			require.Equal(t, tt.canCreateInternalComments, tt.role.CanCreateInternalComments())
		})
	}
}

func TestRole_Level(t *testing.T) {
	tests := []struct {
		role  domain.Role
		level int
	}{
		{domain.RoleCustomer, 1},
		{domain.RoleAgent, 2},
		{domain.RoleAdmin, 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			require.Equal(t, tt.level, tt.role.Level())
		})
	}
}

func TestRole_HasHigherOrEqualLevel(t *testing.T) {
	tests := []struct {
		role     domain.Role
		other    domain.Role
		hasLevel bool
	}{
		{domain.RoleAdmin, domain.RoleCustomer, true},
		{domain.RoleAdmin, domain.RoleAgent, true},
		{domain.RoleAdmin, domain.RoleAdmin, true},
		{domain.RoleAgent, domain.RoleCustomer, true},
		{domain.RoleAgent, domain.RoleAgent, true},
		{domain.RoleAgent, domain.RoleAdmin, false},
		{domain.RoleCustomer, domain.RoleCustomer, true},
		{domain.RoleCustomer, domain.RoleAgent, false},
		{domain.RoleCustomer, domain.RoleAdmin, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role)+"_vs_"+string(tt.other), func(t *testing.T) {
			require.Equal(t, tt.hasLevel, tt.role.HasHigherOrEqualLevel(tt.other))
		})
	}
}

func TestParseRole(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.Role
		hasError bool
	}{
		{"customer", domain.RoleCustomer, false},
		{"AGENT", domain.RoleAgent, false},
		{"  admin  ", domain.RoleAdmin, false},
		{"invalid", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := domain.ParseRole(tt.input)
			if tt.hasError {
				require.Error(t, err)
				require.Equal(t, domain.ErrInvalidRole, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
