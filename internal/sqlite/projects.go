package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type ProjectsRepository struct {
	db *gorm.DB
}

type CreateProjectParams struct {
	Name        string
	Key         string
	Description string
	Visibility  domain.ProjectVisibility
	OwnerID     string
}

type UpdateProjectParams struct {
	Name        *string
	Description *string
	Visibility  *domain.ProjectVisibility
}

func NewProjectsRepository(db *gorm.DB) *ProjectsRepository {
	return &ProjectsRepository{db: db}
}

func (r *ProjectsRepository) Upsert(ctx context.Context, params CreateProjectParams) (domain.Project, error) {
	if params.Visibility == "" {
		params.Visibility = domain.ProjectVisibilityPrivate
	}

	if existing, err := r.FindByKey(ctx, params.Key); err == nil && existing.ID != "" {
		return r.UpdateByID(ctx, existing.ID, UpdateProjectParams{
			Name:        &params.Name,
			Description: &params.Description,
			Visibility:  &params.Visibility,
		})
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return domain.Project{}, err
	}

	model := projectModel{
		ID:          newID(),
		Name:        params.Name,
		Key:         params.Key,
		Description: params.Description,
		Visibility:  string(params.Visibility),
		OwnerID:     params.OwnerID,
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Project{}, err
	}
	return projectFromModel(model), nil
}

func (r *ProjectsRepository) EnsurePersonalProject(ctx context.Context, ownerID string) (domain.Project, error) {
	return r.Upsert(ctx, CreateProjectParams{
		Name:       "PERSONAL",
		Key:        "PERSONAL",
		Visibility: domain.ProjectVisibilityPrivate,
		OwnerID:    ownerID,
	})
}

func (r *ProjectsRepository) FindByID(ctx context.Context, id string) (domain.Project, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *ProjectsRepository) FindByKey(ctx context.Context, key string) (domain.Project, error) {
	return r.findOne(ctx, "key = ?", key)
}

func (r *ProjectsRepository) FindByOwner(ctx context.Context, ownerID string) ([]domain.Project, error) {
	var models []projectModel
	if err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	projects := make([]domain.Project, 0, len(models))
	for _, model := range models {
		projects = append(projects, projectFromModel(model))
	}
	return projects, nil
}

func (r *ProjectsRepository) UpdateByID(ctx context.Context, id string, params UpdateProjectParams) (domain.Project, error) {
	var model projectModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Project{}, ErrNotFound
		}
		return domain.Project{}, err
	}

	if params.Name != nil {
		model.Name = *params.Name
	}
	if params.Description != nil {
		model.Description = *params.Description
	}
	if params.Visibility != nil {
		model.Visibility = string(*params.Visibility)
	}

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.Project{}, err
	}
	return projectFromModel(model), nil
}

func (r *ProjectsRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&projectModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProjectsRepository) findOne(ctx context.Context, query string, args ...any) (domain.Project, error) {
	var model projectModel
	if err := r.db.WithContext(ctx).Where(query, args...).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Project{}, ErrNotFound
		}
		return domain.Project{}, err
	}
	return projectFromModel(model), nil
}

func projectFromModel(model projectModel) domain.Project {
	return domain.Project{
		ID:          model.ID,
		Name:        model.Name,
		Key:         model.Key,
		Description: model.Description,
		Visibility:  domain.ProjectVisibility(model.Visibility),
		OwnerID:     model.OwnerID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}
