package categories

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCategoryNotFound     = errors.New("category not found")
	ErrInvalidCategory      = errors.New("invalid category")
	ErrCategoryValidation   = errors.New("category validation error")
	ErrCategoryAlreadyExist = errors.New("category already exist")
	ErrCircularReference    = errors.New("circular reference detected")
)

const (
	MinNameLength        = 2
	MaxNameLength        = 100
	MaxDescriptionLength = 500
)

type Category struct {
	id             uuid.UUID
	name           string
	description    string
	organizationID uuid.UUID
	parentID       *uuid.UUID // Указатель, так как может быть nil для корневых категорий
	isActive       bool
	createdAt      time.Time
	updatedAt      time.Time
}

// NewCategory создает новую категорию с указанным ID
func NewCategory(id uuid.UUID,
	name, description string,
	organizationID uuid.UUID,
	parentID *uuid.UUID) (*Category, error) {
	name, err := validateName(name)
	if err != nil {
		return nil, err
	}
	description, err = validateDescription(description)
	if err != nil {
		return nil, err
	}
	if err = validateOrganizationID(organizationID); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Category{
		id:             id,
		name:           name,
		description:    description,
		organizationID: organizationID,
		parentID:       parentID,
		isActive:       true,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

// CreateCategory создает новую категорию с автоматически сгенерированным ID
func CreateCategory(name, description string, organizationID uuid.UUID, parentID *uuid.UUID) (*Category, error) {
	return NewCategory(uuid.New(), name, description, organizationID, parentID)
}

// CreateRootCategory создает корневую категорию (без родителя)
func CreateRootCategory(name, description string, organizationID uuid.UUID) (*Category, error) {
	return CreateCategory(name, description, organizationID, nil)
}

// CreateSubCategory создает дочернюю категорию
func CreateSubCategory(name, description string, organizationID, parentID uuid.UUID) (*Category, error) {
	return CreateCategory(name, description, organizationID, &parentID)
}

// ID возвращает идентификатор категории
func (c *Category) ID() uuid.UUID {
	return c.id
}

// Name возвращает название категории
func (c *Category) Name() string {
	return c.name
}

// Description возвращает описание категории
func (c *Category) Description() string {
	return c.description
}

// OrganizationID возвращает ID организации
func (c *Category) OrganizationID() uuid.UUID {
	return c.organizationID
}

// ParentID возвращает ID родительской категории (может быть nil)
func (c *Category) ParentID() *uuid.UUID {
	return c.parentID
}

// IsActive возвращает статус активности категории
func (c *Category) IsActive() bool {
	return c.isActive
}

// CreatedAt возвращает время создания категории
func (c *Category) CreatedAt() time.Time {
	return c.createdAt
}

// UpdatedAt возвращает время последнего обновления категории
func (c *Category) UpdatedAt() time.Time {
	return c.updatedAt
}

// IsRootCategory проверяет, является ли категория корневой
func (c *Category) IsRootCategory() bool {
	return c.parentID == nil
}

// HasParent проверяет, есть ли у категории родитель
func (c *Category) HasParent() bool {
	return c.parentID != nil
}

// ChangeName изменяет название категории
func (c *Category) ChangeName(newName string) error {
	newName, err := validateName(newName)
	if err != nil {
		return err
	}
	c.name = newName
	c.updatedAt = time.Now()
	return nil
}

// ChangeDescription изменяет описание категории
func (c *Category) ChangeDescription(newDescription string) error {
	newDescription, err := validateDescription(newDescription)
	if err != nil {
		return err
	}
	c.description = newDescription
	c.updatedAt = time.Now()
	return nil
}

// ChangeParent изменяет родительскую категорию
// Валидация на циклические ссылки должна происходить на уровне сервиса
func (c *Category) ChangeParent(newParentID *uuid.UUID) error {
	// Нельзя сделать себя своим родителем
	if newParentID != nil && *newParentID == c.id {
		return fmt.Errorf("%w: category cannot be parent of itself", ErrCircularReference)
	}

	c.parentID = newParentID
	c.updatedAt = time.Now()
	return nil
}

// MoveToRoot делает категорию корневой (убирает родителя)
func (c *Category) MoveToRoot() {
	c.parentID = nil
	c.updatedAt = time.Now()
}

// Activate активирует категорию
func (c *Category) Activate() {
	c.isActive = true
	c.updatedAt = time.Now()
}

// Deactivate деактивирует категорию
func (c *Category) Deactivate() {
	c.isActive = false
	c.updatedAt = time.Now()
}

// BelongsToOrganization проверяет, принадлежит ли категория указанной организации
func (c *Category) BelongsToOrganization(organizationID uuid.UUID) bool {
	return c.organizationID == organizationID
}

// FullPath возвращает полный путь категории (для отображения в UI)
// Например: "IT / Hardware / Laptops"
func (c *Category) FullPath(getParentName func(uuid.UUID) (string, error)) (string, error) {
	if c.IsRootCategory() {
		return c.name, nil
	}

	parentName, err := getParentName(*c.parentID)
	if err != nil {
		return "", fmt.Errorf("failed to get parent name: %w", err)
	}

	return parentName + " / " + c.name, nil
}

// validateName проверяет валидность названия категории и возвращает очищенное имя
func validateName(name string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", fmt.Errorf("%w: name is required", ErrCategoryValidation)
	}
	if len(trimmedName) < MinNameLength {
		return "", fmt.Errorf("%w: name must be at least %d characters long", ErrCategoryValidation, MinNameLength)
	}
	if len(trimmedName) > MaxNameLength {
		return "", fmt.Errorf("%w: name must be no more than %d characters long", ErrCategoryValidation, MaxNameLength)
	}
	return trimmedName, nil
}

// validateDescription проверяет валидность описания категории
func validateDescription(description string) (string, error) {
	trimmedDescription := strings.TrimSpace(description)
	if len(trimmedDescription) > MaxDescriptionLength {
		return trimmedDescription, fmt.Errorf(
			"%w: description must be no more than %d characters long",
			ErrCategoryValidation,
			MaxDescriptionLength,
		)
	}
	return trimmedDescription, nil
}

// validateOrganizationID проверяет валидность ID организации
func validateOrganizationID(organizationID uuid.UUID) error {
	if organizationID == uuid.Nil {
		return fmt.Errorf("%w: organization ID is required", ErrCategoryValidation)
	}
	return nil
}
