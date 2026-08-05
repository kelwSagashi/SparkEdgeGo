package tags

import (
	"context"
	"errors"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var ErrInvalidTag = errors.New("invalid tag")

type TagRepository interface {
	FindAll(ctx context.Context, projectID string) ([]domain.Tag, error)
	FindByID(ctx context.Context, id string) (domain.Tag, error)
	FindByNameAndProject(ctx context.Context, name string, projectID string) (domain.Tag, error)
	Search(ctx context.Context, query string, projectID string) ([]domain.Tag, error)
	Upsert(ctx context.Context, params sqlite.UpsertTagParams) (domain.Tag, error)
	Delete(ctx context.Context, id string) (domain.Tag, error)
}

type InstanceTagRepository interface {
	FindByInstance(ctx context.Context, instanceID string) ([]domain.Tag, error)
	Link(ctx context.Context, instanceID string, tagID string) (domain.InstanceTag, error)
	Unlink(ctx context.Context, instanceID string, tagID string) error
	SyncTags(ctx context.Context, instanceID string, tagIDs []string) ([]domain.Tag, error)
}

type Service struct {
	tags         TagRepository
	instanceTags InstanceTagRepository
}

type CreateRequest struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	ProjectID string `json:"project_id"`
}

func NewService(tags TagRepository, instanceTags InstanceTagRepository) *Service {
	return &Service{tags: tags, instanceTags: instanceTags}
}

func (s *Service) ListAll(ctx context.Context, projectID string) ([]domain.Tag, error) {
	return s.tags.FindAll(ctx, projectID)
}

func (s *Service) FindByID(ctx context.Context, id string) (domain.Tag, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Tag{}, ErrInvalidTag
	}
	return s.tags.FindByID(ctx, id)
}

func (s *Service) Search(ctx context.Context, query string, projectID string) ([]domain.Tag, error) {
	return s.tags.Search(ctx, strings.TrimSpace(query), projectID)
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (domain.Tag, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.Tag{}, ErrInvalidTag
	}
	return s.tags.Upsert(ctx, sqlite.UpsertTagParams{
		Name:      name,
		Color:     req.Color,
		ProjectID: req.ProjectID,
	})
}

func (s *Service) Delete(ctx context.Context, id string) (domain.Tag, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Tag{}, ErrInvalidTag
	}
	return s.tags.Delete(ctx, id)
}

func (s *Service) FindByInstance(ctx context.Context, instanceID string) ([]domain.Tag, error) {
	if strings.TrimSpace(instanceID) == "" {
		return []domain.Tag{}, ErrInvalidTag
	}
	return s.instanceTags.FindByInstance(ctx, instanceID)
}

func (s *Service) LinkTag(ctx context.Context, instanceID string, tagID string) (domain.InstanceTag, error) {
	if strings.TrimSpace(instanceID) == "" || strings.TrimSpace(tagID) == "" {
		return domain.InstanceTag{}, ErrInvalidTag
	}
	return s.instanceTags.Link(ctx, instanceID, tagID)
}

func (s *Service) UnlinkTag(ctx context.Context, instanceID string, tagID string) error {
	if strings.TrimSpace(instanceID) == "" || strings.TrimSpace(tagID) == "" {
		return ErrInvalidTag
	}
	return s.instanceTags.Unlink(ctx, instanceID, tagID)
}

func (s *Service) SyncTags(ctx context.Context, instanceID string, tagIDs []string) ([]domain.Tag, error) {
	if strings.TrimSpace(instanceID) == "" {
		return []domain.Tag{}, ErrInvalidTag
	}
	return s.instanceTags.SyncTags(ctx, instanceID, tagIDs)
}

func (s *Service) FindOrCreateByNames(ctx context.Context, names []string, projectID string) ([]string, error) {
	tagIDs := make([]string, 0, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}

		existing, err := s.tags.FindByNameAndProject(ctx, trimmed, projectID)
		if err == nil && existing.ID != "" {
			tagIDs = append(tagIDs, existing.ID)
			continue
		}
		if err != nil && !errors.Is(err, sqlite.ErrNotFound) {
			return nil, err
		}

		created, err := s.Create(ctx, CreateRequest{Name: trimmed, ProjectID: projectID})
		if err != nil {
			return nil, err
		}
		tagIDs = append(tagIDs, created.ID)
	}
	return tagIDs, nil
}
