package users_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "simpleservicedesk/internal/domain/users"
)

func TestNewUser_Valid(t *testing.T) {
	id := uuid.New()
	name := "Alice"
	email := "alice@example.com"
	user, err := domain.NewUser(id, name, email)
	require.NoError(t, err)
	require.Equal(t, id, user.ID())
	require.Equal(t, name, user.Name())
	require.Equal(t, email, user.Email())
}

func TestNewUser_InvalidName(t *testing.T) {
	id := uuid.New()
	_, err := domain.NewUser(id, "", "test@example.com")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestNewUser_InvalidEmail(t *testing.T) {
	id := uuid.New()
	_, err := domain.NewUser(id, "Bob", "")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestCreateUser(t *testing.T) {
	name := "Carol"
	email := "carol@example.com"
	user, err := domain.CreateUser(name, email)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, user.ID())
	require.Equal(t, name, user.Name())
	require.Equal(t, email, user.Email())
}

func TestSendToEmail_NotImplemented(t *testing.T) {
	user, err := domain.CreateUser("Dave", "dave@example.com")
	require.NoError(t, err)
	err = user.SendToEmail("anything")
	require.Error(t, err)
	require.EqualError(t, err, "not implemented")
}

func TestChangeEmail_Valid(t *testing.T) {
	user, err := domain.CreateUser("Eve", "eve@example.com")
	require.NoError(t, err)
	newEmail := "eve.new@example.com"
	err = user.ChangeEmail(newEmail)
	require.NoError(t, err)
	require.Equal(t, newEmail, user.Email())
}

func TestChangeEmail_Invalid(t *testing.T) {
	user, err := domain.CreateUser("Frank", "frank@example.com")
	require.NoError(t, err)
	originalEmail := user.Email()
	err = user.ChangeEmail("")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
	require.Equal(t, originalEmail, user.Email())
}
