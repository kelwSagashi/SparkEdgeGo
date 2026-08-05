package executions

import (
	"context"
	"errors"
	"strings"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/sqlite"
)

var ErrInvalidExecution = errors.New("invalid execution")

type Repository interface {
	Create(ctx context.Context, params sqlite.CreateInstanceExecutionParams) (domain.InstanceExecution, error)
	FindByID(ctx context.Context, id string) (domain.InstanceExecution, error)
	ListByInstance(ctx context.Context, instanceID string, limit int) ([]domain.InstanceExecution, error)
	ListAll(ctx context.Context, limit int) ([]domain.InstanceExecution, error)
	UpdateStatus(ctx context.Context, id string, params sqlite.UpdateInstanceExecutionStatusParams) (domain.InstanceExecution, error)
}

type Service struct {
	executions Repository
}

func NewService(executions Repository) *Service {
	return &Service{executions: executions}
}

func (s *Service) ListAll(ctx context.Context, limit int) ([]domain.InstanceExecution, error) {
	return s.executions.ListAll(ctx, limit)
}

func (s *Service) FindByID(ctx context.Context, id string) (domain.InstanceExecution, error) {
	if strings.TrimSpace(id) == "" {
		return domain.InstanceExecution{}, ErrInvalidExecution
	}
	return s.executions.FindByID(ctx, id)
}

func (s *Service) ListByInstance(ctx context.Context, instanceID string, limit int) ([]domain.InstanceExecution, error) {
	if strings.TrimSpace(instanceID) == "" {
		return []domain.InstanceExecution{}, ErrInvalidExecution
	}
	return s.executions.ListByInstance(ctx, instanceID, limit)
}

func (s *Service) Create(ctx context.Context, params sqlite.CreateInstanceExecutionParams) (domain.InstanceExecution, error) {
	if strings.TrimSpace(params.InstanceID) == "" {
		return domain.InstanceExecution{}, ErrInvalidExecution
	}
	return s.executions.Create(ctx, params)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, params sqlite.UpdateInstanceExecutionStatusParams) (domain.InstanceExecution, error) {
	if strings.TrimSpace(id) == "" || params.Status == "" {
		return domain.InstanceExecution{}, ErrInvalidExecution
	}
	return s.executions.UpdateStatus(ctx, id, params)
}
