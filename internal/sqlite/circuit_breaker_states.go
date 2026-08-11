package sqlite

import (
	"context"
	"errors"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"gorm.io/gorm"
)

type CircuitBreakerStatesRepository struct {
	db *gorm.DB
}

func NewCircuitBreakerStatesRepository(db *gorm.DB) *CircuitBreakerStatesRepository {
	return &CircuitBreakerStatesRepository{db: db}
}

func (r *CircuitBreakerStatesRepository) GetByDestination(ctx context.Context, destinationID string) (domain.CircuitBreakerState, error) {
	var model circuitBreakerStateModel
	if err := r.db.WithContext(ctx).Where("destination_id = ?", destinationID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.CircuitBreakerState{}, ErrNotFound
		}
		return domain.CircuitBreakerState{}, err
	}
	return circuitBreakerStateFromModel(model), nil
}

func (r *CircuitBreakerStatesRepository) Upsert(ctx context.Context, state domain.CircuitBreakerState) (domain.CircuitBreakerState, error) {
	model := circuitBreakerStateModel{
		DestinationID:       state.DestinationID,
		ConsecutiveFailures: state.ConsecutiveFailures,
		OpenedUntil:         state.OpenedUntil,
		UpdatedAt:           time.Now().UTC(),
	}
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.CircuitBreakerState{}, err
	}
	return circuitBreakerStateFromModel(model), nil
}

func (r *CircuitBreakerStatesRepository) Delete(ctx context.Context, destinationID string) error {
	return r.db.WithContext(ctx).Where("destination_id = ?", destinationID).Delete(&circuitBreakerStateModel{}).Error
}

func (r *CircuitBreakerStatesRepository) ListByDestinationIDs(ctx context.Context, destinationIDs []string) ([]domain.CircuitBreakerState, error) {
	if len(destinationIDs) == 0 {
		return []domain.CircuitBreakerState{}, nil
	}
	var models []circuitBreakerStateModel
	if err := r.db.WithContext(ctx).Where("destination_id IN ?", destinationIDs).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]domain.CircuitBreakerState, 0, len(models))
	for _, model := range models {
		result = append(result, circuitBreakerStateFromModel(model))
	}
	return result, nil
}

func circuitBreakerStateFromModel(model circuitBreakerStateModel) domain.CircuitBreakerState {
	return domain.CircuitBreakerState{
		DestinationID:       model.DestinationID,
		ConsecutiveFailures: model.ConsecutiveFailures,
		OpenedUntil:         model.OpenedUntil,
		UpdatedAt:           model.UpdatedAt,
	}
}
