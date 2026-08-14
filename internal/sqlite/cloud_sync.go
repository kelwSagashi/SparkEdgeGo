package sqlite

import (
	"context"
	"errors"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type CloudSyncQueueRepository struct {
	db *gorm.DB
}

func NewCloudSyncQueueRepository(db *gorm.DB) *CloudSyncQueueRepository {
	return &CloudSyncQueueRepository{db: db}
}

func (r *CloudSyncQueueRepository) Enqueue(ctx context.Context, eventType string, priority int, payload map[string]any) (domain.CloudSyncItem, error) {
	model := cloudSyncQueueModel{
		ID:        newID(),
		EventType: eventType,
		Priority:  priority,
		Payload:   mapJSON(payload),
		Status:    string(domain.CloudSyncPending),
	}
	if model.Payload == nil {
		model.Payload = mapJSON{}
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.CloudSyncItem{}, err
	}
	_ = r.prune(ctx)
	return cloudSyncItemFromModel(model), nil
}

func (r *CloudSyncQueueRepository) FindByID(ctx context.Context, id string) (domain.CloudSyncItem, error) {
	var model cloudSyncQueueModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.CloudSyncItem{}, ErrNotFound
		}
		return domain.CloudSyncItem{}, err
	}
	return cloudSyncItemFromModel(model), nil
}

func (r *CloudSyncQueueRepository) ListPending(ctx context.Context, limit int) ([]domain.CloudSyncItem, error) {
	var models []cloudSyncQueueModel
	now := time.Now().UTC()
	query := r.db.WithContext(ctx).
		Where("status IN ?", []string{string(domain.CloudSyncPending), string(domain.CloudSyncFailed)}).
		Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
		Order("priority DESC").
		Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.CloudSyncItem, 0, len(models))
	for _, model := range models {
		result = append(result, cloudSyncItemFromModel(model))
	}
	return result, nil
}

func (r *CloudSyncQueueRepository) ListRecent(ctx context.Context, limit int) ([]domain.CloudSyncItem, error) {
	var models []cloudSyncQueueModel
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.CloudSyncItem, 0, len(models))
	for _, model := range models {
		result = append(result, cloudSyncItemFromModel(model))
	}
	return result, nil
}

