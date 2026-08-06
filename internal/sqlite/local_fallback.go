package sqlite

import (
	"context"
	"errors"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type LocalFallbackRepository struct {
	db *gorm.DB
}

type CreateLocalFallbackParams struct {
	InstanceID    string
	DestinationID string
	ExecutionID   string
	Payload       string
	Filepath      string
}

func NewLocalFallbackRepository(db *gorm.DB) *LocalFallbackRepository {
	return &LocalFallbackRepository{db: db}
}

func (r *LocalFallbackRepository) Create(ctx context.Context, params CreateLocalFallbackParams) (domain.LocalFallbackItem, error) {
	now := time.Now().UTC()
	model := localFallbackStorageModel{
		ID:            newID(),
		InstanceID:    params.InstanceID,
		DestinationID: params.DestinationID,
		ExecutionID:   params.ExecutionID,
		Status:        string(domain.FallbackPending),
		Payload:       params.Payload,
		Filepath:      params.Filepath,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.LocalFallbackItem{}, err
	}
	return localFallbackFromModel(model), nil
}

func (r *LocalFallbackRepository) FindByID(ctx context.Context, id string) (domain.LocalFallbackItem, error) {
	var model localFallbackStorageModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.LocalFallbackItem{}, ErrNotFound
		}
		return domain.LocalFallbackItem{}, err
	}
	return localFallbackFromModel(model), nil
}

func (r *LocalFallbackRepository) ListPending(ctx context.Context) ([]domain.LocalFallbackItem, error) {
	var models []localFallbackStorageModel
	if err := r.db.WithContext(ctx).Where("status = ?", string(domain.FallbackPending)).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	return localFallbackListFromModels(models), nil
}

func (r *LocalFallbackRepository) ListByInstance(ctx context.Context, instanceID string) ([]domain.LocalFallbackItem, error) {
	var models []localFallbackStorageModel
	if err := r.db.WithContext(ctx).Where("instance_id = ?", instanceID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	return localFallbackListFromModels(models), nil
}

func (r *LocalFallbackRepository) ListAll(ctx context.Context) ([]domain.LocalFallbackItem, error) {
	var models []localFallbackStorageModel
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	return localFallbackListFromModels(models), nil
}

func (r *LocalFallbackRepository) MarkAsSending(ctx context.Context, id string) (domain.LocalFallbackItem, error) {
	return r.updateStatus(ctx, id, domain.FallbackSending, "")
}

func (r *LocalFallbackRepository) MarkAsSent(ctx context.Context, id string) (domain.LocalFallbackItem, error) {
	return r.updateStatus(ctx, id, domain.FallbackSent, "")
}

func (r *LocalFallbackRepository) IncrementRetry(ctx context.Context, id string, lastError string) (domain.LocalFallbackItem, error) {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&localFallbackStorageModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":        string(domain.FallbackPending),
		"retry_count":   gorm.Expr("retry_count + ?", 1),
		"last_retry_at": &now,
		"last_error":    lastError,
		"updated_at":    now,
	})
	if result.Error != nil {
		return domain.LocalFallbackItem{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.LocalFallbackItem{}, ErrNotFound
	}
	return r.FindByID(ctx, id)
}

func (r *LocalFallbackRepository) MarkAsFailed(ctx context.Context, id string, lastError string) (domain.LocalFallbackItem, error) {
	return r.updateStatus(ctx, id, domain.FallbackFailed, lastError)
}

func (r *LocalFallbackRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&localFallbackStorageModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *LocalFallbackRepository) updateStatus(ctx context.Context, id string, status domain.FallbackItemStatus, lastError string) (domain.LocalFallbackItem, error) {
	updates := map[string]any{"status": string(status), "updated_at": time.Now().UTC()}
	if lastError != "" {
		updates["last_error"] = lastError
	}
	result := r.db.WithContext(ctx).Model(&localFallbackStorageModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return domain.LocalFallbackItem{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.LocalFallbackItem{}, ErrNotFound
	}
	return r.FindByID(ctx, id)
}

func localFallbackListFromModels(models []localFallbackStorageModel) []domain.LocalFallbackItem {
	items := make([]domain.LocalFallbackItem, 0, len(models))
	for _, model := range models {
		items = append(items, localFallbackFromModel(model))
	}
	return items
}

func localFallbackFromModel(model localFallbackStorageModel) domain.LocalFallbackItem {
	return domain.LocalFallbackItem{
		ID:            model.ID,
		InstanceID:    model.InstanceID,
		DestinationID: model.DestinationID,
		ExecutionID:   model.ExecutionID,
		Status:        domain.FallbackItemStatus(model.Status),
		Payload:       model.Payload,
		Filepath:      model.Filepath,
		RetryCount:    model.RetryCount,
		LastRetryAt:   model.LastRetryAt,
		NextRetryAt:   model.NextRetryAt,
		LastError:     model.LastError,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}
