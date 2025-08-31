package services

import (
	"context"
	"errors"

	"simpleservicedesk/internal/domain/tickets"
	domain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/interfaces"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// userService implements UserService interface
type userService struct {
	userRepo   interfaces.UserRepository
	ticketRepo interfaces.TicketRepository
}

// NewUserService creates a new UserService implementation
func NewUserService(userRepo interfaces.UserRepository, ticketRepo interfaces.TicketRepository) UserService {
	return &userService{
		userRepo:   userRepo,
		ticketRepo: ticketRepo,
	}
}

// CreateUser creates a new user with business logic validation
func (s *userService) CreateUser(ctx context.Context, req CreateUserRequest) (*domain.User, error) {
	// Password validation
	if req.Password == "" {
		return nil, errors.New("password is required")
	}
	if len(req.Password) < domain.MinPasswordLength {
		return nil, errors.New("password must be at least 6 characters long")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	// Create user through repository
	user, err := s.userRepo.CreateUser(ctx, req.Email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(req.Name, req.Email, passwordHash)
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetUser(ctx, id)
}

// UpdateUser updates user information with business logic
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req UpdateUserRequest) (*domain.User, error) {
	user, err := s.userRepo.UpdateUser(ctx, id, func(user *domain.User) (bool, error) {
		updated := false

		// Update name if provided
		if req.Name != nil {
			if err := user.ChangeName(*req.Name); err != nil {
				return false, err
			}
			updated = true
		}

		// Update email if provided
		if req.Email != nil {
			if err := user.ChangeEmail(*req.Email); err != nil {
				return false, err
			}
			updated = true
		}

		// Update password if provided
		if req.Password != nil {
			// Note: Password updates are not supported in the general update method
			// Password changes should be handled by a separate dedicated endpoint
			// This is a security best practice
			return false, errors.New(
				"password updates not supported in general update - use dedicated password change endpoint",
			)
		}

		// Update active status if provided
		if req.IsActive != nil {
			if *req.IsActive {
				user.Activate()
			} else {
				user.Deactivate()
			}
			updated = true
		}

		return updated, nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deactivates a user by ID (soft delete)
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := s.userRepo.UpdateUser(ctx, id, func(user *domain.User) (bool, error) {
		user.Deactivate()
		return true, nil
	})
	return err
}

// ListUsers returns a paginated list of users
func (s *userService) ListUsers(ctx context.Context, filter queries.UserFilter) ([]*domain.User, int64, error) {
	users, err := s.userRepo.ListUsers(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.userRepo.CountUsers(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

// UpdateUserRole updates a user's role
func (s *userService) UpdateUserRole(ctx context.Context, id uuid.UUID, role domain.Role) (*domain.User, error) {
	user, err := s.userRepo.UpdateUser(ctx, id, func(user *domain.User) (bool, error) {
		if err := user.ChangeRole(role); err != nil {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserTickets returns tickets for a specific user
func (s *userService) GetUserTickets(
	ctx context.Context,
	userID uuid.UUID,
	filter queries.TicketFilter,
) ([]*tickets.Ticket, int64, error) {
	// Set the user filter
	filter.AuthorID = &userID

	ticketList, err := s.ticketRepo.ListTickets(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// For count, we need to implement a count method in ticket repository
	// For now, return the length of the list as count
	count := int64(len(ticketList))

	return ticketList, count, nil
}