func (r *CloudSyncQueueRepository) MarkSent(ctx context.Context, id string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&cloudSyncQueueModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":          string(domain.CloudSyncSent),
		"last_attempt_at": &now,
		"next_retry_at":   nil,
		"updated_at":      now,
		"last_error":      "",
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CloudSyncQueueRepository) MarkFailed(ctx context.Context, id string, errText string, nextRetryAt *time.Time) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&cloudSyncQueueModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":          string(domain.CloudSyncFailed),
		"attempts":        gorm.Expr("attempts + ?", 1),
		"last_attempt_at": &now,
		"next_retry_at":   nextRetryAt,
		"updated_at":      now,
		"last_error":      errText,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CloudSyncQueueRepository) Stats(ctx context.Context) (map[string]any, error) {
	type statRow struct {
		Status string
		Count  int64
	}
	type eventTypeRow struct {
		EventType string
		Count     int64
	}
	type priorityRow struct {
		Priority int
		Count    int64
	}
	var rows []statRow
	if err := r.db.WithContext(ctx).Model(&cloudSyncQueueModel{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	stats := map[string]any{
		"pending":       int64(0),
		"sent":          int64(0),
		"failed":        int64(0),
		"ready":         int64(0),
		"by_status":     map[string]any{},
		"by_event_type": map[string]any{},
		"by_priority_band": map[string]any{
			"critical": int64(0),
			"high":     int64(0),
			"normal":   int64(0),
			"low":      int64(0),
		},
	}
	byStatus := stats["by_status"].(map[string]any)
	for _, row := range rows {
		stats[row.Status] = row.Count
		byStatus[row.Status] = row.Count
	}

	var eventTypeRows []eventTypeRow
	if err := r.db.WithContext(ctx).Model(&cloudSyncQueueModel{}).
		Select("event_type, count(*) as count").
		Where("status IN ?", []string{string(domain.CloudSyncPending), string(domain.CloudSyncFailed)}).
		Group("event_type").
		Scan(&eventTypeRows).Error; err != nil {
		return nil, err
	}
	byEventType := stats["by_event_type"].(map[string]any)
	for _, row := range eventTypeRows {
		byEventType[row.EventType] = row.Count
	}

	var priorityRows []priorityRow
	if err := r.db.WithContext(ctx).Model(&cloudSyncQueueModel{}).
		Select("priority, count(*) as count").
		Where("status IN ?", []string{string(domain.CloudSyncPending), string(domain.CloudSyncFailed)}).
		Group("priority").
		Scan(&priorityRows).Error; err != nil {
		return nil, err
	}
	byPriorityBand := stats["by_priority_band"].(map[string]any)
	for _, row := range priorityRows {
		band := priorityBand(row.Priority)
		byPriorityBand[band] = asInt64(byPriorityBand[band]) + row.Count
	}

	var oldest cloudSyncQueueModel
	if err := r.db.WithContext(ctx).
		Where("status IN ?", []string{string(domain.CloudSyncPending), string(domain.CloudSyncFailed)}).
		Order("created_at ASC").
		First(&oldest).Error; err == nil {
		stats["oldest_pending_created_at"] = oldest.CreatedAt
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var readyCount int64
	if err := r.db.WithContext(ctx).
		Model(&cloudSyncQueueModel{}).
		Where("status IN ?", []string{string(domain.CloudSyncPending), string(domain.CloudSyncFailed)}).
		Where("next_retry_at IS NULL OR next_retry_at <= ?", time.Now().UTC()).
		Count(&readyCount).Error; err != nil {
		return nil, err
	}
	stats["ready"] = readyCount
	return stats, nil
}

func priorityBand(priority int) string {
	switch {
	case priority >= 80:
		return "critical"
	case priority >= 60:
		return "high"
	case priority >= 40:
		return "normal"
	default:
		return "low"
	}
}

func asInt64(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func (r *CloudSyncQueueRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&cloudSyncQueueModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CloudSyncQueueRepository) prune(ctx context.Context) error {
	policy := currentRetentionPolicy()
	if err := deleteRowsOlderThan(
		ctx,
		r.db,
		&cloudSyncQueueModel{},
		"updated_at",
		time.Now().UTC().Add(-policy.CloudSyncSentRetention),
		"status = ?",
		[]any{string(domain.CloudSyncSent)},
	); err != nil {
		return err
	}

	if err := deleteRowsOlderThan(
		ctx,
		r.db,
		&cloudSyncQueueModel{},
		"updated_at",
		time.Now().UTC().Add(-policy.CloudSyncFailedRetention),
		"status = ?",
		[]any{string(domain.CloudSyncFailed)},
	); err != nil {
		return err
	}

	if err := deleteOldestRows(
		ctx,
		r.db,
		&cloudSyncQueueModel{},
		"status = ?",
		[]any{string(domain.CloudSyncSent)},
		policy.CloudSyncKeepSentItems,
	); err != nil {
		return err
	}

	return deleteOldestRows(
		ctx,
		r.db,
		&cloudSyncQueueModel{},
		"status = ?",
		[]any{string(domain.CloudSyncFailed)},
		policy.CloudSyncKeepFailedItems,
	)
}

func cloudSyncItemFromModel(model cloudSyncQueueModel) domain.CloudSyncItem {
	return domain.CloudSyncItem{
		ID:            model.ID,
		EventType:     model.EventType,
		Priority:      model.Priority,
		Payload:       map[string]any(model.Payload),
		Status:        domain.CloudSyncStatus(model.Status),
		Attempts:      model.Attempts,
		LastAttemptAt: model.LastAttemptAt,
		NextAttemptAt: model.NextRetryAt,
		LastError:     model.LastError,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}
