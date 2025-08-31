package services

import (
	"context"
	"errors"

	"simpleservicedesk/internal/domain/categories"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/interfaces"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
)

// categoryService implements CategoryService interface
type categoryService struct {
	categoryRepo interfaces.CategoryRepository
	ticketRepo   interfaces.TicketRepository
}

// NewCategoryService creates a new CategoryService implementation
func NewCategoryService(
	categoryRepo interfaces.CategoryRepository,
	ticketRepo interfaces.TicketRepository,
) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		ticketRepo:   ticketRepo,
	}
}

// CreateCategory creates a new category with business logic validation
func (s *categoryService) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*categories.Category, error) {
	// Validate required fields
	if req.Name == "" {
		return nil, errors.New("category name is required")
	}

	// Extract description value
	var description string
	if req.Description != nil {
		description = *req.Description
	}

	// Create category through repository
	category, err := s.categoryRepo.CreateCategory(ctx, func() (*categories.Category, error) {
		return categories.CreateCategory(
			req.Name,
			description,
			req.OrganizationID,
			req.ParentID,
		)
	})
	if err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (s *categoryService) GetCategory(ctx context.Context, id uuid.UUID) (*categories.Category, error) {
	return s.categoryRepo.GetCategory(ctx, id)
}

// UpdateCategory updates category information with business logic
func (s *categoryService) UpdateCategory(
	ctx context.Context,
	id uuid.UUID,
	req UpdateCategoryRequest,
) (*categories.Category, error) {
	category, err := s.categoryRepo.UpdateCategory(ctx, id, func(category *categories.Category) (bool, error) {
		updated := false

		// Update common entity fields (name, parent, active status)
		if commonUpdated, err := updateEntityFields(
			category,
			req.Name,
			req.ParentID,
			req.IsActive,
			category.ChangeName,
			category.ChangeParent,
		); err != nil {
			return false, err
		} else if commonUpdated {
			updated = true
		}

		// Update description if provided (specific to category)
		if descUpdated, err := updateStringField(req.Description, category.ChangeDescription); err != nil {
			return false, err
		} else if descUpdated {
			updated = true
		}

		return updated, nil
	})

	if err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory deletes a category by ID
func (s *categoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.categoryRepo.DeleteCategory(ctx, id)
}

// ListCategories returns a paginated list of categories
func (s *categoryService) ListCategories(
	ctx context.Context,
	filter queries.CategoryFilter,
) ([]*categories.Category, int64, error) {
	categoryList, err := s.categoryRepo.ListCategories(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// For count, we return the length of the list
	// In a real implementation, we would have a separate count method
	count := int64(len(categoryList))

	return categoryList, count, nil
}

// GetCategoryTickets returns tickets for a specific category
func (s *categoryService) GetCategoryTickets(
	ctx context.Context,
	categoryID uuid.UUID,
	filter queries.TicketFilter,
) ([]*tickets.Ticket, int64, error) {
	// First, check if the category exists
	_, err := s.categoryRepo.GetCategory(ctx, categoryID)
	if err != nil {
		return nil, 0, err
	}

	// Set the category filter
	filter.CategoryID = &categoryID

	ticketList, err := s.ticketRepo.ListTickets(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// For count, we return the length of the list
	// In a real implementation, we would have a separate count method
	count := int64(len(ticketList))

	return ticketList, count, nil
}
