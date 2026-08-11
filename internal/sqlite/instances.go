package sqlite

import (
	"context"
	"errors"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type InstancesRepository struct {
	db *gorm.DB
}

type UpsertInstanceParams struct {
	ID                           string
	Name                         string
	Description                  string
	Tags                         []string
	Status                       domain.InstanceStatus
	Active                       *bool
	ProjectID                    string
	DeviceID                     string
	ScriptID                     string
	IncludeDeviceData            bool
	ScriptParameters             map[string]any
	TriggerType                  domain.TriggerType
	TriggerConfig                map[string]any
	DependsOn                    []string
	ExecutionMode                domain.ExecutionMode
	OrchestrationConfig          map[string]any
	FallbackEnabled              bool
	FallbackStrategy             domain.FallbackStrategy
	FallbackRetryIntervalSeconds int
	OnErrorAction                domain.OnErrorAction
	OnErrorConfig                map[string]any
	CreatedBy                    string
}

type UpdateInstanceParams struct {
	Name                         *string
	Description                  *string
	Tags                         *[]string
	Status                       *domain.InstanceStatus
	Active                       *bool
	ProjectID                    *string
	DeviceID                     *string
	ScriptID                     *string
	IncludeDeviceData            *bool
	ScriptParameters             *map[string]any
	TriggerType                  *domain.TriggerType
	TriggerConfig                *map[string]any
	DependsOn                    *[]string
	ExecutionMode                *domain.ExecutionMode
	OrchestrationConfig          *map[string]any
	FallbackEnabled              *bool
	FallbackStrategy             *domain.FallbackStrategy
	FallbackRetryIntervalSeconds *int
	OnErrorAction                *domain.OnErrorAction
	OnErrorConfig                *map[string]any
}

func NewInstancesRepository(db *gorm.DB) *InstancesRepository {
	return &InstancesRepository{db: db}
}

func (r *InstancesRepository) Create(ctx context.Context, params UpsertInstanceParams) (domain.Instance, error) {
	model := instanceModelFromParams(params)
	if model.ID == "" {
		model.ID = newID()
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Instance{}, err
	}
	return instanceFromModel(model), nil
}

func (r *InstancesRepository) Upsert(ctx context.Context, params UpsertInstanceParams) (domain.Instance, error) {
	if params.ID == "" {
		return r.Create(ctx, params)
	}

	var model instanceModel
	err := r.db.WithContext(ctx).Where("id = ?", params.ID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.Create(ctx, params)
	}
	if err != nil {
		return domain.Instance{}, err
	}

	applyInstanceParams(&model, params)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.Instance{}, err
	}
	return instanceFromModel(model), nil
}

func (r *InstancesRepository) FindByID(ctx context.Context, id string) (domain.Instance, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *InstancesRepository) ListAll(ctx context.Context) ([]domain.Instance, error) {
	var models []instanceModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return instancesFromModels(models), nil
}

func (r *InstancesRepository) ListByProject(ctx context.Context, projectID string) ([]domain.Instance, error) {
	var models []instanceModel
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return instancesFromModels(models), nil
}

func (r *InstancesRepository) ListActive(ctx context.Context) ([]domain.Instance, error) {
	var models []instanceModel
	if err := r.db.WithContext(ctx).Where("active = ?", true).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	return instancesFromModels(models), nil
}

func (r *InstancesRepository) Update(ctx context.Context, id string, params UpdateInstanceParams) (domain.Instance, error) {
	var model instanceModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Instance{}, ErrNotFound
		}
		return domain.Instance{}, err
	}

	applyInstanceUpdate(&model, params)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.Instance{}, err
	}
	return instanceFromModel(model), nil
}

func (r *InstancesRepository) UpdateStatus(ctx context.Context, id string, status domain.InstanceStatus) (domain.Instance, error) {
	return r.Update(ctx, id, UpdateInstanceParams{Status: &status})
}

func (r *InstancesRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&instanceModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *InstancesRepository) findOne(ctx context.Context, query string, args ...any) (domain.Instance, error) {
	var model instanceModel
	if err := r.db.WithContext(ctx).Where(query, args...).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Instance{}, ErrNotFound
		}
		return domain.Instance{}, err
	}
	return instanceFromModel(model), nil
}

func instanceModelFromParams(params UpsertInstanceParams) instanceModel {
	model := instanceModel{ID: params.ID}
	applyInstanceParams(&model, params)
	return model
}

