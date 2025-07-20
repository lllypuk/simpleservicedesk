package users

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidUser      = errors.New("invalid user")
	ErrUserValidation   = errors.New("validation error")
	ErrUserAlreadyExist = errors.New("user already exist")
)

type User struct {
	id           uuid.UUID
	name         string
	email        string
	passwordHash []byte
}

func NewUser(id uuid.UUID, name, email string, passwordHash []byte) (*User, error) {
	if err := validateUsername(name); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if len(passwordHash) == 0 {
		return nil, fmt.Errorf("%w: password hash is required", ErrUserValidation)
	}

	return &User{
		id:           id,
		name:         name,
		email:        email,
		passwordHash: passwordHash,
	}, nil
}

func CreateUser(name, email string, passwordHash []byte) (*User, error) {
	return NewUser(uuid.New(), name, email, passwordHash)
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Email() string {
	return u.email
}

func (u *User) SendToEmail(_ string) error {
	return errors.New("not implemented")
}

func (u *User) ChangeEmail(email string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	u.email = email
	return nil
}

// CheckPassword проверяет, соответствует ли предоставленный пароль хешу пароля пользователя
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.passwordHash, []byte(password))
	return err == nil
}

// ChangePassword изменяет пароль пользователя
func (u *User) ChangePassword(newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.passwordHash = passwordHash
	return nil
}

func validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("%w: name is required", ErrUserValidation)
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("%w: email is required", ErrUserValidation)
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("%w: password is required", ErrUserValidation)
	}
	if len(password) < 6 {
		return fmt.Errorf("%w: password must be at least 6 characters long", ErrUserValidation)
	}
	return nil
}
