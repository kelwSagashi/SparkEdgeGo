package users

import (
	"context"
	"errors"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var ErrInvalidUser = errors.New("invalid user")

type Repository interface {
	Create(ctx context.Context, params sqlite.CreateUserParams) (domain.User, error)
	Upsert(ctx context.Context, params sqlite.UpdateUserParams) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
	ListAll(ctx context.Context) ([]domain.User, error)
	Delete(ctx context.Context, id string) error
	CreateAPIKey(ctx context.Context, userID string) (string, error)
	FindProjectUserByName(ctx context.Context, id string, projectName string) (domain.ProjectUser, error)
}

type Service struct {
	users Repository
}

type CreateRequest struct {
	Email        string          `json:"email"`
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	PasswordHash string          `json:"password_hash"`
	Role         domain.UserRole `json:"role"`
	IsActive     *bool           `json:"is_active"`
}

type UpdateRequest struct {
	Email        *string          `json:"email"`
	FirstName    *string          `json:"first_name"`
	LastName     *string          `json:"last_name"`
	PasswordHash *string          `json:"password_hash"`
	Role         *domain.UserRole `json:"role"`
	IsActive     *bool            `json:"is_active"`
}

func NewService(users Repository) *Service {
	return &Service{users: users}
}

func (s *Service) ListAll(ctx context.Context) ([]domain.User, error) {
	return s.users.ListAll(ctx)
}

func (s *Service) FindByID(ctx context.Context, id string) (domain.User, error) {
	if strings.TrimSpace(id) == "" {
		return domain.User{}, ErrInvalidUser
	}
	return s.users.FindByID(ctx, id)
}

func (s *Service) FindProjectUserByName(ctx context.Context, id string, projectName string) (domain.ProjectUser, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(projectName) == "" {
		return domain.ProjectUser{}, ErrInvalidUser
	}
	return s.users.FindProjectUserByName(ctx, id, projectName)
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (domain.User, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return domain.User{}, ErrInvalidUser
	}

	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	if req.Role == "" {
		req.Role = domain.RoleViewer
	}

	return s.users.Create(ctx, sqlite.CreateUserParams{
		Email:        email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: req.PasswordHash,
		Role:         req.Role,
		IsActive:     active,
	})
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (domain.User, error) {
	if strings.TrimSpace(id) == "" {
		return domain.User{}, ErrInvalidUser
	}
	if req.Email != nil {
		normalized := strings.TrimSpace(strings.ToLower(*req.Email))
		req.Email = &normalized
	}
	return s.users.Upsert(ctx, sqlite.UpdateUserParams{
		ID:           id,
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: req.PasswordHash,
		Role:         req.Role,
		IsActive:     req.IsActive,
	})
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidUser
	}
	return s.users.Delete(ctx, id)
}

func (s *Service) CreateAPIKey(ctx context.Context, id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", ErrInvalidUser
	}
	return s.users.CreateAPIKey(ctx, id)
}
