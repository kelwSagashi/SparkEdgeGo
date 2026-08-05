package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredential = errors.New("invalid credentials")
)

type UserRepository interface {
	Create(ctx context.Context, params sqlite.CreateUserParams) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByAPIKey(ctx context.Context, apiKey string) (domain.User, error)
	CreateAPIKey(ctx context.Context, userID string) (string, error)
}

type ProjectRepository interface {
	EnsurePersonalProject(ctx context.Context, ownerID string) (domain.Project, error)
}

type Service struct {
	users    UserRepository
	projects ProjectRepository
	tokens   *TokenManager
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResult struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func NewService(users UserRepository, projects ProjectRepository, secret string) *Service {
	return &Service{
		users:    users,
		projects: projects,
		tokens:   NewTokenManager(secret, 7*24*time.Hour),
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (domain.User, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return domain.User{}, ErrInvalidCredential
	}

	if existing, err := s.users.FindByEmail(ctx, email); err == nil && existing.ID != "" {
		return domain.User{}, ErrUserAlreadyExists
	} else if err != nil && !errors.Is(err, sqlite.ErrNotFound) {
		return domain.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return domain.User{}, err
	}

	user, err := s.users.Create(ctx, sqlite.CreateUserParams{
		Email:        email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: string(hash),
		Role:         domain.RoleAdmin,
		IsActive:     true,
	})
	if err != nil {
		return domain.User{}, err
	}

	if err := s.ensurePersonalProject(ctx, user.ID); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResult, error) {
	user, err := s.users.FindByEmail(ctx, strings.TrimSpace(strings.ToLower(req.Email)))
	if err != nil {
		if errors.Is(err, sqlite.ErrNotFound) {
			return LoginResult{}, ErrInvalidCredential
		}
		return LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return LoginResult{}, ErrInvalidCredential
	}

	if err := s.ensurePersonalProject(ctx, user.ID); err != nil {
		return LoginResult{}, err
	}

	token, err := s.tokens.Sign(user)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: token, User: user}, nil
}

func (s *Service) VerifyToken(ctx context.Context, token string) (domain.User, error) {
	claims, err := s.tokens.Verify(token)
	if err != nil {
		return domain.User{}, err
	}
	return s.users.FindByID(ctx, claims.UserID)
}

func (s *Service) VerifyAPIKey(ctx context.Context, apiKey string) (domain.User, error) {
	return s.users.FindByAPIKey(ctx, apiKey)
}

func (s *Service) GenerateAPIKey(ctx context.Context, userID string) (string, error) {
	return s.users.CreateAPIKey(ctx, userID)
}

func (s *Service) ensurePersonalProject(ctx context.Context, userID string) error {
	if s.projects == nil {
		return nil
	}
	_, err := s.projects.EnsurePersonalProject(ctx, userID)
	return err
}
