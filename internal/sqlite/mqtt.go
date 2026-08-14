package sqlite

import (
	"context"
	"errors"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type MqttCommandsRepository struct {
	db *gorm.DB
}

type MqttQueueRepository struct {
	db *gorm.DB
}

func NewMqttCommandsRepository(db *gorm.DB) *MqttCommandsRepository {
	return &MqttCommandsRepository{db: db}
}

func NewMqttQueueRepository(db *gorm.DB) *MqttQueueRepository {
	return &MqttQueueRepository{db: db}
}

func (r *MqttCommandsRepository) FindByCommandID(ctx context.Context, commandID string) (domain.MqttCommand, error) {
	var model mqttCommandModel
	if err := r.db.WithContext(ctx).Where("command_id = ?", commandID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.MqttCommand{}, ErrNotFound
		}
		return domain.MqttCommand{}, err
	}
	return mqttCommandFromModel(model), nil
}

func (r *MqttCommandsRepository) Save(ctx context.Context, commandID string, commandType string, payload map[string]any) (domain.MqttCommand, error) {
	model := mqttCommandModel{
		ID:        newID(),
		CommandID: commandID,
		Type:      commandType,
		Payload:   mapJSON(payload),
		Status:    string(domain.MqttCommandPending),
	}
	if model.Payload == nil {
		model.Payload = mapJSON{}
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.MqttCommand{}, err
	}
	return mqttCommandFromModel(model), nil
}

func (r *MqttCommandsRepository) UpdateStatus(ctx context.Context, commandID string, status domain.MqttCommandStatus, result map[string]any, errText string) (domain.MqttCommand, error) {
	now := time.Now().UTC()
	updates := map[string]any{"status": string(status)}
	if status == domain.MqttCommandRunning {
		updates["started_at"] = &now
	}
	if status == domain.MqttCommandDone || status == domain.MqttCommandError || status == domain.MqttCommandIgnored {
		updates["finished_at"] = &now
	}
	if result != nil {
		updates["result"] = mapJSON(result)
	}
	if errText != "" {
		updates["error"] = errText
	}
	if err := r.db.WithContext(ctx).Model(&mqttCommandModel{}).Where("command_id = ?", commandID).Updates(updates).Error; err != nil {
		return domain.MqttCommand{}, err
	}
	return r.FindByCommandID(ctx, commandID)
}

func (r *MqttQueueRepository) Enqueue(ctx context.Context, topic string, payload string) (domain.MqttQueueItem, error) {
	model := mqttQueueModel{ID: newID(), Topic: topic, Payload: payload}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.MqttQueueItem{}, err
	}
	return mqttQueueFromModel(model), nil
}

func (r *MqttQueueRepository) ListPending(ctx context.Context, maxAttempts int) ([]domain.MqttQueueItem, error) {
	var models []mqttQueueModel
	query := r.db.WithContext(ctx).Order("created_at ASC")
	if maxAttempts > 0 {
		query = query.Where("attempts < ?", maxAttempts)
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.MqttQueueItem, 0, len(models))
	for _, model := range models {
		result = append(result, mqttQueueFromModel(model))
	}
	return result, nil
}

func (r *MqttQueueRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&mqttQueueModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MqttQueueRepository) IncrementAttempt(ctx context.Context, id string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&mqttQueueModel{}).Where("id = ?", id).Updates(map[string]any{
		"attempts":        gorm.Expr("attempts + ?", 1),
		"last_attempt_at": &now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func mqttCommandFromModel(model mqttCommandModel) domain.MqttCommand {
	return domain.MqttCommand{
		ID:         model.ID,
		CommandID:  model.CommandID,
		Type:       model.Type,
		Payload:    map[string]any(model.Payload),
		Status:     domain.MqttCommandStatus(model.Status),
		Result:     map[string]any(model.Result),
		Error:      model.Error,
		CreatedAt:  model.CreatedAt,
		StartedAt:  model.StartedAt,
		FinishedAt: model.FinishedAt,
	}
}

func mqttQueueFromModel(model mqttQueueModel) domain.MqttQueueItem {
	return domain.MqttQueueItem{ID: model.ID, Topic: model.Topic, Payload: model.Payload, Attempts: model.Attempts, LastAttemptAt: model.LastAttemptAt, CreatedAt: model.CreatedAt}
}
