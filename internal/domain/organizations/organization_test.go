package organizations_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domainOrg "simpleservicedesk/internal/domain/organizations"
)

func TestNewOrganization_Valid(t *testing.T) {
	id := uuid.New()
	name := "Test Organization"
	domain := "test.com"

	org, err := domainOrg.NewOrganization(id, name, domain)

	require.NoError(t, err)
	require.Equal(t, id, org.ID())
	require.Equal(t, name, org.Name())
	require.Equal(t, domain, org.Domain())
	require.True(t, org.IsActive())
	require.False(t, org.CreatedAt().IsZero())
	require.False(t, org.UpdatedAt().IsZero())
	require.Equal(t, domainOrg.DefaultSettings(), org.Settings())
}

func TestNewOrganization_InvalidName(t *testing.T) {
	tests := []struct {
		name     string
		orgName  string
		hasError bool
	}{
		{"empty name", "", true},
		{"too short", "A", true},
		{"valid short", "AB", false},
		{"valid long", "Test Organization Name", false},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domainOrg.NewOrganization(uuid.New(), tt.orgName, "test.com")
			if tt.hasError {
				require.Error(t, err)
				require.ErrorIs(t, err, domainOrg.ErrOrganizationValidation)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewOrganization_InvalidDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		hasError bool
	}{
		{"empty domain", "", false}, // домен может быть пустым
		{"valid domain", "test.com", false},
		{"subdomain", "mail.test.com", false},
		{"invalid no dot", "testcom", true},
		{"invalid chars", "test@.com", true},
		{"valid with dash", "test-org.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domainOrg.NewOrganization(uuid.New(), "Test Org", tt.domain)
			if tt.hasError {
				require.Error(t, err)
				require.ErrorIs(t, err, domainOrg.ErrOrganizationValidation)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateOrganization(t *testing.T) {
	name := "Test Organization"
	domain := "test.com"

	org, err := domainOrg.CreateOrganization(name, domain)

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, org.ID())
	require.Equal(t, name, org.Name())
	require.Equal(t, domain, org.Domain())
}

func TestOrganization_ChangeName(t *testing.T) {
	org, err := domainOrg.CreateOrganization("Original Name", "test.com")
	require.NoError(t, err)

	originalUpdatedAt := org.UpdatedAt()
	time.Sleep(time.Millisecond) // чтобы время обновления изменилось

	newName := "New Organization Name"
	err = org.ChangeName(newName)

	require.NoError(t, err)
	require.Equal(t, newName, org.Name())
	require.True(t, org.UpdatedAt().After(originalUpdatedAt))
}

func TestOrganization_ChangeName_Invalid(t *testing.T) {
	org, err := domainOrg.CreateOrganization("Original Name", "test.com")
	require.NoError(t, err)

	originalName := org.Name()
	originalUpdatedAt := org.UpdatedAt()

	err = org.ChangeName("")

	require.Error(t, err)
	require.ErrorIs(t, err, domainOrg.ErrOrganizationValidation)
	require.Equal(t, originalName, org.Name())           // имя не должно измениться
	require.Equal(t, originalUpdatedAt, org.UpdatedAt()) // время обновления не должно измениться
}

func TestOrganization_ChangeDomain(t *testing.T) {
	org, err := domainOrg.CreateOrganization("Test Org", "old.com")
	require.NoError(t, err)

	originalUpdatedAt := org.UpdatedAt()
	time.Sleep(time.Millisecond)

	newDomain := "new.com"
	err = org.ChangeDomain(newDomain)

	require.NoError(t, err)
	require.Equal(t, newDomain, org.Domain())
	require.True(t, org.UpdatedAt().After(originalUpdatedAt))
}

func TestOrganization_UpdateSettings(t *testing.T) {
	org, err := domainOrg.CreateOrganization("Test Org", "test.com")
	require.NoError(t, err)

	originalUpdatedAt := org.UpdatedAt()
	time.Sleep(time.Millisecond)

	newSettings := domainOrg.OrganizationSettings{
		AllowPublicTickets:    true,
		DefaultTicketPriority: "high",
		EmailNotifications:    false,
		MaxFileSize:           5 * 1024 * 1024,
	}

	org.UpdateSettings(newSettings)

	require.Equal(t, newSettings, org.Settings())
	require.True(t, org.UpdatedAt().After(originalUpdatedAt))
}

func TestOrganization_ActivateDeactivate(t *testing.T) {
	org, err := domainOrg.CreateOrganization("Test Org", "test.com")
	require.NoError(t, err)
	require.True(t, org.IsActive()) // по умолчанию активна

	originalUpdatedAt := org.UpdatedAt()
	time.Sleep(time.Millisecond)

	org.Deactivate()
	require.False(t, org.IsActive())
	require.True(t, org.UpdatedAt().After(originalUpdatedAt))

	deactivatedAt := org.UpdatedAt()
	time.Sleep(time.Millisecond)

	org.Activate()
	require.True(t, org.IsActive())
	require.True(t, org.UpdatedAt().After(deactivatedAt))
}

func TestOrganization_CanUserJoinByEmail(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		email   string
		canJoin bool
	}{
		{"matching domain", "test.com", "user@test.com", true},
		{"case insensitive", "Test.Com", "user@TEST.COM", true},
		{"different domain", "test.com", "user@other.com", false},
		{"no domain", "", "user@test.com", false},
		{"invalid email", "test.com", "invalid-email", false},
		{"subdomain", "test.com", "user@mail.test.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, err := domainOrg.CreateOrganization("Test Org", tt.domain)
			require.NoError(t, err)

			canJoin := org.CanUserJoinByEmail(tt.email)
			require.Equal(t, tt.canJoin, canJoin)
		})
	}
}

func TestDefaultSettings(t *testing.T) {
	settings := domainOrg.DefaultSettings()

	require.False(t, settings.AllowPublicTickets)
	require.Equal(t, "normal", settings.DefaultTicketPriority)
	require.True(t, settings.EmailNotifications)
	require.Equal(t, int64(10*1024*1024), settings.MaxFileSize)
}
