package runtime

import (
	"context"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
	"github.com/kelwSagashi/sparkedge-go/internal/providers"
)

type SparkitExecutor interface {
	Run(ctx context.Context, scriptPath string, input map[string]any) (domain.ScriptResult, error)
	Schema(ctx context.Context, scriptPath string) (map[string]any, error)
}

type Dependencies struct {
	Sparkit   SparkitExecutor
	Providers *providers.Registry
}

type Runner struct {
	deps Dependencies
}

func NewRunner(deps Dependencies) *Runner {
	return &Runner{deps: deps}
}

type TriggerRequest struct {
	Instance domain.Instance
	Script   domain.Script
	Trigger  domain.TriggerType
	Input    map[string]any
}

type TriggerResult struct {
	ExecutionID string
	Status      domain.ExecutionStatus
	Output      map[string]any
	Error       string
	StartedAt   time.Time
	FinishedAt  time.Time
}

func (r *Runner) Trigger(ctx context.Context, req TriggerRequest) (TriggerResult, error) {
	startedAt := time.Now().UTC()

	// TODO: Port InstanceRunnerService behavior:
	// resolve templates, call Sparkit, parse output, map payload, send destinations,
	// enqueue fallback, persist execution history, and reset instance status.
	return TriggerResult{
		ExecutionID: "",
		Status:      domain.ExecutionQueued,
		StartedAt:   startedAt,
		FinishedAt:  time.Now().UTC(),
	}, ctx.Err()
}
