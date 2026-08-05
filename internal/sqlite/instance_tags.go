package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type InstanceTagsRepository struct {
	db *gorm.DB
}

func NewInstanceTagsRepository(db *gorm.DB) *InstanceTagsRepository {
	return &InstanceTagsRepository{db: db}
}

func (r *InstanceTagsRepository) FindByInstance(ctx context.Context, instanceID string) ([]domain.Tag, error) {
	var links []instanceTagModel
	if err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Find(&links).Error; err != nil {
		return nil, err
	}

	tagIDs := make([]string, 0, len(links))
	for _, link := range links {
		tagIDs = append(tagIDs, link.TagID)
	}
	if len(tagIDs) == 0 {
		return []domain.Tag{}, nil
	}

	var tags []tagModel
	if err := r.db.WithContext(ctx).Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tagsFromModels(tags), nil
}

func (r *InstanceTagsRepository) Link(ctx context.Context, instanceID string, tagID string) (domain.InstanceTag, error) {
	existing, err := r.find(ctx, instanceID, tagID)
	if err == nil && existing.ID != "" {
		return existing, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return domain.InstanceTag{}, err
	}

	model := instanceTagModel{
		ID:         newID(),
		InstanceID: instanceID,
		TagID:      tagID,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.InstanceTag{}, err
	}
	return instanceTagFromModel(model), nil
}

func (r *InstanceTagsRepository) Unlink(ctx context.Context, instanceID string, tagID string) error {
	return r.db.WithContext(ctx).Where("instance_id = ? AND tag_id = ?", instanceID, tagID).Delete(&instanceTagModel{}).Error
}

func (r *InstanceTagsRepository) SyncTags(ctx context.Context, instanceID string, tagIDs []string) ([]domain.Tag, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("instance_id = ?", instanceID).Delete(&instanceTagModel{}).Error; err != nil {
			return err
		}
		for _, tagID := range tagIDs {
			if err := tx.Create(&instanceTagModel{ID: newID(), InstanceID: instanceID, TagID: tagID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.FindByInstance(ctx, instanceID)
}

func (r *InstanceTagsRepository) find(ctx context.Context, instanceID string, tagID string) (domain.InstanceTag, error) {
	var model instanceTagModel
	if err := r.db.WithContext(ctx).Where("instance_id = ? AND tag_id = ?", instanceID, tagID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.InstanceTag{}, ErrNotFound
		}
		return domain.InstanceTag{}, err
	}
	return instanceTagFromModel(model), nil
}

func instanceTagFromModel(model instanceTagModel) domain.InstanceTag {
	return domain.InstanceTag{
		ID:         model.ID,
		InstanceID: model.InstanceID,
		TagID:      model.TagID,
		CreatedAt:  model.CreatedAt,
	}
}
