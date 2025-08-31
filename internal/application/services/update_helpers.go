package services

import "github.com/google/uuid"

// Entity represents an entity that can be activated/deactivated
type Entity interface {
	Activate()
	Deactivate()
}

// NamedEntity represents an entity that has a name that can be changed
type NamedEntity interface {
	Entity
	ChangeName(name string) error
}

// ParentEntity represents an entity that can have a parent
type ParentEntity interface {
	ChangeParent(parentID *uuid.UUID) error
}

// updateEntityFields applies common entity field updates
func updateEntityFields(
	entity Entity,
	name *string,
	parentID *uuid.UUID,
	isActive *bool,
	nameUpdater func(string) error,
	parentUpdater func(*uuid.UUID) error,
) (bool, error) {
	updated := false

	// Update name if provided
	if name != nil {
		if err := nameUpdater(*name); err != nil {
			return false, err
		}
		updated = true
	}

	// Update parent if provided
	if parentID != nil {
		if err := parentUpdater(parentID); err != nil {
			return false, err
		}
		updated = true
	}

	// Update active status if provided
	if isActive != nil {
		if *isActive {
			entity.Activate()
		} else {
			entity.Deactivate()
		}
		updated = true
	}

	return updated, nil
}

// updateStringField updates a string field if the new value is provided
func updateStringField(newValue *string, updateFn func(string) error) (bool, error) {
	if newValue == nil {
		return false, nil
	}
	if err := updateFn(*newValue); err != nil {
		return false, err
	}
	return true, nil
}