func applyInstanceParams(model *instanceModel, params UpsertInstanceParams) {
	if params.Status == "" {
		params.Status = domain.InstanceStatusIdle
	}
	active := true
	if params.Active != nil {
		active = *params.Active
	}
	if params.TriggerType == "" {
		params.TriggerType = domain.TriggerInterval
	}
	if params.TriggerConfig == nil {
		params.TriggerConfig = map[string]any{"interval_seconds": float64(60)}
	}
	if params.DependsOn == nil {
		params.DependsOn = []string{}
	}
	if params.ExecutionMode == "" {
		params.ExecutionMode = domain.ExecutionModeSequential
	}
	if params.OrchestrationConfig == nil {
		params.OrchestrationConfig = map[string]any{}
	}
	if params.ScriptParameters == nil {
		params.ScriptParameters = map[string]any{}
	}
	if params.FallbackStrategy == "" {
		params.FallbackStrategy = domain.FallbackBackgroundJob
	}
	if params.FallbackRetryIntervalSeconds == 0 {
		params.FallbackRetryIntervalSeconds = 300
	}
	if params.OnErrorAction == "" {
		params.OnErrorAction = domain.OnErrorLogOnly
	}
	if params.OnErrorConfig == nil {
		params.OnErrorConfig = map[string]any{}
	}

	model.Name = params.Name
	model.Description = params.Description
	model.Tags = stringSliceJSON(params.Tags)
	model.Status = string(params.Status)
	model.Active = active
	model.ProjectID = params.ProjectID
	model.DeviceID = params.DeviceID
	model.ScriptID = params.ScriptID
	model.IncludeDeviceData = params.IncludeDeviceData
	model.ScriptParameters = mapJSON(params.ScriptParameters)
	model.TriggerType = string(params.TriggerType)
	model.TriggerConfig = mapJSON(params.TriggerConfig)
	model.DependsOn = stringSliceJSON(params.DependsOn)
	model.ExecutionMode = string(params.ExecutionMode)
	model.OrchestrationConfig = mapJSON(params.OrchestrationConfig)
	model.FallbackEnabled = params.FallbackEnabled
	model.FallbackStrategy = string(params.FallbackStrategy)
	model.FallbackRetryIntervalSeconds = params.FallbackRetryIntervalSeconds
	model.OnErrorAction = string(params.OnErrorAction)
	model.OnErrorConfig = mapJSON(params.OnErrorConfig)
	model.CreatedBy = params.CreatedBy
}

func applyInstanceUpdate(model *instanceModel, params UpdateInstanceParams) {
	if params.Name != nil {
		model.Name = *params.Name
	}
	if params.Description != nil {
		model.Description = *params.Description
	}
	if params.Tags != nil {
		model.Tags = stringSliceJSON(*params.Tags)
	}
	if params.Status != nil {
		model.Status = string(*params.Status)
	}
	if params.Active != nil {
		model.Active = *params.Active
	}
	if params.ProjectID != nil {
		model.ProjectID = *params.ProjectID
	}
	if params.DeviceID != nil {
		model.DeviceID = *params.DeviceID
	}
	if params.ScriptID != nil {
		model.ScriptID = *params.ScriptID
	}
	if params.IncludeDeviceData != nil {
		model.IncludeDeviceData = *params.IncludeDeviceData
	}
	if params.ScriptParameters != nil {
		model.ScriptParameters = mapJSON(*params.ScriptParameters)
	}
	if params.TriggerType != nil {
		model.TriggerType = string(*params.TriggerType)
	}
	if params.TriggerConfig != nil {
		model.TriggerConfig = mapJSON(*params.TriggerConfig)
	}
	if params.DependsOn != nil {
		model.DependsOn = stringSliceJSON(*params.DependsOn)
	}
	if params.ExecutionMode != nil {
		model.ExecutionMode = string(*params.ExecutionMode)
	}
	if params.OrchestrationConfig != nil {
		model.OrchestrationConfig = mapJSON(*params.OrchestrationConfig)
	}
	if params.FallbackEnabled != nil {
		model.FallbackEnabled = *params.FallbackEnabled
	}
	if params.FallbackStrategy != nil {
		model.FallbackStrategy = string(*params.FallbackStrategy)
	}
	if params.FallbackRetryIntervalSeconds != nil {
		model.FallbackRetryIntervalSeconds = *params.FallbackRetryIntervalSeconds
	}
	if params.OnErrorAction != nil {
		model.OnErrorAction = string(*params.OnErrorAction)
	}
	if params.OnErrorConfig != nil {
		model.OnErrorConfig = mapJSON(*params.OnErrorConfig)
	}
}

func instancesFromModels(models []instanceModel) []domain.Instance {
	instances := make([]domain.Instance, 0, len(models))
	for _, model := range models {
		instances = append(instances, instanceFromModel(model))
	}
	return instances
}

func instanceFromModel(model instanceModel) domain.Instance {
	return domain.Instance{
		ID:                           model.ID,
		Name:                         model.Name,
		Description:                  model.Description,
		Tags:                         []string(model.Tags),
		Status:                       domain.InstanceStatus(model.Status),
		Active:                       model.Active,
		ProjectID:                    model.ProjectID,
		DeviceID:                     model.DeviceID,
		ScriptID:                     model.ScriptID,
		IncludeDeviceData:            model.IncludeDeviceData,
		ScriptParameters:             map[string]any(model.ScriptParameters),
		TriggerType:                  domain.TriggerType(model.TriggerType),
		TriggerConfig:                map[string]any(model.TriggerConfig),
		DependsOn:                    []string(model.DependsOn),
		ExecutionMode:                domain.ExecutionMode(model.ExecutionMode),
		OrchestrationConfig:          map[string]any(model.OrchestrationConfig),
		FallbackEnabled:              model.FallbackEnabled,
		FallbackStrategy:             domain.FallbackStrategy(model.FallbackStrategy),
		FallbackRetryIntervalSeconds: model.FallbackRetryIntervalSeconds,
		OnErrorAction:                domain.OnErrorAction(model.OnErrorAction),
		OnErrorConfig:                map[string]any(model.OnErrorConfig),
		CreatedBy:                    model.CreatedBy,
		CreatedAt:                    model.CreatedAt,
		UpdatedAt:                    model.UpdatedAt,
	}
}
