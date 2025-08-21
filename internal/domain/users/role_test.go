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

func TestRole_String(t *testing.T) {
	tests := []struct {
		role     domain.Role
		expected string
	}{
		{domain.RoleCustomer, "customer"},
		{domain.RoleAgent, "agent"},
		{domain.RoleAdmin, "admin"},
		{domain.Role("invalid"), "invalid"},
		{domain.Role(""), ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			require.Equal(t, tt.expected, tt.role.String())
		})
	}
}

func TestRole_DisplayName(t *testing.T) {
	tests := []struct {
		role     domain.Role
		expected string
	}{
		{domain.RoleCustomer, "Клиент"},
		{domain.RoleAgent, "Агент"},
		{domain.RoleAdmin, "Администратор"},
		{domain.Role("invalid"), "invalid"},
		{domain.Role(""), ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			require.Equal(t, tt.expected, tt.role.DisplayName())
		})
	}
}

func TestRole_CanManageOrganization(t *testing.T) {
	tests := []struct {
		role     domain.Role
		expected bool
	}{
		{domain.RoleCustomer, false},
		{domain.RoleAgent, false},
		{domain.RoleAdmin, true},
		{domain.Role("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			require.Equal(t, tt.expected, tt.role.CanManageOrganization())
		})
	}
}

func TestAllRoles(t *testing.T) {
	roles := domain.AllRoles()

	require.Len(t, roles, 3)
	require.Contains(t, roles, domain.RoleCustomer)
	require.Contains(t, roles, domain.RoleAgent)
	require.Contains(t, roles, domain.RoleAdmin)

	// Verify order is preserved
	require.Equal(t, domain.RoleCustomer, roles[0])
	require.Equal(t, domain.RoleAgent, roles[1])
	require.Equal(t, domain.RoleAdmin, roles[2])
}

func TestRole_Level_EdgeCases(t *testing.T) {
	tests := []struct {
		role     domain.Role
		expected int
	}{
		{domain.Role("invalid"), 0},
		{domain.Role(""), 0},
		{domain.Role("CUSTOMER"), 0}, // case sensitive
		{domain.Role("mixed_Case"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			require.Equal(t, tt.expected, tt.role.Level())
		})
	}
}

func TestRole_PermissionConsistency(t *testing.T) {
	// Test that higher level roles have at least the same permissions as lower level roles

	// Customer permissions
	customerPerms := map[string]bool{
		"CanCreateTickets": domain.RoleCustomer.CanCreateTickets(),
	}

	// Agent should have all customer permissions plus more
	require.True(t, domain.RoleAgent.CanCreateTickets(), "Agent should inherit customer permissions")
	require.True(t, domain.RoleAgent.CanAssignTickets(), "Agent should have additional permissions")
	require.True(t, domain.RoleAgent.CanViewAllTickets())
	require.True(t, domain.RoleAgent.CanViewInternalComments())
	require.True(t, domain.RoleAgent.CanCreateInternalComments())

	// Admin should have all agent permissions plus more
	require.True(t, domain.RoleAdmin.CanCreateTickets(), "Admin should inherit all permissions")
	require.True(t, domain.RoleAdmin.CanAssignTickets())
	require.True(t, domain.RoleAdmin.CanViewAllTickets())
	require.True(t, domain.RoleAdmin.CanViewInternalComments())
	require.True(t, domain.RoleAdmin.CanCreateInternalComments())
	require.True(t, domain.RoleAdmin.CanManageUsers(), "Admin should have exclusive permissions")
	require.True(t, domain.RoleAdmin.CanManageOrganization())

	// Verify hierarchical consistency for create tickets
	for _, role := range domain.AllRoles() {
		if customerPerms["CanCreateTickets"] {
			require.True(t, role.CanCreateTickets(), "All roles should be able to create tickets")
		}
	}
}

func TestRole_InvalidRolePermissions(t *testing.T) {
	invalidRole := domain.Role("invalid")

	// Invalid roles should have no permissions
	require.False(t, invalidRole.CanCreateTickets())
	require.False(t, invalidRole.CanAssignTickets())
	require.False(t, invalidRole.CanViewAllTickets())
	require.False(t, invalidRole.CanManageUsers())
	require.False(t, invalidRole.CanManageOrganization())
	require.False(t, invalidRole.CanViewInternalComments())
	require.False(t, invalidRole.CanCreateInternalComments())

	// Invalid roles should have level 0
	require.Equal(t, 0, invalidRole.Level())
	require.False(t, invalidRole.IsValid())
	require.False(t, invalidRole.HasHigherOrEqualLevel(domain.RoleCustomer))
}

func TestParseRole_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected domain.Role
		hasError bool
	}{
		{"whitespace only", "   ", "", true},
		{"tabs and spaces", "\t  customer  \t", domain.RoleCustomer, false},
		{"mixed case", "Customer", domain.RoleCustomer, false},
		{"all caps", "CUSTOMER", domain.RoleCustomer, false},
		{"newlines", "agent\n", domain.RoleAgent, false},
		{"special characters", "admin!", "", true},
		{"partial match", "custom", "", true},
		{"extra text", "customer_role", "", true},
		{"unicode spaces", "\u00A0admin\u00A0", domain.RoleAdmin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := domain.ParseRole(tt.input)
			if tt.hasError {
				require.Error(t, err)
				require.Equal(t, domain.ErrInvalidRole, err)
				require.Equal(t, domain.Role(""), result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
