package sqlite

import (
	"context"
	"errors"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type TagsRepository struct {
	db *gorm.DB
}

type UpsertTagParams struct {
	ID        string
	Name      string
	Color     string
	ProjectID string
}

func NewTagsRepository(db *gorm.DB) *TagsRepository {
	return &TagsRepository{db: db}
}

func (r *TagsRepository) FindAll(ctx context.Context, projectID string) ([]domain.Tag, error) {
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	var models []tagModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	return tagsFromModels(models), nil
}

func (r *TagsRepository) FindByID(ctx context.Context, id string) (domain.Tag, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *TagsRepository) FindByNameAndProject(ctx context.Context, name string, projectID string) (domain.Tag, error) {
	if projectID == "" {
		return r.findOne(ctx, "name = ? AND project_id = ?", name, "")
	}
	return r.findOne(ctx, "name = ? AND project_id = ?", name, projectID)
}

func (r *TagsRepository) Search(ctx context.Context, query string, projectID string) ([]domain.Tag, error) {
	search := "%" + query + "%"
	db := r.db.WithContext(ctx).Where("name LIKE ?", search).Order("name ASC")
	if projectID != "" {
		db = db.Where("project_id = ?", projectID)
	} else {
		db = db.Where("project_id = ?", "")
	}

	var models []tagModel
	if err := db.Find(&models).Error; err != nil {
		return nil, err
	}
	return tagsFromModels(models), nil
}

func (r *TagsRepository) Upsert(ctx context.Context, params UpsertTagParams) (domain.Tag, error) {
	existing, err := r.FindByNameAndProject(ctx, params.Name, params.ProjectID)
	if err == nil && existing.ID != "" {
		model := tagModel{
			ID:        existing.ID,
			Name:      existing.Name,
			Color:     params.Color,
			ProjectID: existing.ProjectID,
			CreatedAt: existing.CreatedAt,
		}
		if model.Color == "" {
			model.Color = "#6b7280"
		}
		if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
			return domain.Tag{}, err
		}
		return tagFromModel(model), nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return domain.Tag{}, err
	}

	model := tagModel{
		ID:        params.ID,
		Name:      strings.TrimSpace(params.Name),
		Color:     params.Color,
		ProjectID: params.ProjectID,
	}
	if model.ID == "" {
		model.ID = newID()
	}
	if model.Color == "" {
		model.Color = "#6b7280"
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Tag{}, err
	}
	return tagFromModel(model), nil
}

func (r *TagsRepository) Delete(ctx context.Context, id string) (domain.Tag, error) {
	tag, err := r.FindByID(ctx, id)
	if err != nil {
		return domain.Tag{}, err
	}
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&tagModel{}).Error; err != nil {
		return domain.Tag{}, err
	}
	return tag, nil
}

func (r *TagsRepository) findOne(ctx context.Context, query string, args ...any) (domain.Tag, error) {
	var model tagModel
	if err := r.db.WithContext(ctx).Where(query, args...).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Tag{}, ErrNotFound
		}
		return domain.Tag{}, err
	}
	return tagFromModel(model), nil
}

func tagsFromModels(models []tagModel) []domain.Tag {
	tags := make([]domain.Tag, 0, len(models))
	for _, model := range models {
		tags = append(tags, tagFromModel(model))
	}
	return tags
}

func tagFromModel(model tagModel) domain.Tag {
	return domain.Tag{
		ID:        model.ID,
		Name:      model.Name,
		Color:     model.Color,
		ProjectID: model.ProjectID,
		CreatedAt: model.CreatedAt,
	}
}
