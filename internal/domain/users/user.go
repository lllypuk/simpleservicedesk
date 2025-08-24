package users

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidUser      = errors.New("invalid user")
	ErrUserValidation   = errors.New("validation error")
	ErrUserAlreadyExist = errors.New("user already exist")
)

const MinPasswordLength = 6

type User struct {
	id             uuid.UUID
	name           string
	email          string
	passwordHash   []byte
	role           Role
	organizationID *uuid.UUID
	isActive       bool
	createdAt      time.Time
	updatedAt      time.Time
}

func NewUser(id uuid.UUID, name, email string, passwordHash []byte) (*User, error) {
	return NewUserWithDetails(id, name, email, passwordHash, RoleCustomer, nil, true, time.Now(), time.Now())
}

func NewUserWithDetails(
	id uuid.UUID,
	name, email string,
	passwordHash []byte,
	role Role,
	organizationID *uuid.UUID,
	isActive bool,
	createdAt, updatedAt time.Time,
) (*User, error) {
	if err := validateUsername(name); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if len(passwordHash) == 0 {
		return nil, fmt.Errorf("%w: password hash is required", ErrUserValidation)
	}
	if !role.IsValid() {
		return nil, fmt.Errorf("%w: invalid role", ErrUserValidation)
	}

	return &User{
		id:             id,
		name:           name,
		email:          email,
		passwordHash:   passwordHash,
		role:           role,
		organizationID: organizationID,
		isActive:       isActive,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
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

func (u *User) Role() Role {
	return u.role
}

func (u *User) OrganizationID() *uuid.UUID {
	return u.organizationID
}

func (u *User) IsActive() bool {
	return u.isActive
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) SendToEmail(_ string) error {
	return errors.New("not implemented")
}

func (u *User) ChangeEmail(email string) error {
	if err := validateEmail(email); err != nil {
		return err
	}
	u.email = email
	u.updatedAt = time.Now()
	return nil
}

func (u *User) ChangeName(name string) error {
	if err := validateUsername(name); err != nil {
		return err
	}
	u.name = name
	u.updatedAt = time.Now()
	return nil
}

func (u *User) ChangeRole(role Role) error {
	if !role.IsValid() {
		return fmt.Errorf("%w: invalid role", ErrUserValidation)
	}
	u.role = role
	u.updatedAt = time.Now()
	return nil
}

func (u *User) ChangeOrganization(organizationID *uuid.UUID) error {
	u.organizationID = organizationID
	u.updatedAt = time.Now()
	return nil
}

func (u *User) Activate() {
	u.isActive = true
	u.updatedAt = time.Now()
}

func (u *User) Deactivate() {
	u.isActive = false
	u.updatedAt = time.Now()
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.passwordHash, []byte(password))
	return err == nil
}

func (u *User) ChangePassword(newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.passwordHash = passwordHash
	u.updatedAt = time.Now()
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
	if len(password) < MinPasswordLength {
		return fmt.Errorf("%w: password must be at least %d characters long", ErrUserValidation, MinPasswordLength)
	}
	return nil
}
