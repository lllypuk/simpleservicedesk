package users_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	domain "simpleservicedesk/internal/domain/users"
)

func TestNewUser_Valid(t *testing.T) {
	id := uuid.New()
	name := "Alice"
	email := "alice@example.com"
	password := "password123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.NewUser(id, name, email, passwordHash)
	require.NoError(t, err)
	require.Equal(t, id, user.ID())
	require.Equal(t, name, user.Name())
	require.Equal(t, email, user.Email())
	require.True(t, user.CheckPassword(password))
}

func TestNewUser_InvalidName(t *testing.T) {
	id := uuid.New()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_, err := domain.NewUser(id, "", "test@example.com", passwordHash)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestNewUser_InvalidEmail(t *testing.T) {
	id := uuid.New()
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_, err := domain.NewUser(id, "Bob", "", passwordHash)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestNewUser_InvalidPasswordHash_Empty(t *testing.T) {
	id := uuid.New()
	_, err := domain.NewUser(id, "Bob", "bob@example.com", []byte{})
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestNewUser_InvalidPasswordHash_Nil(t *testing.T) {
	id := uuid.New()
	_, err := domain.NewUser(id, "Bob", "bob@example.com", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)
}

func TestCreateUser_Valid(t *testing.T) {
	name := "Charlie"
	email := "charlie@example.com"
	password := "securepass"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.CreateUser(name, email, passwordHash)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, user.ID())
	require.Equal(t, name, user.Name())
	require.Equal(t, email, user.Email())
	require.True(t, user.CheckPassword(password))
}

func TestUser_CheckPassword(t *testing.T) {
	password := "mypassword"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.CreateUser("Dave", "dave@example.com", passwordHash)
	require.NoError(t, err)

	// Правильный пароль
	require.True(t, user.CheckPassword(password))

	// Неправильный пароль
	require.False(t, user.CheckPassword("wrongpassword"))
	require.False(t, user.CheckPassword(""))
}

func TestUser_ChangePassword(t *testing.T) {
	oldPassword := "oldpassword"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.CreateUser("Eve", "eve@example.com", passwordHash)
	require.NoError(t, err)

	// Проверяем, что старый пароль работает
	require.True(t, user.CheckPassword(oldPassword))

	// Меняем пароль
	newPassword := "newpassword123"
	err = user.ChangePassword(newPassword)
	require.NoError(t, err)

	// Проверяем, что новый пароль работает, а старый - нет
	require.True(t, user.CheckPassword(newPassword))
	require.False(t, user.CheckPassword(oldPassword))
}

func TestUser_ChangePassword_Invalid(t *testing.T) {
	password := "validpass"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := domain.CreateUser("Frank", "frank@example.com", passwordHash)
	require.NoError(t, err)

	// Попытка установить пустой пароль
	err = user.ChangePassword("")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)

	// Попытка установить слишком короткий пароль
	err = user.ChangePassword("123")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)

	// Проверяем, что старый пароль все еще работает
	require.True(t, user.CheckPassword(password))
}

func TestUser_ChangeEmail(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("Grace", "grace@example.com", passwordHash)
	require.NoError(t, err)

	newEmail := "grace.new@example.com"
	err = user.ChangeEmail(newEmail)
	require.NoError(t, err)
	require.Equal(t, newEmail, user.Email())
}

func TestUser_ChangeEmail_Invalid(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user, err := domain.CreateUser("Henry", "henry@example.com", passwordHash)
	require.NoError(t, err)

	err = user.ChangeEmail("")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUserValidation)

	// Проверяем, что email не изменился
	require.Equal(t, "henry@example.com", user.Email())
}
