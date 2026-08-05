package sqlite

import (
	"context"
	"errors"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type InstanceExecutionsRepository struct {
	db *gorm.DB
}

type CreateInstanceExecutionParams struct {
	ID              string
	InstanceID      string
	Status          domain.ExecutionStatus
	TriggerType     domain.TriggerType
	StartedAt       *time.Time
	FinishedAt      *time.Time
	DurationMS      *int
	Logs            []domain.ExecutionLog
	Output          string
	ErrorMessage    string
	DestinationSent bool
	FallbackUsed    bool
}

type UpdateInstanceExecutionStatusParams struct {
	Status          domain.ExecutionStatus
	FinishedAt      *time.Time
	DurationMS      *int
	ErrorMessage    *string
	Output          *string
	DestinationSent *bool
	FallbackUsed    *bool
	Logs            *[]domain.ExecutionLog
}

func NewInstanceExecutionsRepository(db *gorm.DB) *InstanceExecutionsRepository {
	return &InstanceExecutionsRepository{db: db}
}

func (r *InstanceExecutionsRepository) Create(ctx context.Context, params CreateInstanceExecutionParams) (domain.InstanceExecution, error) {
	model := instanceExecutionModelFromCreate(params)
	if model.ID == "" {
		model.ID = newID()
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.InstanceExecution{}, err
	}
	return instanceExecutionFromModel(model), nil
}

func (r *InstanceExecutionsRepository) FindByID(ctx context.Context, id string) (domain.InstanceExecution, error) {
	var model instanceExecutionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.InstanceExecution{}, ErrNotFound
		}
		return domain.InstanceExecution{}, err
	}
	return instanceExecutionFromModel(model), nil
}

func (r *InstanceExecutionsRepository) ListByInstance(ctx context.Context, instanceID string, limit int) ([]domain.InstanceExecution, error) {
	if limit <= 0 {
		limit = 50
	}
	var models []instanceExecutionModel
	if err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("created_at DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	return instanceExecutionsFromModels(models), nil
}

func (r *InstanceExecutionsRepository) ListAll(ctx context.Context, limit int) ([]domain.InstanceExecution, error) {
	if limit <= 0 {
		limit = 100
	}
	var models []instanceExecutionModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&models).Error; err != nil {
		return nil, err
	}
	return instanceExecutionsFromModels(models), nil
}

func (r *InstanceExecutionsRepository) UpdateStatus(ctx context.Context, id string, params UpdateInstanceExecutionStatusParams) (domain.InstanceExecution, error) {
	var model instanceExecutionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.InstanceExecution{}, ErrNotFound
		}
		return domain.InstanceExecution{}, err
	}

	if params.Status != "" {
		model.Status = string(params.Status)
	}
	if params.FinishedAt != nil {
		model.FinishedAt = params.FinishedAt
	}
	if params.DurationMS != nil {
		model.DurationMS = params.DurationMS
	}
	if params.ErrorMessage != nil {
		model.ErrorMessage = *params.ErrorMessage
	}
	if params.Output != nil {
		model.Output = *params.Output
	}
	if params.DestinationSent != nil {
		model.DestinationSent = *params.DestinationSent
	}
	if params.FallbackUsed != nil {
		model.FallbackUsed = *params.FallbackUsed
	}
	if params.Logs != nil {
		model.Logs = executionLogsJSON(*params.Logs)
	}

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.InstanceExecution{}, err
	}
	return instanceExecutionFromModel(model), nil
}

func instanceExecutionModelFromCreate(params CreateInstanceExecutionParams) instanceExecutionModel {
	status := params.Status
	if status == "" {
		status = domain.ExecutionQueued
	}
	triggerType := params.TriggerType
	if triggerType == "" {
		triggerType = domain.TriggerInterval
	}
	if params.Logs == nil {
		params.Logs = []domain.ExecutionLog{}
	}
	return instanceExecutionModel{
		ID:              params.ID,
		InstanceID:      params.InstanceID,
		Status:          string(status),
		TriggerType:     string(triggerType),
		StartedAt:       params.StartedAt,
		FinishedAt:      params.FinishedAt,
		DurationMS:      params.DurationMS,
		Logs:            executionLogsJSON(params.Logs),
		Output:          params.Output,
		ErrorMessage:    params.ErrorMessage,
		DestinationSent: params.DestinationSent,
		FallbackUsed:    params.FallbackUsed,
	}
}

func instanceExecutionsFromModels(models []instanceExecutionModel) []domain.InstanceExecution {
	executions := make([]domain.InstanceExecution, 0, len(models))
	for _, model := range models {
		executions = append(executions, instanceExecutionFromModel(model))
	}
	return executions
}

func instanceExecutionFromModel(model instanceExecutionModel) domain.InstanceExecution {
	return domain.InstanceExecution{
		ID:              model.ID,
		InstanceID:      model.InstanceID,
		Status:          domain.ExecutionStatus(model.Status),
		TriggerType:     domain.TriggerType(model.TriggerType),
		StartedAt:       model.StartedAt,
		FinishedAt:      model.FinishedAt,
		DurationMS:      model.DurationMS,
		Logs:            []domain.ExecutionLog(model.Logs),
		Output:          model.Output,
		ErrorMessage:    model.ErrorMessage,
		DestinationSent: model.DestinationSent,
		FallbackUsed:    model.FallbackUsed,
		CreatedAt:       model.CreatedAt,
	}
}
