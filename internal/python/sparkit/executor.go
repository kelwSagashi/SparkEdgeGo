package sparkit

import (
	"context"
	"os/exec"
	"time"

	"github.com/kelwSagashi/sparkedge-go/internal/domain"
)

type Executor struct {
	DefaultTimeout time.Duration
}

func NewExecutor() *Executor {
	return &Executor{DefaultTimeout: 60 * time.Second}
}

func (e *Executor) Run(ctx context.Context, scriptPath string, input map[string]any) (domain.ScriptResult, error) {
	// TODO: Serialize input through Sparkit contract and capture stdout/stderr.
	cmd := exec.CommandContext(ctx, "python", scriptPath)
	_ = cmd
	return domain.ScriptResult{}, nil
}

func (e *Executor) Schema(ctx context.Context, scriptPath string) (map[string]any, error) {
	cmd := exec.CommandContext(ctx, "python", scriptPath, "--schema")
	_ = cmd
	return map[string]any{}, nil
}
