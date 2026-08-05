package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type ProjectMembersRepository struct {
	db *gorm.DB
}

type UpsertProjectMemberParams struct {
	ProjectID string
	UserID    string
	Role      domain.ProjectRole
}

func NewProjectMembersRepository(db *gorm.DB) *ProjectMembersRepository {
	return &ProjectMembersRepository{db: db}
}

func (r *ProjectMembersRepository) Upsert(ctx context.Context, params UpsertProjectMemberParams) (domain.ProjectMember, error) {
	if params.Role == "" {
		params.Role = domain.ProjectRoleViewer
	}

	if existing, err := r.findByProjectAndUser(ctx, params.ProjectID, params.UserID); err == nil && existing.ID != "" {
		existing.Role = params.Role
		model := projectMemberModel{
			ID:        existing.ID,
			ProjectID: existing.ProjectID,
			UserID:    existing.UserID,
			Role:      string(existing.Role),
			CreatedAt: existing.CreatedAt,
		}
		if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
			return domain.ProjectMember{}, err
		}
		return projectMemberFromModel(model), nil
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return domain.ProjectMember{}, err
	}

	model := projectMemberModel{
		ID:        newID(),
		ProjectID: params.ProjectID,
		UserID:    params.UserID,
		Role:      string(params.Role),
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.ProjectMember{}, err
	}
	return projectMemberFromModel(model), nil
}

func (r *ProjectMembersRepository) FindByProject(ctx context.Context, projectID string) ([]domain.ProjectMember, error) {
	var models []projectMemberModel
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	members := make([]domain.ProjectMember, 0, len(models))
	for _, model := range models {
		members = append(members, projectMemberFromModel(model))
	}
	return members, nil
}

func (r *ProjectMembersRepository) findByProjectAndUser(ctx context.Context, projectID string, userID string) (domain.ProjectMember, error) {
	var model projectMemberModel
	if err := r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ProjectMember{}, ErrNotFound
		}
		return domain.ProjectMember{}, err
	}
	return projectMemberFromModel(model), nil
}

func projectMemberFromModel(model projectMemberModel) domain.ProjectMember {
	return domain.ProjectMember{
		ID:        model.ID,
		ProjectID: model.ProjectID,
		UserID:    model.UserID,
		Role:      domain.ProjectRole(model.Role),
		CreatedAt: model.CreatedAt,
	}
}
