package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	authdomain "simpleservicedesk/internal/domain/auth"
	"simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

const (
	emailLookupLimit = 1
	//nolint:gosec // Static bcrypt hash is intentionally used only to equalize credential-check timing on auth failures.
	dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
)

type UserRepository interface {
	ListUsers(ctx context.Context, filter queries.UserFilter) ([]*users.User, error)
}

type Service struct {
	userRepo        UserRepository
	signingKey      []byte
	tokenExpiration time.Duration
	currentTime     func() time.Time
}

func NewService(userRepo UserRepository, signingKey string, tokenExpiration time.Duration) (*Service, error) {
	if userRepo == nil {
		return nil, errors.New("user repository is required")
	}
	if strings.TrimSpace(signingKey) == "" {
		return nil, errors.New("jwt signing key is required")
	}
	if tokenExpiration <= 0 {
		return nil, errors.New("jwt expiration must be greater than zero")
	}

	return &Service{
		userRepo:        userRepo,
		signingKey:      []byte(signingKey),
		tokenExpiration: tokenExpiration,
		currentTime:     time.Now,
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	normalizedEmail := strings.TrimSpace(email)
	if normalizedEmail == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	emailPattern := "^" + regexp.QuoteMeta(normalizedEmail) + "$"
	usersByEmail, err := s.userRepo.ListUsers(ctx, queries.UserFilter{
		BaseFilter: queries.BaseFilter{Limit: emailLookupLimit},
		Email:      &emailPattern,
	})
	if err != nil {
		return "", fmt.Errorf("failed to find user by email: %w", err)
	}

	user, err := findExactEmailUser(usersByEmail, normalizedEmail)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			consumePasswordTiming(password)
		}
		return "", err
	}

	passwordMatches := user.CheckPassword(password)
	if !passwordMatches || !user.IsActive() {
		return "", ErrInvalidCredentials
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate auth token: %w", err)
	}

	return token, nil
}

func (s *Service) GenerateToken(user *users.User) (string, error) {
	if user == nil {
		return "", errors.New("user is required")
	}

	issuedAt := s.currentTime().UTC()
	claims := authdomain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID().String(),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(s.tokenExpiration)),
		},
		UserID: user.ID().String(),
		Role:   user.Role(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *Service) ValidateToken(tokenString string) (*authdomain.Claims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, ErrInvalidToken
	}

	claims := &authdomain.Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		s.validateSigningMethod,
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, errors.Join(ErrInvalidToken, err)
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if _, parseErr := uuid.Parse(claims.UserID); parseErr != nil {
		return nil, fmt.Errorf("%w: invalid user id", ErrInvalidToken)
	}
	if !claims.Role.IsValid() {
		return nil, fmt.Errorf("%w: invalid role", ErrInvalidToken)
	}

	return claims, nil
}

func (s *Service) validateSigningMethod(token *jwt.Token) (any, error) {
	if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		return nil, fmt.Errorf("%w: unexpected signing method: %s", ErrInvalidToken, token.Method.Alg())
	}

	return s.signingKey, nil
}

func findExactEmailUser(usersByEmail []*users.User, email string) (*users.User, error) {
	for _, user := range usersByEmail {
		if strings.EqualFold(user.Email(), email) {
			return user, nil
		}
	}

	return nil, ErrInvalidCredentials
}

func consumePasswordTiming(password string) {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(password))
}
