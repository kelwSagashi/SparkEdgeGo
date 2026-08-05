package projects

import (
	"context"
	"errors"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var ErrInvalidProject = errors.New("invalid project")

type ProjectRepository interface {
	Upsert(ctx context.Context, params sqlite.CreateProjectParams) (domain.Project, error)
	FindByID(ctx context.Context, id string) (domain.Project, error)
	FindByOwner(ctx context.Context, ownerID string) ([]domain.Project, error)
	UpdateByID(ctx context.Context, id string, params sqlite.UpdateProjectParams) (domain.Project, error)
	Delete(ctx context.Context, id string) error
}

type ProjectMembersRepository interface {
	FindByProject(ctx context.Context, projectID string) ([]domain.ProjectMember, error)
	Upsert(ctx context.Context, params sqlite.UpsertProjectMemberParams) (domain.ProjectMember, error)
}

type Service struct {
	projects ProjectRepository
	members  ProjectMembersRepository
}

type CreateRequest struct {
	Name        string                   `json:"name"`
	Key         string                   `json:"key"`
	Description string                   `json:"description"`
	Visibility  domain.ProjectVisibility `json:"visibility"`
	OwnerID     string                   `json:"-"`
}

type UpdateRequest struct {
	Name        *string                   `json:"name"`
	Description *string                   `json:"description"`
	Visibility  *domain.ProjectVisibility `json:"visibility"`
}

type AddMemberRequest struct {
	UserID string             `json:"user_id"`
	Role   domain.ProjectRole `json:"role"`
}

func NewService(projects ProjectRepository, members ProjectMembersRepository) *Service {
	return &Service{projects: projects, members: members}
}

func (s *Service) ListByOwner(ctx context.Context, ownerID string) ([]domain.Project, error) {
	if strings.TrimSpace(ownerID) == "" {
		return []domain.Project{}, nil
	}
	return s.projects.FindByOwner(ctx, ownerID)
}

func (s *Service) FindByID(ctx context.Context, id string) (domain.Project, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Project{}, ErrInvalidProject
	}
	return s.projects.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (domain.Project, error) {
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Key) == "" || strings.TrimSpace(req.OwnerID) == "" {
		return domain.Project{}, ErrInvalidProject
	}
	if req.Visibility == "" {
		req.Visibility = domain.ProjectVisibilityPrivate
	}

	project, err := s.projects.Upsert(ctx, sqlite.CreateProjectParams{
		Name:        req.Name,
		Key:         req.Key,
		Description: req.Description,
		Visibility:  req.Visibility,
		OwnerID:     req.OwnerID,
	})
	if err != nil {
		return domain.Project{}, err
	}

	_, err = s.members.Upsert(ctx, sqlite.UpsertProjectMemberParams{
		ProjectID: project.ID,
		UserID:    req.OwnerID,
		Role:      domain.ProjectRoleOwner,
	})
	if err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (domain.Project, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Project{}, ErrInvalidProject
	}
	return s.projects.UpdateByID(ctx, id, sqlite.UpdateProjectParams{
		Name:        req.Name,
		Description: req.Description,
		Visibility:  req.Visibility,
	})
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidProject
	}
	return s.projects.Delete(ctx, id)
}

func (s *Service) ListMembers(ctx context.Context, projectID string) ([]domain.ProjectMember, error) {
	if strings.TrimSpace(projectID) == "" {
		return []domain.ProjectMember{}, ErrInvalidProject
	}
	return s.members.FindByProject(ctx, projectID)
}

func (s *Service) AddMember(ctx context.Context, projectID string, req AddMemberRequest) (domain.ProjectMember, error) {
	if strings.TrimSpace(projectID) == "" || strings.TrimSpace(req.UserID) == "" {
		return domain.ProjectMember{}, ErrInvalidProject
	}
	if req.Role == "" {
		req.Role = domain.ProjectRoleViewer
	}
	return s.members.Upsert(ctx, sqlite.UpsertProjectMemberParams{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
	})
}
