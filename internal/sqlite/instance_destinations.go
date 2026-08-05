package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type InstanceDestinationsRepository struct {
	db *gorm.DB
}

type UpsertInstanceDestinationParams struct {
	ID                  string
	InstanceID          string
	ResourceOperationID string
	Enabled             bool
	Priority            int
	RetryPolicy         domain.RetryPolicy
}

func NewInstanceDestinationsRepository(db *gorm.DB) *InstanceDestinationsRepository {
	return &InstanceDestinationsRepository{db: db}
}

func (r *InstanceDestinationsRepository) Create(ctx context.Context, params UpsertInstanceDestinationParams) (domain.InstanceDestination, error) {
	model := destinationModelFromParams(params)
	if model.ID == "" {
		model.ID = newID()
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.InstanceDestination{}, err
	}
	return destinationFromModel(model), nil
}

func (r *InstanceDestinationsRepository) Upsert(ctx context.Context, params UpsertInstanceDestinationParams) (domain.InstanceDestination, error) {
	if params.ID == "" {
		return r.Create(ctx, params)
	}

	var model instanceDestinationModel
	err := r.db.WithContext(ctx).Where("id = ?", params.ID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.Create(ctx, params)
	}
	if err != nil {
		return domain.InstanceDestination{}, err
	}

	applyDestinationParams(&model, params)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.InstanceDestination{}, err
	}
	return destinationFromModel(model), nil
}

func (r *InstanceDestinationsRepository) FindByID(ctx context.Context, id string) (domain.InstanceDestination, error) {
	var model instanceDestinationModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.InstanceDestination{}, ErrNotFound
		}
		return domain.InstanceDestination{}, err
	}
	return destinationFromModel(model), nil
}

func (r *InstanceDestinationsRepository) ListByInstance(ctx context.Context, instanceID string) ([]domain.InstanceDestination, error) {
	var models []instanceDestinationModel
	if err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("priority ASC, created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	destinations := make([]domain.InstanceDestination, 0, len(models))
	for _, model := range models {
		destinations = append(destinations, destinationFromModel(model))
	}
	return destinations, nil
}

func (r *InstanceDestinationsRepository) DeleteByInstance(ctx context.Context, instanceID string) error {
	return r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Delete(&instanceDestinationModel{}).Error
}

func destinationModelFromParams(params UpsertInstanceDestinationParams) instanceDestinationModel {
	model := instanceDestinationModel{ID: params.ID}
	applyDestinationParams(&model, params)
	return model
}

func applyDestinationParams(model *instanceDestinationModel, params UpsertInstanceDestinationParams) {
	if params.RetryPolicy.MaxRetries == 0 {
		params.RetryPolicy.MaxRetries = 3
	}
	if params.RetryPolicy.RetryInterval == 0 {
		params.RetryPolicy.RetryInterval = 60
	}

	model.InstanceID = params.InstanceID
	model.ResourceOperationID = params.ResourceOperationID
	model.Enabled = params.Enabled
	model.Priority = params.Priority
	model.RetryPolicy = mapJSON{
		"max_retries":    params.RetryPolicy.MaxRetries,
		"retry_interval": params.RetryPolicy.RetryInterval,
	}
}

func destinationFromModel(model instanceDestinationModel) domain.InstanceDestination {
	return domain.InstanceDestination{
		ID:                  model.ID,
		InstanceID:          model.InstanceID,
		ResourceOperationID: model.ResourceOperationID,
		Enabled:             model.Enabled,
		Priority:            model.Priority,
		RetryPolicy: domain.RetryPolicy{
			MaxRetries:    intFromJSON(model.RetryPolicy, "max_retries"),
			RetryInterval: intFromJSON(model.RetryPolicy, "retry_interval"),
		},
		CreatedAt: model.CreatedAt,
	}
}

func intFromJSON(value mapJSON, key string) int {
	switch raw := value[key].(type) {
	case float64:
		return int(raw)
	case int:
		return raw
	default:
		return 0
	}
}
