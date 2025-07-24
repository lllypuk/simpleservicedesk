package organizations

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrganizationNotFound     = errors.New("organization not found")
	ErrInvalidOrganization      = errors.New("invalid organization")
	ErrOrganizationValidation   = errors.New("organization validation error")
	ErrOrganizationAlreadyExist = errors.New("organization already exist")
)

const (
	MinNameLength        = 2
	MaxNameLength        = 100
	BytesInKB            = 1024
	DefaultMaxFileSizeMB = 10
	EmailPartsCount      = 2
)

// OrganizationSettings представляет настройки организации
type OrganizationSettings struct {
	AllowPublicTickets    bool   `json:"allow_public_tickets"`
	DefaultTicketPriority string `json:"default_ticket_priority"`
	EmailNotifications    bool   `json:"email_notifications"`
	MaxFileSize           int64  `json:"max_file_size"` // в байтах
}

// DefaultSettings возвращает настройки по умолчанию
func DefaultSettings() OrganizationSettings {
	return OrganizationSettings{
		AllowPublicTickets:    false,
		DefaultTicketPriority: "normal",
		EmailNotifications:    true,
		MaxFileSize:           DefaultMaxFileSizeMB * BytesInKB * BytesInKB, // 10MB
	}
}

type Organization struct {
	id        uuid.UUID
	name      string
	domain    string // домен организации для автоматического определения пользователей
	isActive  bool
	settings  OrganizationSettings
	createdAt time.Time
	updatedAt time.Time
}

// NewOrganization создает новую организацию с указанным ID
func NewOrganization(id uuid.UUID, name, domain string) (*Organization, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateDomain(domain); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Organization{
		id:        id,
		name:      name,
		domain:    domain,
		isActive:  true,
		settings:  DefaultSettings(),
		createdAt: now,
		updatedAt: now,
	}, nil
}

// CreateOrganization создает новую организацию с автоматически сгенерированным ID
func CreateOrganization(name, domain string) (*Organization, error) {
	return NewOrganization(uuid.New(), name, domain)
}

// ID возвращает идентификатор организации
func (o *Organization) ID() uuid.UUID {
	return o.id
}

// Name возвращает название организации
func (o *Organization) Name() string {
	return o.name
}

// Domain возвращает домен организации
func (o *Organization) Domain() string {
	return o.domain
}

// IsActive возвращает статус активности организации
func (o *Organization) IsActive() bool {
	return o.isActive
}

// Settings возвращает настройки организации
func (o *Organization) Settings() OrganizationSettings {
	return o.settings
}

// CreatedAt возвращает время создания организации
func (o *Organization) CreatedAt() time.Time {
	return o.createdAt
}

// UpdatedAt возвращает время последнего обновления организации
func (o *Organization) UpdatedAt() time.Time {
	return o.updatedAt
}

// ChangeName изменяет название организации
func (o *Organization) ChangeName(newName string) error {
	if err := validateName(newName); err != nil {
		return err
	}
	o.name = newName
	o.updatedAt = time.Now()
	return nil
}

// ChangeDomain изменяет домен организации
func (o *Organization) ChangeDomain(newDomain string) error {
	if err := validateDomain(newDomain); err != nil {
		return err
	}
	o.domain = newDomain
	o.updatedAt = time.Now()
	return nil
}

// UpdateSettings обновляет настройки организации
func (o *Organization) UpdateSettings(settings OrganizationSettings) {
	o.settings = settings
	o.updatedAt = time.Now()
}

// Activate активирует организацию
func (o *Organization) Activate() {
	o.isActive = true
	o.updatedAt = time.Now()
}

// Deactivate деактивирует организацию
func (o *Organization) Deactivate() {
	o.isActive = false
	o.updatedAt = time.Now()
}

// CanUserJoinByEmail проверяет, может ли пользователь с указанным email автоматически присоединиться к организации
func (o *Organization) CanUserJoinByEmail(email string) bool {
	if o.domain == "" {
		return false
	}

	emailParts := strings.Split(email, "@")
	if len(emailParts) != EmailPartsCount {
		return false
	}

	return strings.EqualFold(emailParts[1], o.domain)
}

// validateName проверяет валидность названия организации
func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrOrganizationValidation)
	}
	if len(name) < MinNameLength {
		return fmt.Errorf("%w: name must be at least %d characters long", ErrOrganizationValidation, MinNameLength)
	}
	if len(name) > MaxNameLength {
		return fmt.Errorf("%w: name must be no more than %d characters long", ErrOrganizationValidation, MaxNameLength)
	}
	return nil
}

// validateDomain проверяет валидность домена организации
func validateDomain(domain string) error {
	if domain == "" {
		return nil // домен может быть пустым
	}

	domain = strings.TrimSpace(strings.ToLower(domain))

	// Простая проверка домена
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("%w: domain must contain at least one dot", ErrOrganizationValidation)
	}

	// Проверка на недопустимые символы
	for _, char := range domain {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '.' && char != '-' {
			return fmt.Errorf("%w: domain contains invalid characters", ErrOrganizationValidation)
		}
	}

	return nil
}
